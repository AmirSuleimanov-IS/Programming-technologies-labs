package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Метрики регистрируются в реестре Prometheus по умолчанию и отдаются на /metrics.
// Метка route — шаблон маршрута (GET /api/competitions/{id}), а не путь,
// иначе каждый id давал бы новый временной ряд.
var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Количество HTTP-запросов",
	}, []string{"method", "route", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Время ответа на HTTP-запрос",
		Buckets: []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
	}, []string{"method", "route"})

	httpErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_errors_total",
		Help: "Количество ответов с ошибкой по классу: 4xx или 5xx",
	}, []string{"class", "route"})

	authAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_attempts_total",
		Help: "Попытки входа: success или failure",
	}, []string{"result"})

	authDenied = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_denied_total",
		Help: "Отказы в доступе к защищённым маршрутам: no_token, invalid_token, forbidden",
	}, []string{"reason"})
)

// RecordLogin учитывает попытку входа.
func RecordLogin(success bool) {
	result := "failure"
	if success {
		result = "success"
	}
	authAttempts.WithLabelValues(result).Inc()
}

// RecordDenied учитывает отказ в доступе.
func RecordDenied(reason string) {
	authDenied.WithLabelValues(reason).Inc()
}
