//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// reset очищает БД и кэш и заново создаёт администратора: каждый тест начинается
// с одинакового состояния и не зависит от порядка запуска.
func reset(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `TRUNCATE competitions, users, applications RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	if err := rdb.FlushAll(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	if err := authSvc.EnsureAdmin(ctx, adminEmail, adminPassword); err != nil {
		t.Fatal(err)
	}
}

type response struct {
	Status int
	Header http.Header
	Body   []byte
}

// call отправляет запрос к API; body кодируется в JSON, token — JWT или "".
func call(t *testing.T, method, path, token string, body any) response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, server.URL+"/api"+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response{Status: res.StatusCode, Header: res.Header, Body: raw}
}

// expect проверяет код ответа и возвращает поле data ответа.
func expect[T any](t *testing.T, r response, status int) T {
	t.Helper()
	if r.Status != status {
		t.Fatalf("код %d, ожидался %d: %s", r.Status, status, r.Body)
	}
	var body struct {
		Data T `json:"data"`
	}
	if len(r.Body) > 0 {
		if err := json.Unmarshal(r.Body, &body); err != nil {
			t.Fatalf("тело ответа: %v: %s", err, r.Body)
		}
	}
	return body.Data
}

// errorCode возвращает error.code из ответа с ошибкой.
func errorCode(t *testing.T, r response) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatalf("тело ответа: %v: %s", err, r.Body)
	}
	return body.Error.Code
}

type competitionDTO struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Capacity   int    `json:"capacity"`
	FreePlaces int    `json:"free_places"`
}

type applicationDTO struct {
	ID            int64  `json:"id"`
	CompetitionID int64  `json:"competition_id"`
	UserID        int64  `json:"user_id"`
	Status        string `json:"status"`
}

func login(t *testing.T, email, password string) string {
	t.Helper()
	res := expect[struct {
		Token string `json:"token"`
	}](t, call(t, http.MethodPost, "/auth/login", "", map[string]string{"email": email, "password": password}), http.StatusOK)
	return res.Token
}

func adminToken(t *testing.T) string {
	t.Helper()
	return login(t, adminEmail, adminPassword)
}

// newUser регистрирует пользователя и возвращает его токен.
func newUser(t *testing.T, name string) string {
	t.Helper()
	email := fmt.Sprintf("%s@test.local", name)
	expect[any](t, call(t, http.MethodPost, "/auth/register", "",
		map[string]string{"email": email, "full_name": "Пользователь " + name, "password": "Sport2026"}), http.StatusCreated)
	return login(t, email, "Sport2026")
}

// newCompetition создаёт соревнование через месяц от сегодняшнего дня.
func newCompetition(t *testing.T, admin, title string, capacity int) competitionDTO {
	t.Helper()
	return expect[competitionDTO](t, call(t, http.MethodPost, "/competitions", admin, map[string]any{
		"title": title, "sport": "Волейбол", "location": "Норильск",
		"starts_on": time.Now().AddDate(0, 1, 0).Format("2006-01-02"), "capacity": capacity,
	}), http.StatusCreated)
}

func submit(t *testing.T, token string, competitionID int64) response {
	t.Helper()
	return call(t, http.MethodPost, "/applications", token, map[string]int64{"competition_id": competitionID})
}

func setStatus(t *testing.T, token string, applicationID int64, status string) response {
	t.Helper()
	return call(t, http.MethodPatch, fmt.Sprintf("/applications/%d", applicationID), token, map[string]string{"status": status})
}

// countRows считает строки прямым SQL-запросом — проверка, что изменилось в БД на самом деле.
func countRows(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
