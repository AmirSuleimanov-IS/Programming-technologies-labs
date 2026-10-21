// Package user — учётные записи, роли и текущий пользователь запроса.
package user

import (
	"context"
	"net/http"
	"time"
)

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

// User — учётная запись. Хеш пароля наружу не отдаётся.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	PasswordHash string    `json:"-"`
}

// Current — пользователь, от имени которого выполняется запрос (из JWT).
type Current struct {
	ID   int64
	Role Role
}

func (c Current) IsAdmin() bool {
	return c.Role == RoleAdmin
}

// Guard пропускает к обработчику только пользователей с перечисленными ролями.
// Реализация — auth.Guard.Allow; модули получают её параметром и не зависят от auth.
type Guard func(roles ...Role) func(http.Handler) http.Handler

type ctxKey struct{}

func WithCurrent(ctx context.Context, c Current) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// CurrentFrom возвращает пользователя запроса. Внутри маршрутов под Guard он есть всегда.
func CurrentFrom(ctx context.Context) (Current, bool) {
	c, ok := ctx.Value(ctxKey{}).(Current)
	return c, ok
}
