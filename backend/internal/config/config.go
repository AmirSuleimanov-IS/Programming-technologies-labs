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

	RedisAddr string        // пусто — кэш выключен
	CacheTTL  time.Duration // одно соревнование
	ListTTL   time.Duration // результаты поиска по каталогу
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
	var err error
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < minJWTSecretLen {
		return Config{}, fmt.Errorf("JWT_SECRET должен быть не короче %d символов", minJWTSecretLen)
	}
	if cfg.JWTTTL, err = duration("JWT_TTL", "12h"); err != nil {
		return Config{}, err
	}
	cfg.AdminEmail = os.Getenv("ADMIN_EMAIL")
	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")

	cfg.RedisAddr = os.Getenv("REDIS_ADDR")
	if cfg.CacheTTL, err = duration("CACHE_ITEM_TTL", "5m"); err != nil {
		return Config{}, err
	}
	if cfg.ListTTL, err = duration("CACHE_LIST_TTL", "60s"); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func duration(key, fallback string) (time.Duration, error) {
	d, err := time.ParseDuration(getenv(key, fallback))
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s: ожидается положительная длительность вида 30s, 5m, 12h", key)
	}
	return d, nil
}
