package observability

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// statusRecorder запоминает код ответа для логов и метрик.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// serviceRoutes — служебные маршруты, успешные запросы к которым пишутся только на уровне DEBUG.
var serviceRoutes = map[string]bool{
	"GET /metrics":    true,
	"GET /api/health": true,
}

// Telemetry — внешняя middleware: присваивает запросу request_id, а после ответа
// пишет строку журнала доступа и обновляет метрики.
func Telemetry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		info := &RequestInfo{ID: r.Header.Get("X-Request-ID"), Method: r.Method, Path: r.URL.Path}
		if info.ID == "" || len(info.ID) > 64 {
			info.ID = newRequestID()
		}
		w.Header().Set("X-Request-ID", info.ID)
		ctx := withInfo(r.Context(), info)
		r = r.WithContext(ctx)
		slog.DebugContext(ctx, "request started", "user_agent", r.UserAgent())

		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}

		route := info.Route
		if route == "" {
			route = "unmatched"
		}
		elapsed := time.Since(start)
		status := strconv.Itoa(rec.status)
		httpRequests.WithLabelValues(r.Method, route, status).Inc()
		httpDuration.WithLabelValues(r.Method, route).Observe(elapsed.Seconds())
		if rec.status >= 400 {
			httpErrors.WithLabelValues(status[:1]+"xx", route).Inc()
		}

		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		case serviceRoutes[route]:
			level = slog.LevelDebug // Prometheus опрашивает /metrics каждые 5 с — в production это шум
		}
		slog.Log(ctx, level, "http request",
			"method", r.Method,
			"route", route,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", float64(elapsed.Microseconds())/1000,
		)
	})
}

// CaptureRoute ставится сразу перед маршрутизатором: после обработки запроса
// ServeMux записывает в r.Pattern шаблон найденного маршрута.
func CaptureRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { // defer — чтобы маршрут попал в метрики и при панике в обработчике
			if info := Info(r.Context()); info != nil {
				info.Route = r.Pattern
			}
		}()
		next.ServeHTTP(w, r)
	})
}
