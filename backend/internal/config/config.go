package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTP   HTTPConfig
	Store  StoreConfig
	Source SourceConfig
	AI     AIConfig
}

type HTTPConfig struct {
	Port int
}

type SourceConfig struct {
	GuardianAPIKey string
}

type StoreConfig struct {
	SQLitePath string
}

type AIConfig struct {
	ZAIAPIKey  string
	ZAIBaseURL string
	ZAIModel   string
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Port: envInt("HTTP_PORT", 8080),
		},
		Store: StoreConfig{
			SQLitePath: envString("SQLITE_PATH", "data/news.db"),
		},
		Source: SourceConfig{
			GuardianAPIKey: os.Getenv("GUARDIAN_API_KEY"),
		},
		AI: AIConfig{
			ZAIAPIKey:  os.Getenv("ZAI_API_KEY"),
			ZAIBaseURL: envString("ZAI_BASE_URL", "https://api.z.ai/api/paas/v4/chat/completions"),
			ZAIModel:   envString("ZAI_MODEL", "glm-5.2"),
		},
	}
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func envString(key, fallback string) string {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	return raw
}
