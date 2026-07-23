package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCAddr           string
	LLMProvider        string
	OpenAIAPIKey       string
	OpenAIModel        string
	OpenAIBaseURL      string
	GroqAPIKey         string
	GroqModel          string
	GroqBaseURL        string
	DatabaseURL        string
	EmbeddingProvider  string
	EmbeddingAPIKey    string
	EmbeddingModel     string
	EmbeddingBaseURL   string
	RAGTopK            int
}

func Load() Config {
	return Config{
		GRPCAddr:          getenv("GRPC_ADDR", ":50051"),
		LLMProvider:       getenv("LLM_PROVIDER", "mock"),
		OpenAIAPIKey:      os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:       getenv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:     getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		GroqAPIKey:        os.Getenv("GROQ_API_KEY"),
		GroqModel:         getenv("GROQ_MODEL", "llama-3.3-70b-versatile"),
		GroqBaseURL:       getenv("GROQ_BASE_URL", "https://api.groq.com/openai/v1"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		EmbeddingProvider: getenv("EMBEDDING_PROVIDER", "mock"),
		EmbeddingAPIKey:   firstNonEmpty(os.Getenv("EMBEDDING_API_KEY"), os.Getenv("OPENAI_API_KEY")),
		EmbeddingModel:    getenv("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingBaseURL:  getenv("EMBEDDING_BASE_URL", "https://api.openai.com/v1"),
		RAGTopK:           getenvInt("RAG_TOP_K", 4),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
