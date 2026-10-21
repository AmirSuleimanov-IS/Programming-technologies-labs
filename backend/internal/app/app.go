// Package app собирает приложение из модулей: репозитории, сервисы, маршруты и middleware.
// Его используют main и integration-тесты, поэтому тесты проверяют тот же HTTP-обработчик,
// что работает в production.
package app

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"sportreg/internal/application"
	"sportreg/internal/auth"
	"sportreg/internal/cache"
	"sportreg/internal/competition"
	"sportreg/internal/httpapi"
	"sportreg/internal/observability"
	"sportreg/internal/user"
)

// Deps — внешние зависимости приложения.
type Deps struct {
	Pool      *pgxpool.Pool
	Cache     *cache.Client // nil — без кэша
	JWTSecret string
	JWTTTL    time.Duration
	CacheTTL  competition.CacheTTL
}

// App — собранное приложение: HTTP-обработчик и сервис учётных записей (для создания администратора).
type App struct {
	Handler http.Handler
	Auth    *auth.Service
}

func New(d Deps) *App {
	users := user.NewRepository(d.Pool)
	tokens := auth.NewTokens(d.JWTSecret, d.JWTTTL)
	authService := auth.NewService(users, tokens)
	guard := auth.NewGuard(tokens).Allow

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := d.Pool.Ping(r.Context()); err != nil {
			httpapi.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "База данных недоступна", nil)
			return
		}
		httpapi.WriteData(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
	})
	mux.Handle("GET /metrics", promhttp.Handler())
	auth.NewHandler(authService, users).Register(mux, guard)
	user.NewHandler(users).Register(mux, guard)
	competitions := competition.NewService(competition.NewRepository(d.Pool), d.Cache, d.CacheTTL)
	competition.NewHandler(competitions).Register(mux, guard)
	application.NewHandler(application.NewService(application.NewRepository(d.Pool), competitions.Invalidate)).Register(mux, guard)

	handler := httpapi.Chain(mux,
		observability.Telemetry, // request_id, журнал доступа, метрики
		httpapi.Recover,
		httpapi.JSONErrors,
		observability.CaptureRoute,
	)
	return &App{Handler: handler, Auth: authService}
}
