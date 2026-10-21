package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

// Chain применяет middleware так, что первая в списке выполняется первой.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Recover превращает панику в ответ 500 и запись уровня ERROR.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.ErrorContext(r.Context(), "panic", "method", r.Method, "path", r.URL.Path, "panic", v)
				WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// JSONErrors заменяет текстовые ответы 404 и 405 стандартного маршрутизатора
// на ответы в общем формате API.
func JSONErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&jsonErrorWriter{ResponseWriter: w}, r)
	})
}

type jsonErrorWriter struct {
	http.ResponseWriter
	replaced bool
}

func (j *jsonErrorWriter) WriteHeader(code int) {
	plain := strings.HasPrefix(j.Header().Get("Content-Type"), "text/plain")
	switch {
	case plain && code == http.StatusNotFound:
		j.replaced = true
		WriteError(j.ResponseWriter, code, "NOT_FOUND", "Маршрут не найден", nil)
	case plain && code == http.StatusMethodNotAllowed:
		j.replaced = true
		WriteError(j.ResponseWriter, code, "METHOD_NOT_ALLOWED", "Метод не поддерживается для этого маршрута", nil)
	default:
		j.ResponseWriter.WriteHeader(code)
	}
}

func (j *jsonErrorWriter) Write(b []byte) (int, error) {
	if j.replaced {
		return len(b), nil
	}
	return j.ResponseWriter.Write(b)
}
