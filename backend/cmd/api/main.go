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

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"sportreg/internal/application"
	"sportreg/internal/auth"
	"sportreg/internal/competition"
	"sportreg/internal/config"
	"sportreg/internal/database"
	"sportreg/internal/httpapi"
	"sportreg/internal/observability"
	"sportreg/internal/user"
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

	users := user.NewRepository(pool)
	tokens := auth.NewTokens(cfg.JWTSecret, cfg.JWTTTL)
	authService := auth.NewService(users, tokens)
	if cfg.AdminEmail != "" {
		if err := authService.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
			return err
		}
	}
	guard := auth.NewGuard(tokens).Allow

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httpapi.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "База данных недоступна", nil)
			return
		}
		httpapi.WriteData(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
	})
	mux.Handle("GET /metrics", promhttp.Handler())
	auth.NewHandler(authService, users).Register(mux, guard)
	user.NewHandler(users).Register(mux, guard)
	competition.NewHandler(competition.NewService(competition.NewRepository(pool))).Register(mux, guard)
	application.NewHandler(application.NewService(application.NewRepository(pool))).Register(mux, guard)

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpapi.Chain(mux,
			observability.Telemetry, // request_id, журнал доступа, метрики
			httpapi.Recover,
			httpapi.JSONErrors,
			observability.CaptureRoute,
		),
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
