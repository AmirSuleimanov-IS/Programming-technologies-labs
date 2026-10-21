package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"sportreg/internal/apperr"
	"sportreg/internal/user"
)

// Store — то, что сервису нужно от хранилища заявок.
type Store interface {
	InTx(ctx context.Context, fn func(tx Tx) error) error
	ListByUser(ctx context.Context, userID int64) ([]Application, error)
	List(ctx context.Context, f Filter) ([]Application, error)
}

// Service — бизнес-логика заявок.
type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) ListMine(ctx context.Context, actor user.Current) ([]Application, error) {
	return s.store.ListByUser(ctx, actor.ID)
}

func (s *Service) List(ctx context.Context, f Filter) ([]Application, error) {
	return s.store.List(ctx, f)
}

// Submit подаёт заявку от имени actor. Проверки и вставка идут в одной транзакции
// с заблокированной строкой соревнования.
func (s *Service) Submit(ctx context.Context, actor user.Current, competitionID int64) (Application, error) {
	if competitionID <= 0 {
		return Application{}, apperr.Validation([]apperr.FieldError{{Field: "competition_id", Message: "обязательное поле"}})
	}
	var created Application
	err := s.store.InTx(ctx, func(tx Tx) error {
		slot, err := tx.LockSlot(ctx, competitionID)
		if errors.Is(err, ErrCompetitionNotFound) {
			return apperr.NotFound("Соревнование не найдено")
		}
		if err != nil {
			return err
		}
		if err := CheckSubmit(slot, s.now()); err != nil {
			return err
		}
		created, err = tx.Create(ctx, actor.ID, competitionID)
		if errors.Is(err, ErrDuplicate) {
			return apperr.Conflict("DUPLICATE_APPLICATION", "У вас уже есть активная заявка на это соревнование")
		}
		return err
	})
	if err != nil {
		return Application{}, err
	}
	slog.InfoContext(ctx, "application submitted", "application_id", created.ID, "competition_id", competitionID)
	return created, nil
}

// ChangeStatus — одобрение, отклонение (администратор) или отмена (автор заявки).
func (s *Service) ChangeStatus(ctx context.Context, actor user.Current, id int64, to Status) (Application, error) {
	var updated Application
	var from Status
	err := s.store.InTx(ctx, func(tx Tx) error {
		app, err := tx.GetForUpdate(ctx, id)
		if errors.Is(err, ErrNotFound) {
			return apperr.NotFound("Заявка не найдена")
		}
		if err != nil {
			return err
		}
		from = app.Status
		if err := CheckTransition(app.Status, to, app.UserID == actor.ID, actor.IsAdmin()); err != nil {
			return err
		}
		slot, err := tx.LockSlot(ctx, app.CompetitionID)
		if err != nil {
			return err
		}
		if err := CheckSlotForTransition(slot, to, s.now()); err != nil {
			return err
		}
		updated, err = tx.SetStatus(ctx, id, to)
		return err
	})
	if err != nil {
		return Application{}, err
	}
	slog.InfoContext(ctx, "application status changed", "application_id", id,
		"competition_id", updated.CompetitionID, "from", from, "to", to)
	return updated, nil
}
