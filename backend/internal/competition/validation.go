package competition

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"sportreg/internal/apperr"
)

const (
	DateLayout     = "2006-01-02"
	MinCapacity    = 1
	MaxCapacity    = 1000
	maxDescription = 1000
)

// missingFields возвращает ошибки для обязательных полей, которых нет в POST или PUT.
func (in Input) missingFields() []apperr.FieldError {
	var errs []apperr.FieldError
	required := []struct {
		name    string
		present bool
	}{
		{"title", in.Title != nil},
		{"sport", in.Sport != nil},
		{"location", in.Location != nil},
		{"starts_on", in.StartsOn != nil},
		{"capacity", in.Capacity != nil},
	}
	for _, f := range required {
		if !f.present {
			errs = append(errs, apperr.FieldError{Field: f.name, Message: "обязательное поле"})
		}
	}
	return errs
}

// apply переносит переданные поля в соревнование, обрезая пробелы по краям строк.
func (in Input) apply(c Competition) Competition {
	set := func(dst *string, src *string) {
		if src != nil {
			*dst = strings.TrimSpace(*src)
		}
	}
	set(&c.Title, in.Title)
	set(&c.Sport, in.Sport)
	set(&c.Location, in.Location)
	set(&c.StartsOn, in.StartsOn)
	set(&c.Description, in.Description)
	if in.Capacity != nil {
		c.Capacity = *in.Capacity
	}
	return c
}

// Validate проверяет соревнование целиком. Дата сравнивается с today, только если checkDate:
// при изменении других полей уже прошедшего соревнования дата не мешает сохранить правку.
func Validate(c Competition, today time.Time, checkDate bool) []apperr.FieldError {
	var errs []apperr.FieldError
	add := func(field, msg string) { errs = append(errs, apperr.FieldError{Field: field, Message: msg}) }

	checkLength := func(field, value string, minLen, maxLen int) {
		if n := utf8.RuneCountInString(value); n < minLen || n > maxLen {
			add(field, lengthMessage(minLen, maxLen))
		}
	}
	checkLength("title", c.Title, 3, 100)
	checkLength("sport", c.Sport, 2, 50)
	checkLength("location", c.Location, 2, 100)
	if utf8.RuneCountInString(c.Description) > maxDescription {
		add("description", "не длиннее 1000 символов")
	}
	if c.Capacity < MinCapacity || c.Capacity > MaxCapacity {
		add("capacity", "от 1 до 1000")
	}

	date, err := time.Parse(DateLayout, c.StartsOn)
	switch {
	case err != nil:
		add("starts_on", "дата в формате ГГГГ-ММ-ДД")
	case checkDate && date.Before(startOfDay(today)):
		add("starts_on", "дата не может быть в прошлом")
	}
	return errs
}

func lengthMessage(minLen, maxLen int) string {
	return fmt.Sprintf("от %d до %d символов", minLen, maxLen)
}

// startOfDay — полночь календарного дня t, в UTC, как и даты из time.Parse.
func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
