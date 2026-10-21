//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAuth_RegisterLoginMe(t *testing.T) {
	reset(t)
	expect[any](t, call(t, http.MethodPost, "/auth/register", "",
		map[string]string{"email": "Ivan@Mail.ru", "full_name": "Иван Петров", "password": "Sport2026"}), http.StatusCreated)

	// пароль хранится только в виде bcrypt-хеша
	if n := countRows(t, `SELECT count(*) FROM users WHERE email = 'ivan@mail.ru' AND password_hash LIKE '$2a$%' AND role = 'USER'`); n != 1 {
		t.Fatal("пользователь не сохранён с bcrypt-хешем и ролью USER")
	}
	// email без учёта регистра: повторная регистрация невозможна
	r := call(t, http.MethodPost, "/auth/register", "", map[string]string{"email": "ivan@MAIL.ru", "full_name": "Иван", "password": "Sport2026"})
	if r.Status != http.StatusConflict || errorCode(t, r) != "EMAIL_TAKEN" {
		t.Fatalf("повторная регистрация: %d %s", r.Status, r.Body)
	}

	token := login(t, "IVAN@mail.ru", "Sport2026")
	me := expect[struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}](t, call(t, http.MethodGet, "/auth/me", token, nil), http.StatusOK)
	if me.Email != "ivan@mail.ru" || me.Role != "USER" {
		t.Errorf("me: %+v", me)
	}

	if r := call(t, http.MethodPost, "/auth/login", "", map[string]string{"email": "ivan@mail.ru", "password": "wrong-pass1"}); r.Status != http.StatusUnauthorized {
		t.Errorf("вход с неверным паролем: %d", r.Status)
	}
}

// Матрица доступа: маршрут × роль → код ответа (проверка ролей на бэкенде через настоящий роутер)
func TestAuth_AccessMatrix(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	user := newUser(t, "anna")
	c := newCompetition(t, admin, "Турнир", 5)
	item := fmt.Sprintf("/competitions/%d", c.ID)

	tests := []struct {
		method, path       string
		guest, user, admin int
	}{
		{http.MethodGet, "/competitions", 401, 200, 200},
		{http.MethodGet, item, 401, 200, 200},
		{http.MethodPatch, item, 401, 403, 200},
		{http.MethodGet, "/applications/my", 401, 200, 200},
		{http.MethodGet, "/applications", 401, 403, 200},
		{http.MethodGet, "/users", 401, 403, 200},
		{http.MethodGet, "/auth/me", 401, 200, 200},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			var body any
			if tt.method == http.MethodPatch {
				body = map[string]int{"capacity": 6}
			}
			for _, role := range []struct {
				token string
				want  int
			}{{"", tt.guest}, {user, tt.user}, {admin, tt.admin}} {
				if got := call(t, tt.method, tt.path, role.token, body).Status; got != role.want {
					t.Errorf("токен %.10q…: код %d, ожидался %d", role.token, got, role.want)
				}
			}
		})
	}
}
