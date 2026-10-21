// Package config читает настройки из переменных окружения и файла .env.<APP_ENV>.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type Config struct {
	Env         string
	HTTPAddr    string
	DatabaseURL string
	LogLevel    slog.Level
	SeedDemo    bool
}

// Load читает .env.development или .env.production (по APP_ENV, по умолчанию development).
// Переменные, уже заданные в окружении процесса (например, docker-compose), важнее файла.
func Load() (Config, error) {
	env := getenv("APP_ENV", EnvDevelopment)
	if env != EnvDevelopment && env != EnvProduction {
		return Config{}, fmt.Errorf("APP_ENV=%q: допустимы development и production", env)
	}
	if err := godotenv.Load(".env." + env); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read .env.%s: %w", env, err)
	}

	cfg := Config{
		Env:         env,
		HTTPAddr:    getenv("HTTP_ADDR", ":8090"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SeedDemo:    os.Getenv("SEED_DEMO") == "true",
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL не задан")
	}
	if err := cfg.LogLevel.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
