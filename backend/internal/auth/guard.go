package auth

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"sportreg/internal/httpapi"
	"sportreg/internal/observability"
	"sportreg/internal/user"
)

// Guard защищает маршруты: проверяет JWT из заголовка Authorization и роль пользователя.
type Guard struct {
	tokens *Tokens
}

func NewGuard(tokens *Tokens) *Guard {
	return &Guard{tokens: tokens}
}

// Allow пропускает запрос, если токен действителен и роль входит в roles.
// Без токена или с недействительным токеном — 401, с чужой ролью — 403.
func (g *Guard) Allow(roles ...user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			raw, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !found || raw == "" {
				observability.RecordDenied("no_token")
				slog.WarnContext(ctx, "access without token", "result", "unauthorized")
				unauthorized(w, "Требуется вход в систему")
				return
			}
			current, err := g.tokens.Parse(raw)
			if err != nil {
				observability.RecordDenied("invalid_token")
				slog.WarnContext(ctx, "invalid token", "error", err.Error(), "result", "unauthorized")
				unauthorized(w, "Токен недействителен или истёк, войдите снова")
				return
			}
			if info := observability.Info(ctx); info != nil {
				info.UserID, info.Role = current.ID, string(current.Role)
			}
			if !slices.Contains(roles, current.Role) {
				observability.RecordDenied("forbidden")
				slog.WarnContext(ctx, "access denied", "role", current.Role, "required", roles, "result", "forbidden")
				httpapi.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав", nil)
				return
			}
			slog.DebugContext(ctx, "access granted", "role", current.Role)
			next.ServeHTTP(w, r.WithContext(user.WithCurrent(ctx, current)))
		})
	}
}

func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="sportreg"`)
	httpapi.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}
