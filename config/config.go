package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppHost     string
	AppPort     int
	LogLevel    string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // игнорируем ошибку, берём env либо .env

	port := 8000
	if p := os.Getenv("APP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	return &Config{
		DatabaseURL: getenv("DATABASE_URL", os.Getenv("DATABASE_URL")),
		AppHost:     getenv("APP_HOST", "0.0.0.0"),
		AppPort:     port,
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
