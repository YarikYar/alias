package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramBotToken string
	AppURL           string
	PostgresDSN      string
	RedisAddr        string
	ServerPort       string
}

func Load() *Config {
	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		AppURL:           getEnv("APP_URL", "http://localhost:3000"),
		PostgresDSN: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("POSTGRES_USER", "elias"),
			getEnv("POSTGRES_PASSWORD", "elias_secret_password"),
			getEnv("POSTGRES_HOST", "localhost"),
			getEnv("POSTGRES_PORT", "5432"),
			getEnv("POSTGRES_DB", "elias"),
		),
		RedisAddr:  fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379")),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
