package config

import (
	"log"
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	GoogleAppCreds string
	AppEnvironment string
	AllowedOrigins []string
}

func Load() *Config {
	cfg := &Config{
		Port:           os.Getenv("PORT"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		GoogleAppCreds: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		AppEnvironment: os.Getenv("APP_ENV"),
		AllowedOrigins: []string{
			os.Getenv("CORS_ALLOWED_ORIGIN"),
		},
	}

	if cfg.Port == "" {
		log.Fatalf("PORT environment variable is required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatalf("DATABASE_URL environment variable is required")
	}
	if cfg.GoogleAppCreds == "" {
		log.Fatalf("GOOGLE_APPLICATION_CREDENTIALS environment variable is required")
	}
	if cfg.AppEnvironment == "" {
		log.Fatalf("APP_ENV environment variable is required")
	}
	if len(cfg.AllowedOrigins) == 0 || cfg.AllowedOrigins[0] == "" {
		log.Fatalf("CORS_ALLOWED_ORIGIN environment variable is required")
	}

	return cfg
}
