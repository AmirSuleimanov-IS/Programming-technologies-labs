package application

import (
	"time"

	"sportreg/internal/apperr"
)

// Правила статусной модели — чистые функции без БД, их удобно проверять unit-тестами.

// CheckSubmit — можно ли подать заявку на соревнование сегодня.
func CheckSubmit(slot Slot, today time.Time) error {
	if started(slot, today) {
		return apperr.Conflict("COMPETITION_STARTED", "Соревнование уже началось, приём заявок закрыт")
	}
	if slot.FreePlaces() <= 0 {
		return apperr.Conflict("NO_FREE_PLACES", "Свободных мест нет")
	}
	return nil
}

// CheckTransition — может ли пользователь перевести заявку из from в to.
// isOwner — заявка принадлежит пользователю, isAdmin — у него роль ADMIN.
func CheckTransition(from, to Status, isOwner, isAdmin bool) error {
	switch to {
	case StatusCancelled:
		if !isOwner {
			return apperr.Forbidden("Отменить заявку может только её автор")
		}
		if from != StatusPending && from != StatusApproved {
			return invalidTransition(from, to)
		}
	case StatusApproved, StatusRejected:
		if !isAdmin {
			return apperr.Forbidden("Рассматривать заявки может только администратор")
		}
		if from != StatusPending {
			return invalidTransition(from, to)
		}
	default:
		return apperr.Validation([]apperr.FieldError{{Field: "status", Message: "допустимо approved, rejected или cancelled"}})
	}
	return nil
}

// CheckSlotForTransition — ограничения со стороны соревнования: после начала статусы
// не меняются, одобрить можно только при наличии свободного места.
func CheckSlotForTransition(slot Slot, to Status, today time.Time) error {
	if started(slot, today) {
		return apperr.Conflict("COMPETITION_STARTED", "Соревнование уже началось, заявку изменить нельзя")
	}
	if to == StatusApproved && slot.FreePlaces() <= 0 {
		return apperr.Conflict("NO_FREE_PLACES", "Свободных мест нет")
	}
	return nil
}

// started — наступил ли день начала соревнования (по календарной дате).
func started(slot Slot, today time.Time) bool {
	y, m, d := today.Date()
	return !slot.StartsOn.After(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func invalidTransition(from, to Status) error {
	return apperr.Conflict("INVALID_STATUS_TRANSITION",
		"Нельзя перевести заявку из статуса «"+statusTitle(from)+"» в «"+statusTitle(to)+"»")
}

func statusTitle(s Status) string {
	switch s {
	case StatusPending:
		return "на рассмотрении"
	case StatusApproved:
		return "одобрена"
	case StatusRejected:
		return "отклонена"
	case StatusCancelled:
		return "отменена"
	default:
		return string(s)
	}
}
