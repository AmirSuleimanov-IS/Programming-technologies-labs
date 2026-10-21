// Package httpapi содержит общий для всех контроллеров формат ответов и middleware.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"sportreg/internal/apperr"
)

// Успешный ответ: {"data": ..., "meta": {...}}.
type successBody struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

// Ответ с ошибкой: {"error": {"code": ..., "message": ..., "details": [...]}}.
type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []apperr.FieldError `json:"details,omitempty"`
}

func WriteData(w http.ResponseWriter, status int, data, meta any) {
	writeJSON(w, status, successBody{Data: data, Meta: meta})
}

func WriteError(w http.ResponseWriter, status int, code, message string, details []apperr.FieldError) {
	writeJSON(w, status, errorBody{Error: errorPayload{Code: code, Message: message, Details: details}})
}

// WriteAppError выбирает код ответа по виду ошибки. Неизвестные ошибки
// логируются и отдаются клиенту как 500 без подробностей.
func WriteAppError(w http.ResponseWriter, r *http.Request, err error) {
	appErr, ok := apperr.As(err)
	if !ok {
		slog.ErrorContext(r.Context(), "internal error", "method", r.Method, "path", r.URL.Path, "error", err)
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
		return
	}
	WriteError(w, statusOf(appErr.Kind), appErr.Code, appErr.Message, appErr.Fields)
}

func statusOf(kind apperr.Kind) int {
	switch kind {
	case apperr.KindValidation:
		return http.StatusUnprocessableEntity
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Ответы API не кэширует ни браузер, ни прокси: кэшированием управляют Redis и кэш клиента
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}

// DecodeJSON читает тело запроса. Неизвестные поля и мусор после объекта — ошибка 400.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err == nil && dec.Decode(&struct{}{}) != io.EOF {
		err = errors.New("после JSON-объекта есть лишние данные")
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Некорректный JSON: "+err.Error(), nil)
		return false
	}
	return true
}

// PathID читает положительный числовой параметр пути {id}.
func PathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Параметр "+name+" должен быть положительным числом", nil)
		return 0, false
	}
	return id, true
}
