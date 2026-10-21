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

	"sportreg/internal/competition"
	"sportreg/internal/config"
	"sportreg/internal/database"
	"sportreg/internal/httpapi"
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})))

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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httpapi.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "База данных недоступна", nil)
			return
		}
		httpapi.WriteData(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
	})
	competition.NewHandler(competition.NewService(competition.NewRepository(pool))).Register(mux)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.Chain(mux, httpapi.AccessLog, httpapi.Recover, httpapi.JSONErrors),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server started", "addr", cfg.HTTPAddr, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
