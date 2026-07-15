package config

import "os"

type Config struct {
	GRPCAddr      string
	LLMProvider   string
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string
	GroqAPIKey    string
	GroqModel     string
	GroqBaseURL   string
}

func Load() Config {
	return Config{
		GRPCAddr:      getenv("GRPC_ADDR", ":50051"),
		LLMProvider:   getenv("LLM_PROVIDER", "mock"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:   getenv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL: getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		GroqAPIKey:    os.Getenv("GROQ_API_KEY"),
		GroqModel:     getenv("GROQ_MODEL", "llama-3.3-70b-versatile"),
		GroqBaseURL:   getenv("GROQ_BASE_URL", "https://api.groq.com/openai/v1"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
