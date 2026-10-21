package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sportreg/internal/apperr"
)

func TestDecodeJSON(t *testing.T) {
	type input struct {
		Title string `json:"title"`
	}
	tests := []struct {
		name string
		body string
		ok   bool
	}{
		{"валидный объект", `{"title":"Кубок"}`, true},
		{"пустое тело", ``, false},
		{"неизвестное поле", `{"titel":"Кубок"}`, false},
		{"обрезанный JSON", `{"title":`, false},
		{"мусор после объекта", `{"title":"Кубок"} {"x":1}`, false},
		{"не тот тип", `{"title":5}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			var in input
			if got := DecodeJSON(rec, req, &in); got != tt.ok {
				t.Fatalf("DecodeJSON() = %v, ожидалось %v", got, tt.ok)
			}
			if !tt.ok && rec.Code != http.StatusBadRequest {
				t.Errorf("код %d, ожидался 400", rec.Code)
			}
		})
	}
}

func TestWriteAppError_StatusByKind(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{apperr.Validation(nil), http.StatusUnprocessableEntity},
		{apperr.NotFound("нет"), http.StatusNotFound},
		{apperr.Conflict("X", "конфликт"), http.StatusConflict},
		{apperr.Unauthorized("войдите"), http.StatusUnauthorized},
		{apperr.Forbidden("нельзя"), http.StatusForbidden},
		{errors.New("pq: connection refused"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		WriteAppError(rec, httptest.NewRequest(http.MethodGet, "/", nil), tt.err)
		if rec.Code != tt.want {
			t.Errorf("%v: код %d, ожидался %d", tt.err, rec.Code, tt.want)
		}
		if strings.Contains(rec.Body.String(), "connection refused") {
			t.Error("текст внутренней ошибки попал клиенту")
		}
	}
}

// Ответы стандартного маршрутизатора 404 и 405 приходят в общем формате API
func TestJSONErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items", func(w http.ResponseWriter, _ *http.Request) {
		WriteData(w, http.StatusOK, []string{}, nil)
	})
	h := JSONErrors(mux)

	tests := []struct {
		method, path string
		status       int
		code         string
	}{
		{http.MethodGet, "/api/nothing", http.StatusNotFound, "NOT_FOUND"},
		{http.MethodDelete, "/api/items", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
		var body struct {
			Error struct{ Code string } `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || rec.Code != tt.status || body.Error.Code != tt.code {
			t.Errorf("%s %s: код %d, тело %q", tt.method, tt.path, rec.Code, rec.Body)
		}
	}
}

func TestPathID(t *testing.T) {
	for _, raw := range []string{"0", "-1", "abc", "1.5", "99999999999999999999"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", raw)
		rec := httptest.NewRecorder()
		if _, ok := PathID(rec, req, "id"); ok || rec.Code != http.StatusBadRequest {
			t.Errorf("id=%q принят", raw)
		}
	}
}
