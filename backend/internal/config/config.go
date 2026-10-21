// Package config читает настройки из переменных окружения и файла .env.<APP_ENV>.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

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

	JWTSecret     string
	JWTTTL        time.Duration
	AdminEmail    string // первый администратор создаётся при старте, если задан
	AdminPassword string
}

const minJWTSecretLen = 32

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

	// С пустым или коротким ключом токены можно подобрать, поэтому сервер не запускается
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < minJWTSecretLen {
		return Config{}, fmt.Errorf("JWT_SECRET должен быть не короче %d символов", minJWTSecretLen)
	}
	ttl, err := time.ParseDuration(getenv("JWT_TTL", "12h"))
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL: ожидается длительность вида 12h")
	}
	cfg.JWTTTL = ttl
	cfg.AdminEmail = os.Getenv("ADMIN_EMAIL")
	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
