package auth

import (
	"strings"
	"testing"
)

// Граничные значения пароля: минимум 8 символов, максимум 72 байта (предел bcrypt),
// буква и цифра обязательны. Кириллица занимает 2 байта, поэтому граница в символах другая.
func TestCheckPassword(t *testing.T) {
	const (
		tooShort  = "не короче 8 символов"
		tooLong   = "не длиннее 72 байт (буква кириллицы занимает 2 байта)"
		needMixed = "должен содержать букву и цифру"
	)
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{"пустой пароль", "", tooShort},
		{"7 символов — на один меньше минимума", "abcdef1", tooShort},
		{"8 символов — минимум", "abcdef12", ""},
		{"9 символов", "abcdefg12", ""},
		{"72 байта — максимум", strings.Repeat("a", 71) + "1", ""},
		{"73 байта — на один больше максимума", strings.Repeat("a", 72) + "1", tooLong},
		{"кириллица: 35 букв и 2 цифры = 72 байта", strings.Repeat("ж", 35) + "12", ""},
		{"кириллица: 36 букв и цифра = 73 байта", strings.Repeat("ж", 36) + "1", tooLong},
		{"8 символов кириллицы — проверяются символы, а не байты", "пароль12", ""},
		{"только буквы", "abcdefgh", needMixed},
		{"только цифры", "12345678", needMixed},
		{"пробелы и цифры без букв", "1234 5678", needMixed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.password); got != tt.want {
				t.Errorf("CheckPassword(%q) = %q, want %q", tt.password, got, tt.want)
			}
		})
	}
}

func TestValidateRegistration(t *testing.T) {
	valid := RegisterInput{Email: "ivan@mail.ru", FullName: "Иван Петров", Password: "Sport2026"}
	if errs := ValidateRegistration(valid); len(errs) != 0 {
		t.Fatalf("валидные данные: ошибки %v", errs)
	}

	tests := []struct {
		name  string
		input RegisterInput
		field string
	}{
		{"email без домена", RegisterInput{Email: "ivan@", FullName: "Иван", Password: "Sport2026"}, "email"},
		{"email с именем", RegisterInput{Email: "Иван <ivan@mail.ru>", FullName: "Иван", Password: "Sport2026"}, "email"},
		{"пустой email", RegisterInput{Email: "", FullName: "Иван", Password: "Sport2026"}, "email"},
		{"имя из 1 символа", RegisterInput{Email: "ivan@mail.ru", FullName: "И", Password: "Sport2026"}, "full_name"},
		{"имя из 101 символа", RegisterInput{Email: "ivan@mail.ru", FullName: strings.Repeat("я", 101), Password: "Sport2026"}, "full_name"},
		{"слабый пароль", RegisterInput{Email: "ivan@mail.ru", FullName: "Иван", Password: "password"}, "password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateRegistration(tt.input)
			if len(errs) != 1 || errs[0].Field != tt.field {
				t.Errorf("ошибки %v, ожидалась одна ошибка поля %q", errs, tt.field)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	got := RegisterInput{Email: "  Ivan.Petrov@Mail.RU ", FullName: " Иван "}.normalize()
	if got.Email != "ivan.petrov@mail.ru" || got.FullName != "Иван" {
		t.Errorf("normalize() = %+v", got)
	}
}

func TestMaskEmail(t *testing.T) {
	tests := map[string]string{
		"ivan.petrov@mail.ru": "iv***@mail.ru",
		"a@mail.ru":           "a***@mail.ru",
		"иван@почта.рф":       "ив***@почта.рф",
		"not-an-email":        "***",
	}
	for in, want := range tests {
		if got := MaskEmail(in); got != want {
			t.Errorf("MaskEmail(%q) = %q, want %q", in, got, want)
		}
	}
}
