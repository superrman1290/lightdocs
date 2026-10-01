package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                 string
	HTTPAddr            string
	DatabaseURL         string
	FrontendOrigins     []string
	AccessTokenTTLMin   int
	RememberSessionDays int
	MaxUploadBytes      int64
}

func Load() Config {
	return Config{
		Env:                 env("APP_ENV", "development"),
		HTTPAddr:            env("HTTP_ADDR", ":8080"),
		DatabaseURL:         env("DATABASE_URL", ""),
		FrontendOrigins:     csvEnv("FRONTEND_ORIGINS", []string{"http://localhost:5173"}),
		AccessTokenTTLMin:   intEnv("ACCESS_TOKEN_TTL_MINUTES", 30),
		RememberSessionDays: intEnv("REMEMBER_SESSION_DAYS", 7),
		MaxUploadBytes:      int64Env("MAX_UPLOAD_BYTES", 10*1024*1024),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func int64Env(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(env(key, strconv.FormatInt(fallback, 10)), 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func csvEnv(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
