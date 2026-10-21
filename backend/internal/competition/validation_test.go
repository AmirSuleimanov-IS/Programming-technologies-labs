package competition

import (
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 10, 21, 18, 0, 0, 0, time.Local)

func validCompetition() Competition {
	return Competition{Title: "Кубок ЗГУ", Sport: "Волейбол", Location: "Норильск", StartsOn: "2026-11-01", Capacity: 12}
}

// Классы эквивалентности и граничные значения полей соревнования
func TestValidate_Boundaries(t *testing.T) {
	tests := []struct {
		name   string
		modify func(c *Competition)
		field  string // "" — ошибок нет
	}{
		{"валидное соревнование", func(*Competition) {}, ""},
		{"название 2 символа", func(c *Competition) { c.Title = "Ку" }, "title"},
		{"название 3 символа", func(c *Competition) { c.Title = "Куб" }, ""},
		{"название 100 символов кириллицы", func(c *Competition) { c.Title = strings.Repeat("я", 100) }, ""},
		{"название 101 символ", func(c *Competition) { c.Title = strings.Repeat("я", 101) }, "title"},
		{"вид спорта 1 символ", func(c *Competition) { c.Sport = "Ш" }, "sport"},
		{"место 1 символ", func(c *Competition) { c.Location = "Н" }, "location"},
		{"описание 1000 символов", func(c *Competition) { c.Description = strings.Repeat("я", 1000) }, ""},
		{"описание 1001 символ", func(c *Competition) { c.Description = strings.Repeat("я", 1001) }, "description"},
		{"мест 0", func(c *Competition) { c.Capacity = 0 }, "capacity"},
		{"мест 1", func(c *Competition) { c.Capacity = 1 }, ""},
		{"мест 1000", func(c *Competition) { c.Capacity = 1000 }, ""},
		{"мест 1001", func(c *Competition) { c.Capacity = 1001 }, "capacity"},
		{"мест отрицательно", func(c *Competition) { c.Capacity = -5 }, "capacity"},
		{"дата вчера", func(c *Competition) { c.StartsOn = "2026-10-20" }, "starts_on"},
		{"дата сегодня", func(c *Competition) { c.StartsOn = "2026-10-21" }, ""},
		{"дата в другом формате", func(c *Competition) { c.StartsOn = "21.10.2026" }, "starts_on"},
		{"несуществующая дата", func(c *Competition) { c.StartsOn = "2026-02-30" }, "starts_on"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validCompetition()
			tt.modify(&c)
			errs := Validate(c, now, true)
			switch {
			case tt.field == "" && len(errs) != 0:
				t.Errorf("неожиданные ошибки: %v", errs)
			case tt.field != "" && (len(errs) != 1 || errs[0].Field != tt.field):
				t.Errorf("ошибки %v, ожидалась одна ошибка поля %q", errs, tt.field)
			}
		})
	}
}

func TestValidate_PastDateAllowedWhenNotChecked(t *testing.T) {
	c := validCompetition()
	c.StartsOn = "2026-01-10"
	if errs := Validate(c, now, false); len(errs) != 0 {
		t.Errorf("прошедшая дата без проверки даты: %v", errs)
	}
}
