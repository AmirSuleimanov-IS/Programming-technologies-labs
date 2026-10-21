package competition

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"sportreg/internal/apperr"
	"sportreg/internal/cache"
)

// Store — то, что сервису нужно от хранилища. В тестах его заменяет подделка.
type Store interface {
	List(ctx context.Context, f Filter) ([]Competition, error)
	Get(ctx context.Context, id int64) (Competition, error)
	Create(ctx context.Context, c Competition) (Competition, error)
	Update(ctx context.Context, c Competition) (Competition, error)
	Delete(ctx context.Context, id int64) error
}

// Service — бизнес-логика каталога соревнований.
type Service struct {
	store Store
	cache *cache.Client // nil — без кэша
	ttl   CacheTTL
	now   func() time.Time
}

func NewService(store Store, c *cache.Client, ttl CacheTTL) *Service {
	return &Service{store: store, cache: c, ttl: ttl, now: time.Now}
}

// List и Get читают через кэш (cache-aside) и сообщают, откуда пришли данные.
func (s *Service) List(ctx context.Context, f Filter) ([]Competition, cache.Source, error) {
	return s.loadList(ctx, f)
}

func (s *Service) Get(ctx context.Context, id int64) (Competition, cache.Source, error) {
	return s.loadItem(ctx, id)
}

func (s *Service) Create(ctx context.Context, in Input) (Competition, error) {
	if missing := in.missingFields(); len(missing) > 0 {
		return Competition{}, apperr.Validation(missing)
	}
	c := in.apply(Competition{})
	if errs := Validate(c, s.now(), true); len(errs) > 0 {
		return Competition{}, apperr.Validation(errs)
	}
	created, err := s.store.Create(ctx, c)
	if err != nil {
		return Competition{}, err
	}
	s.Invalidate(ctx, 0) // нового соревнования ещё нет в кэше, но оно должно появиться в списках
	slog.InfoContext(ctx, "competition created", "competition_id", created.ID, "title", created.Title)
	return created, nil
}

// Replace — PUT: все поля передаются заново, непереданное описание становится пустым.
func (s *Service) Replace(ctx context.Context, id int64, in Input) (Competition, error) {
	if missing := in.missingFields(); len(missing) > 0 {
		return Competition{}, apperr.Validation(missing)
	}
	current, err := s.store.Get(ctx, id)
	if err != nil {
		return Competition{}, mapErr(err)
	}
	next := in.apply(Competition{ID: id})
	return s.update(ctx, current, next)
}

// Patch — PATCH: меняются только переданные поля.
func (s *Service) Patch(ctx context.Context, id int64, in Input) (Competition, error) {
	current, err := s.store.Get(ctx, id)
	if err != nil {
		return Competition{}, mapErr(err)
	}
	return s.update(ctx, current, in.apply(current))
}

func (s *Service) update(ctx context.Context, current, next Competition) (Competition, error) {
	if errs := Validate(next, s.now(), next.StartsOn != current.StartsOn); len(errs) > 0 {
		return Competition{}, apperr.Validation(errs)
	}
	updated, err := s.store.Update(ctx, next)
	if err != nil {
		return Competition{}, mapErr(err)
	}
	s.Invalidate(ctx, updated.ID)
	slog.InfoContext(ctx, "competition updated", "competition_id", updated.ID)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return mapErr(err)
	}
	s.Invalidate(ctx, id)
	slog.InfoContext(ctx, "competition deleted", "competition_id", id)
	return nil
}

// mapErr переводит ошибки хранилища в ошибки, понятные клиенту.
func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return apperr.NotFound("Соревнование не найдено")
	case errors.Is(err, ErrCapacityBelowApproved):
		return apperr.Conflict("CAPACITY_BELOW_APPROVED", "Лимит мест не может быть меньше числа одобренных заявок")
	case errors.Is(err, ErrHasActiveApplications):
		return apperr.Conflict("HAS_ACTIVE_APPLICATIONS", "У соревнования есть активные заявки, удалить его нельзя")
	default:
		return err
	}
}
