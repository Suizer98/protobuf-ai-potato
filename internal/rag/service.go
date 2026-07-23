package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	store    *Store
	embedder Embedder
	topK     int
}

func NewService(store *Store, embedder Embedder, topK int) *Service {
	if topK <= 0 {
		topK = 4
	}
	return &Service{
		store:    store,
		embedder: embedder,
		topK:     topK,
	}
}

func (s *Service) Close() {
	if s != nil {
		s.store.Close()
	}
}

func (s *Service) Ingest(ctx context.Context, title, source, content string) (uuid.UUID, int, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return uuid.Nil, 0, fmt.Errorf("title is required")
	}
	if content == "" {
		return uuid.Nil, 0, fmt.Errorf("content is required")
	}

	chunks := ChunkText(content)
	if len(chunks) == 0 {
		return uuid.Nil, 0, fmt.Errorf("no chunks produced")
	}

	embeddings, err := s.embedder.Embed(ctx, chunks)
	if err != nil {
		return uuid.Nil, 0, err
	}

	docID, err := s.store.InsertDocument(ctx, title, strings.TrimSpace(source), chunks, embeddings)
	if err != nil {
		return uuid.Nil, 0, err
	}
	return docID, len(chunks), nil
}

func (s *Service) RetrieveContext(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", nil
	}

	vectors, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return "", err
	}
	if len(vectors) == 0 {
		return "", nil
	}

	hits, err := s.store.Search(ctx, vectors[0], s.topK)
	if err != nil {
		return "", err
	}
	if len(hits) == 0 {
		return "", nil
	}

	var b strings.Builder
	b.WriteString("Use the following retrieved context when it helps answer the user.\n\n")
	for i, hit := range hits {
		fmt.Fprintf(&b, "[%d]\n%s\n\n", i+1, hit.Content)
	}
	return strings.TrimSpace(b.String()), nil
}
