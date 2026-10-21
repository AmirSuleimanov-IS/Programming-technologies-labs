package application

import (
	"context"
	"testing"
	"time"

	"sportreg/internal/apperr"
	"sportreg/internal/user"
)

// fakeStore хранит заявки и соревнования в памяти. Транзакция — просто вызов fn:
// откат и блокировки проверяют integration-тесты на настоящей PostgreSQL.
type fakeStore struct {
	apps  map[int64]Application
	slots map[int64]Slot
}

func (f *fakeStore) InTx(_ context.Context, fn func(tx Tx) error) error { return fn(f) }
func (f *fakeStore) ListByUser(context.Context, int64) ([]Application, error) {
	return nil, nil
}
func (f *fakeStore) List(context.Context, Filter) ([]Application, error) { return nil, nil }

func (f *fakeStore) LockSlot(_ context.Context, competitionID int64) (Slot, error) {
	s, ok := f.slots[competitionID]
	if !ok {
		return Slot{}, ErrCompetitionNotFound
	}
	for _, a := range f.apps {
		if a.CompetitionID == competitionID && a.Status == StatusApproved {
			s.Approved++
		}
	}
	return s, nil
}

func (f *fakeStore) GetForUpdate(_ context.Context, id int64) (Application, error) {
	a, ok := f.apps[id]
	if !ok {
		return Application{}, ErrNotFound
	}
	return a, nil
}

func (f *fakeStore) Create(_ context.Context, userID, competitionID int64) (Application, error) {
	for _, a := range f.apps {
		if a.UserID == userID && a.CompetitionID == competitionID && (a.Status == StatusPending || a.Status == StatusApproved) {
			return Application{}, ErrDuplicate
		}
	}
	a := Application{ID: int64(len(f.apps) + 1), UserID: userID, CompetitionID: competitionID, Status: StatusPending}
	f.apps[a.ID] = a
	return a, nil
}

func (f *fakeStore) SetStatus(_ context.Context, id int64, status Status) (Application, error) {
	a := f.apps[id]
	a.Status = status
	f.apps[id] = a
	return a, nil
}

var (
	owner = user.Current{ID: 10, Role: user.RoleUser}
	other = user.Current{ID: 11, Role: user.RoleUser}
	admin = user.Current{ID: 1, Role: user.RoleAdmin}
)

// newTestService: соревнование 100 через 30 дней на 2 места; invalidated собирает
// вызовы инвалидации кэша свободных мест.
func newTestService(apps ...Application) (*Service, *fakeStore, *[]int64) {
	store := &fakeStore{apps: map[int64]Application{}, slots: map[int64]Slot{
		100: {CompetitionID: 100, StartsOn: time.Now().AddDate(0, 0, 30), Capacity: 2},
	}}
	for _, a := range apps {
		store.apps[a.ID] = a
	}
	var invalidated []int64
	svc := NewService(store, func(_ context.Context, id int64) { invalidated = append(invalidated, id) })
	return svc, store, &invalidated
}

func TestService_Submit(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	created, err := svc.Submit(ctx, owner, 100)
	if err != nil || created.Status != StatusPending || created.UserID != owner.ID {
		t.Fatalf("Submit() = %+v, %v", created, err)
	}
	_, err = svc.Submit(ctx, owner, 100)
	if codeOf(err) != "DUPLICATE_APPLICATION" {
		t.Errorf("повторная заявка: %v", err)
	}
	_, err = svc.Submit(ctx, owner, 999)
	if kindOf(err) != apperr.KindNotFound {
		t.Errorf("несуществующее соревнование: %v", err)
	}
	_, err = svc.Submit(ctx, owner, 0)
	if kindOf(err) != apperr.KindValidation {
		t.Errorf("competition_id = 0: %v", err)
	}
}

// Кэш свободных мест сбрасывается только тогда, когда меняется число одобренных заявок
func TestService_ChangeStatus_InvalidatesFreePlaces(t *testing.T) {
	tests := []struct {
		name           string
		from, to       Status
		actor          user.Current
		wantInvalidate bool
	}{
		{"одобрение", StatusPending, StatusApproved, admin, true},
		{"отмена одобренной", StatusApproved, StatusCancelled, owner, true},
		{"отклонение", StatusPending, StatusRejected, admin, false},
		{"отмена на рассмотрении", StatusPending, StatusCancelled, owner, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store, invalidated := newTestService(Application{ID: 1, UserID: owner.ID, CompetitionID: 100, Status: tt.from})
			updated, err := svc.ChangeStatus(context.Background(), tt.actor, 1, tt.to)
			if err != nil || updated.Status != tt.to || store.apps[1].Status != tt.to {
				t.Fatalf("ChangeStatus() = %+v, %v", updated, err)
			}
			if got := len(*invalidated) == 1 && (*invalidated)[0] == 100; got != tt.wantInvalidate {
				t.Errorf("инвалидация: %v, ожидалась %v", *invalidated, tt.wantInvalidate)
			}
		})
	}
}

func TestService_ChangeStatus_Rejected(t *testing.T) {
	tests := []struct {
		name     string
		actor    user.Current
		to       Status
		approved int // уже одобренных заявок других пользователей
		wantKind apperr.Kind
		wantCode string
	}{
		{"чужой пользователь отменяет", other, StatusCancelled, 0, apperr.KindForbidden, "FORBIDDEN"},
		{"автор одобряет сам", owner, StatusApproved, 0, apperr.KindForbidden, "FORBIDDEN"},
		{"одобрение без свободных мест", admin, StatusApproved, 2, apperr.KindConflict, "NO_FREE_PLACES"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps := []Application{{ID: 1, UserID: owner.ID, CompetitionID: 100, Status: StatusPending}}
			for i := range tt.approved {
				apps = append(apps, Application{ID: int64(10 + i), UserID: int64(50 + i), CompetitionID: 100, Status: StatusApproved})
			}
			svc, store, invalidated := newTestService(apps...)
			_, err := svc.ChangeStatus(context.Background(), tt.actor, 1, tt.to)
			if kindOf(err) != tt.wantKind || codeOf(err) != tt.wantCode {
				t.Errorf("ошибка %v, ожидалась %s", err, tt.wantCode)
			}
			if store.apps[1].Status != StatusPending || len(*invalidated) != 0 {
				t.Errorf("при ошибке заявка изменилась (%s) или кэш сброшен (%v)", store.apps[1].Status, *invalidated)
			}
		})
	}
}
