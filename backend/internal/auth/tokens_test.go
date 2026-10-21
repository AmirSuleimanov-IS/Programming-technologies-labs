package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"sportreg/internal/user"
)

const testSecret = "test-secret-0123456789-0123456789-abc"

func newTestTokens(now time.Time) *Tokens {
	t := NewTokens(testSecret, time.Hour)
	t.now = func() time.Time { return now }
	return t
}

func TestTokens_IssueAndParse(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	tokens := newTestTokens(now)
	raw, expires, err := tokens.Issue(user.User{ID: 42, Role: user.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	if !expires.Equal(now.Add(time.Hour)) {
		t.Errorf("expires = %v", expires)
	}
	got, err := tokens.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != (user.Current{ID: 42, Role: user.RoleAdmin}) {
		t.Errorf("Parse() = %+v", got)
	}
}

func TestTokens_Expired(t *testing.T) {
	issued := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	raw, _, _ := newTestTokens(issued).Issue(user.User{ID: 1, Role: user.RoleUser})

	// за секунду до конца срока токен действует, через секунду после — нет
	if _, err := newTestTokens(issued.Add(time.Hour - time.Second)).Parse(raw); err != nil {
		t.Errorf("токен должен действовать до истечения срока: %v", err)
	}
	if _, err := newTestTokens(issued.Add(time.Hour + time.Second)).Parse(raw); err == nil {
		t.Error("истёкший токен принят")
	}
}

// Подделки токена: чужой ключ, другой алгоритм, alg: none, неизвестная роль, нечисловой sub
func TestTokens_RejectsForged(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	tokens := newTestTokens(now)
	claimsFor := func(sub string, role user.Role) claims {
		return claims{Role: role, RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		}}
	}
	sign := func(method jwt.SigningMethod, key any, c claims) string {
		s, err := jwt.NewWithClaims(method, c).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	tests := []struct {
		name string
		raw  string
	}{
		{"подпись другим ключом", sign(jwt.SigningMethodHS256, []byte("other-secret-0123456789-0123456789"), claimsFor("1", user.RoleAdmin))},
		{"алгоритм HS512 вместо HS256", sign(jwt.SigningMethodHS512, []byte(testSecret), claimsFor("1", user.RoleAdmin))},
		{"alg: none", sign(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, claimsFor("1", user.RoleAdmin))},
		{"неизвестная роль", sign(jwt.SigningMethodHS256, []byte(testSecret), claimsFor("1", "ROOT"))},
		{"нечисловой sub", sign(jwt.SigningMethodHS256, []byte(testSecret), claimsFor("admin", user.RoleAdmin))},
		{"мусор вместо токена", "not.a.jwt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tokens.Parse(tt.raw); err == nil {
				t.Errorf("поддельный токен принят: %+v", got)
			}
		})
	}
}
