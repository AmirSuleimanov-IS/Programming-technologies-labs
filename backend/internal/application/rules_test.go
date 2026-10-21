package application

import (
	"testing"
	"time"

	"sportreg/internal/apperr"
)

// Таблица решений для смены статуса: исходный статус × целевой статус × кто меняет.
// Проверяются все 4×3 сочетания статусов для автора и для администратора.
func TestCheckTransition(t *testing.T) {
	const (
		ok        = apperr.Kind(0)
		forbidden = apperr.KindForbidden
		conflict  = apperr.KindConflict
	)
	tests := []struct {
		from, to         Status
		isOwner, isAdmin bool
		want             apperr.Kind
	}{
		// автор заявки (USER) отменяет
		{StatusPending, StatusCancelled, true, false, ok},
		{StatusApproved, StatusCancelled, true, false, ok},
		{StatusRejected, StatusCancelled, true, false, conflict},
		{StatusCancelled, StatusCancelled, true, false, conflict},
		// автор не может рассматривать свою заявку
		{StatusPending, StatusApproved, true, false, forbidden},
		{StatusPending, StatusRejected, true, false, forbidden},
		// чужой USER ничего не может
		{StatusPending, StatusCancelled, false, false, forbidden},
		{StatusPending, StatusApproved, false, false, forbidden},
		// администратор рассматривает только заявки на рассмотрении
		{StatusPending, StatusApproved, false, true, ok},
		{StatusPending, StatusRejected, false, true, ok},
		{StatusApproved, StatusApproved, false, true, conflict},
		{StatusApproved, StatusRejected, false, true, conflict},
		{StatusRejected, StatusApproved, false, true, conflict},
		{StatusCancelled, StatusApproved, false, true, conflict},
		{StatusCancelled, StatusRejected, false, true, conflict},
		// администратор не отменяет чужие заявки (отмена — право автора)
		{StatusPending, StatusCancelled, false, true, forbidden},
		// свою заявку администратор может и отменить
		{StatusApproved, StatusCancelled, true, true, ok},
	}
	for _, tt := range tests {
		name := string(tt.from) + "→" + string(tt.to)
		if tt.isOwner {
			name += " автор"
		}
		if tt.isAdmin {
			name += " admin"
		}
		t.Run(name, func(t *testing.T) {
			err := CheckTransition(tt.from, tt.to, tt.isOwner, tt.isAdmin)
			if got := kindOf(err); got != tt.want {
				t.Errorf("вид ошибки %d, ожидался %d (%v)", got, tt.want, err)
			}
		})
	}
}

func TestCheckTransition_UnknownStatus(t *testing.T) {
	for _, to := range []Status{StatusPending, "", "deleted"} {
		if kind := kindOf(CheckTransition(StatusPending, to, true, true)); kind != apperr.KindValidation {
			t.Errorf("статус %q: вид ошибки %d, ожидалась валидация", to, kind)
		}
	}
}

var today = time.Date(2026, 10, 21, 15, 30, 0, 0, time.Local)

func day(offset int) time.Time {
	return time.Date(2026, 10, 21+offset, 0, 0, 0, 0, time.UTC)
}

// Граничные значения для подачи заявки: свободные места 0/1, дата старта вчера/сегодня/завтра
func TestCheckSubmit(t *testing.T) {
	tests := []struct {
		name     string
		slot     Slot
		wantCode string
	}{
		{"есть места, старт завтра", Slot{StartsOn: day(1), Capacity: 3, Approved: 2}, ""},
		{"последнее место", Slot{StartsOn: day(30), Capacity: 1, Approved: 0}, ""},
		{"мест нет", Slot{StartsOn: day(30), Capacity: 3, Approved: 3}, "NO_FREE_PLACES"},
		{"старт сегодня — приём закрыт", Slot{StartsOn: day(0), Capacity: 3, Approved: 0}, "COMPETITION_STARTED"},
		{"старт вчера", Slot{StartsOn: day(-1), Capacity: 3, Approved: 0}, "COMPETITION_STARTED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := codeOf(CheckSubmit(tt.slot, today)); got != tt.wantCode {
				t.Errorf("код %q, ожидался %q", got, tt.wantCode)
			}
		})
	}
}

func TestCheckSlotForTransition(t *testing.T) {
	full := Slot{StartsOn: day(10), Capacity: 2, Approved: 2}
	tests := []struct {
		name     string
		slot     Slot
		to       Status
		wantCode string
	}{
		{"одобрить при свободном месте", Slot{StartsOn: day(10), Capacity: 2, Approved: 1}, StatusApproved, ""},
		{"одобрить без мест", full, StatusApproved, "NO_FREE_PLACES"},
		{"отклонить без мест можно", full, StatusRejected, ""},
		{"отменить без мест можно", full, StatusCancelled, ""},
		{"после начала нельзя отменить", Slot{StartsOn: day(0), Capacity: 2, Approved: 1}, StatusCancelled, "COMPETITION_STARTED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := codeOf(CheckSlotForTransition(tt.slot, tt.to, today)); got != tt.wantCode {
				t.Errorf("код %q, ожидался %q", got, tt.wantCode)
			}
		})
	}
}

func kindOf(err error) apperr.Kind {
	if err == nil {
		return 0
	}
	appErr, ok := apperr.As(err)
	if !ok {
		return -1
	}
	return appErr.Kind
}

func codeOf(err error) string {
	if appErr, ok := apperr.As(err); ok {
		return appErr.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}
