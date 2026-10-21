package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"sportreg/internal/user"
)

// Tokens выпускает и проверяет JWT (HS256). В токене — id пользователя (sub) и роль.
type Tokens struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokens(secret string, ttl time.Duration) *Tokens {
	return &Tokens{secret: []byte(secret), ttl: ttl, now: time.Now}
}

type claims struct {
	Role user.Role `json:"role"`
	jwt.RegisteredClaims
}

func (t *Tokens) Issue(u user.User) (string, time.Time, error) {
	expires := t.now().Add(t.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(t.now()),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	})
	signed, err := token.SignedString(t.secret)
	return signed, expires, err
}

// Parse проверяет подпись, алгоритм и срок действия токена.
func (t *Tokens) Parse(raw string) (user.Current, error) {
	var c claims
	_, err := jwt.ParseWithClaims(raw, &c, func(*jwt.Token) (any, error) { return t.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(t.now),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return user.Current{}, err
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil || id <= 0 {
		return user.Current{}, fmt.Errorf("bad subject %q", c.Subject)
	}
	if c.Role != user.RoleAdmin && c.Role != user.RoleUser {
		return user.Current{}, errors.New("unknown role")
	}
	return user.Current{ID: id, Role: c.Role}, nil
}
