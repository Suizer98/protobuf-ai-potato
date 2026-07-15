package config

import "os"

type Config struct {
	GRPCAddr      string
	LLMProvider   string
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string
}

func Load() Config {
	return Config{
		GRPCAddr:      getenv("GRPC_ADDR", ":50051"),
		LLMProvider:   getenv("LLM_PROVIDER", "mock"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:   getenv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL: getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
