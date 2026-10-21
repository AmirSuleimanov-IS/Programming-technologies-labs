// Package apperr описывает ошибки бизнес-логики, которые можно показать клиенту.
// Сервисы возвращают *Error, а HTTP-слой превращает Kind в код ответа.
package apperr

import (
	"errors"
	"fmt"
)

type Kind int

const (
	KindValidation Kind = iota + 1
	KindNotFound
	KindConflict
	KindUnauthorized
	KindForbidden
)

// FieldError — ошибка в конкретном поле запроса.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Kind    Kind
	Code    string
	Message string
	Fields  []FieldError
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func Validation(fields []FieldError) *Error {
	return &Error{Kind: KindValidation, Code: "VALIDATION_ERROR", Message: "Данные не прошли проверку", Fields: fields}
}

func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Code: "NOT_FOUND", Message: message}
}

func Conflict(code, message string) *Error {
	return &Error{Kind: KindConflict, Code: code, Message: message}
}

// As достаёт *Error из цепочки ошибок.
func As(err error) (*Error, bool) {
	var appErr *Error
	ok := errors.As(err, &appErr)
	return appErr, ok
}

func Unauthorized(message string) *Error {
	return &Error{Kind: KindUnauthorized, Code: "UNAUTHORIZED", Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Code: "FORBIDDEN", Message: message}
}
