// Package competition — каталог соревнований: модель, правила, хранилище, сервис и контроллер.
package competition

import "time"

// Competition — соревнование в том виде, в каком его отдаёт API.
type Competition struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Sport       string    `json:"sport"`
	Location    string    `json:"location"`
	StartsOn    string    `json:"starts_on"` // дата в формате ГГГГ-ММ-ДД
	Capacity    int       `json:"capacity"`
	FreePlaces  int       `json:"free_places"` // лимит минус одобренные заявки
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Input — тело POST, PUT и PATCH. Указатели отличают «поле не передано» от пустого значения:
// для POST и PUT обязательны все поля, кроме описания, для PATCH — ни одно.
type Input struct {
	Title       *string `json:"title"`
	Sport       *string `json:"sport"`
	Location    *string `json:"location"`
	StartsOn    *string `json:"starts_on"`
	Capacity    *int    `json:"capacity"`
	Description *string `json:"description"`
}

// Filter — параметры поиска в каталоге.
type Filter struct {
	Query string // часть названия
	Sport string // точное совпадение вида спорта без учёта регистра
}
