package auth

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"sportreg/internal/apperr"
)

const (
	MinPasswordLen   = 8
	MaxPasswordBytes = 72 // bcrypt учитывает только первые 72 байта пароля
	maxEmailLen      = 254
)

// RegisterInput — тело запроса регистрации.
type RegisterInput struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

// normalize обрезает пробелы и приводит email к нижнему регистру: Ivan@Mail.ru и ivan@mail.ru — один адрес.
func (in RegisterInput) normalize() RegisterInput {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.FullName = strings.TrimSpace(in.FullName)
	return in
}

// ValidateRegistration проверяет поля регистрации (после normalize).
func ValidateRegistration(in RegisterInput) []apperr.FieldError {
	var errs []apperr.FieldError
	if msg := checkEmail(in.Email); msg != "" {
		errs = append(errs, apperr.FieldError{Field: "email", Message: msg})
	}
	if n := utf8.RuneCountInString(in.FullName); n < 2 || n > 100 {
		errs = append(errs, apperr.FieldError{Field: "full_name", Message: "от 2 до 100 символов"})
	}
	if msg := CheckPassword(in.Password); msg != "" {
		errs = append(errs, apperr.FieldError{Field: "password", Message: msg})
	}
	return errs
}

func checkEmail(email string) string {
	if email == "" {
		return "обязательное поле"
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > maxEmailLen {
		return "некорректный адрес электронной почты"
	}
	return ""
}

// CheckPassword возвращает текст ошибки или пустую строку, если пароль подходит:
// от 8 символов, не больше 72 байт, хотя бы одна буква и одна цифра.
func CheckPassword(password string) string {
	if utf8.RuneCountInString(password) < MinPasswordLen {
		return "не короче 8 символов"
	}
	if len(password) > MaxPasswordBytes {
		return "не длиннее 72 байт (буква кириллицы занимает 2 байта)"
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		hasLetter = hasLetter || unicode.IsLetter(r)
		hasDigit = hasDigit || unicode.IsDigit(r)
	}
	if !hasLetter || !hasDigit {
		return "должен содержать букву и цифру"
	}
	return ""
}

// MaskEmail скрывает часть адреса для логов: ivan.petrov@mail.ru → iv***@mail.ru.
func MaskEmail(email string) string {
	name, domain, ok := strings.Cut(email, "@")
	if !ok {
		return "***"
	}
	keep := min(2, utf8.RuneCountInString(name))
	return string([]rune(name)[:keep]) + "***@" + domain
}
