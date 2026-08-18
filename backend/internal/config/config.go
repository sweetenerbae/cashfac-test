package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTP HTTPConfig
}

type HTTPConfig struct {
	Port int
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Port: envInt("HTTP_PORT", 8080),
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
