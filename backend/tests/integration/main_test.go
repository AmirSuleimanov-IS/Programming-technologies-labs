//go:build integration

// Package integration — integration-тесты: настоящий HTTP-обработчик приложения (internal/app),
// PostgreSQL и Redis в отдельных контейнерах (testcontainers). Основная БД приложения не используется:
// на каждый запуск создаётся новый контейнер, который удаляется после тестов.
//
// Запуск: go test -tags=integration ./tests/integration/...  (нужен Docker)
package integration

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"sportreg/internal/app"
	"sportreg/internal/auth"
	"sportreg/internal/cache"
	"sportreg/internal/competition"
	"sportreg/internal/database"
)

const (
	adminEmail    = "admin@test.local"
	adminPassword = "Admin12345"
)

var (
	server  *httptest.Server
	pool    *pgxpool.Pool
	rdb     *redis.Client
	authSvc *auth.Service
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil))) // логи приложения не мешают выводу тестов

	pg, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("sportreg_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	defer terminate(pg)
	if err != nil {
		log.Printf("start postgres: %v", err)
		return 1
	}
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Print(err)
		return 1
	}
	if pool, err = database.Connect(ctx, dsn); err != nil {
		log.Print(err)
		return 1
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		log.Print(err)
		return 1
	}

	rc, err := tcredis.Run(ctx, "redis:7.4-alpine")
	defer terminate(rc)
	if err != nil {
		log.Printf("start redis: %v", err)
		return 1
	}
	redisURL, err := rc.ConnectionString(ctx)
	if err != nil {
		log.Print(err)
		return 1
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Print(err)
		return 1
	}
	rdb = redis.NewClient(opts)
	defer func() { _ = rdb.Close() }()

	a := app.New(app.Deps{
		Pool:      pool,
		Cache:     cache.New(rdb),
		JWTSecret: "integration-test-secret-0123456789abcdef",
		JWTTTL:    time.Hour,
		CacheTTL:  competition.CacheTTL{Item: time.Minute, List: time.Minute},
	})
	authSvc = a.Auth
	server = httptest.NewServer(a.Handler)
	defer server.Close()
	return m.Run()
}

func terminate(c testcontainers.Container) {
	if err := testcontainers.TerminateContainer(c); err != nil {
		log.Printf("terminate container: %v", err)
	}
}
