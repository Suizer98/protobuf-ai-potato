package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const EmbeddingDims = 1536

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type OpenAIEmbedder struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIEmbedder(apiKey, baseURL, model string) *OpenAIEmbedder {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &OpenAIEmbedder{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("embedding api key is required")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embeddingRequest{Model: e.model, Input: texts})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed embeddingResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: got %d want %d", len(parsed.Data), len(texts))
	}

	out := make([][]float32, len(texts))
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(out) {
			return nil, fmt.Errorf("invalid embedding index %d", item.Index)
		}
		vec := item.Embedding
		if len(vec) != EmbeddingDims {
			return nil, fmt.Errorf("unexpected embedding dims %d (want %d)", len(vec), EmbeddingDims)
		}
		out[item.Index] = vec
	}
	return out, nil
}

type MockEmbedder struct{}

func NewMockEmbedder() *MockEmbedder {
	return &MockEmbedder{}
}

func (e *MockEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, text := range texts {
		out[i] = mockVector(text)
	}
	return out, nil
}

func mockVector(text string) []float32 {
	vec := make([]float32, EmbeddingDims)
	tokens := strings.Fields(strings.ToLower(text))
	if len(tokens) == 0 {
		tokens = []string{text}
	}
	for _, token := range tokens {
		h := fnv.New64a()
		_, _ = h.Write([]byte(token))
		sum := h.Sum64()
		i := int(sum % uint64(EmbeddingDims))
		j := int((sum >> 17) % uint64(EmbeddingDims))
		vec[i] += 1
		vec[j] += 0.5
	}
	normalize(vec)
	return vec
}

func normalize(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	if sum == 0 {
		return
	}
	norm := float32(math.Sqrt(sum))
	for i := range vec {
		vec[i] /= norm
	}
}
