package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sportreg/internal/user"
)

// Проверка прав на бэкенде: таблица «роль маршрута × заголовок запроса → код ответа»
func TestGuard_Allow(t *testing.T) {
	tokens := NewTokens(testSecret, time.Hour)
	userToken, _, _ := tokens.Issue(user.User{ID: 7, Role: user.RoleUser})
	adminToken, _, _ := tokens.Issue(user.User{ID: 1, Role: user.RoleAdmin})
	guard := NewGuard(tokens)

	var seen user.Current
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = user.CurrentFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	adminOnly := guard.Allow(user.RoleAdmin)(ok)
	anyone := guard.Allow(user.RoleUser, user.RoleAdmin)(ok)

	tests := []struct {
		name    string
		handler http.Handler
		header  string
		want    int
	}{
		{"без заголовка", anyone, "", http.StatusUnauthorized},
		{"не Bearer", anyone, "Token " + userToken, http.StatusUnauthorized},
		{"пустой Bearer", anyone, "Bearer ", http.StatusUnauthorized},
		{"испорченный токен", anyone, "Bearer " + userToken + "x", http.StatusUnauthorized},
		{"USER на маршруте для всех", anyone, "Bearer " + userToken, http.StatusOK},
		{"USER на маршруте ADMIN", adminOnly, "Bearer " + userToken, http.StatusForbidden},
		{"ADMIN на маршруте ADMIN", adminOnly, "Bearer " + adminToken, http.StatusOK},
		{"ADMIN на маршруте для всех", anyone, "Bearer " + adminToken, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			tt.handler.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("код %d, ожидался %d (тело %s)", rec.Code, tt.want, rec.Body)
			}
			if rec.Code == http.StatusUnauthorized && rec.Header().Get("WWW-Authenticate") == "" {
				t.Error("в ответе 401 нет заголовка WWW-Authenticate")
			}
		})
	}

	// обработчик получает пользователя из токена
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	anyone.ServeHTTP(httptest.NewRecorder(), req)
	if seen != (user.Current{ID: 7, Role: user.RoleUser}) {
		t.Errorf("пользователь в контексте: %+v", seen)
	}
}
