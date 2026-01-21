package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken    string
	DatabaseURL string
	RedisURI    string
	XSubKey     string
	API         string
}

func Load() (*Config, error) {
	// Загружаем .env файл, если он существует
	_ = godotenv.Load()

	cfg := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURI:    os.Getenv("REDIS_URI"),
		XSubKey:     os.Getenv("X_SUB_KEY"),
		API:         os.Getenv("API_URL"),
	}

	// Валидация обязательных параметров
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURI == "" {
		return nil, fmt.Errorf("REDIS_URI is required")
	}

	return cfg, nil
}
