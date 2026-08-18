package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTP   HTTPConfig
	Source SourceConfig
}

type HTTPConfig struct {
	Port int
}

type SourceConfig struct {
	GuardianAPIKey string
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Port: envInt("HTTP_PORT", 8080),
		},
		Source: SourceConfig{
			GuardianAPIKey: os.Getenv("GUARDIAN_API_KEY"),
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
