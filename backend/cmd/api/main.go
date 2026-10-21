// Команда api запускает REST API SportReg.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"sportreg/internal/app"
	"sportreg/internal/cache"
	"sportreg/internal/competition"
	"sportreg/internal/config"
	"sportreg/internal/database"
	"sportreg/internal/observability"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger, raised := observability.NewLogger(os.Stdout, cfg.Env == config.EnvProduction, cfg.LogLevel)
	slog.SetDefault(logger)
	logLevel := cfg.LogLevel
	if raised {
		slog.Warn("LOG_LEVEL below info is not allowed in production, using info", "requested", logLevel.String())
		logLevel = slog.LevelInfo
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}
	if cfg.SeedDemo {
		if err := database.SeedDemo(ctx, pool); err != nil {
			return err
		}
	}

	api := app.New(app.Deps{
		Pool:      pool,
		Cache:     newCache(ctx, cfg.RedisAddr),
		JWTSecret: cfg.JWTSecret,
		JWTTTL:    cfg.JWTTTL,
		CacheTTL:  competition.CacheTTL{Item: cfg.CacheTTL, List: cfg.ListTTL},
	})
	if cfg.AdminEmail != "" {
		if err := api.Auth.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
			return err
		}
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server started", "addr", cfg.HTTPAddr, "env", cfg.Env, "log_level", logLevel.String())
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// newCache подключает Redis. Если адрес не задан, кэш выключен. Если Redis недоступен
// при старте, сервер всё равно запускается: кэш работает в режиме fail-open.
func newCache(ctx context.Context, addr string) *cache.Client {
	if addr == "" {
		slog.Warn("REDIS_ADDR is empty, cache disabled")
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:          addr,
		DialTimeout:   300 * time.Millisecond, // при недоступном Redis запрос не должен зависать
		ReadTimeout:   200 * time.Millisecond,
		WriteTimeout:  200 * time.Millisecond,
		MaxRetries:    -1, // без повторов: при ошибке запрос сразу идёт в БД
		DialerRetries: 1,  // по умолчанию 5 попыток соединения — секунды ожидания
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("redis is unavailable, requests will go to the database", "addr", addr, "error", err)
	} else {
		slog.Info("redis connected", "addr", addr)
	}
	return cache.New(rdb)
}
