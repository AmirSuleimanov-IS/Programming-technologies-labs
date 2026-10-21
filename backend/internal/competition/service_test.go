package competition

import (
	"context"
	"errors"
	"testing"
	"time"

	"sportreg/internal/apperr"
)

// fakeStore — хранилище соревнований в памяти. approved — число одобренных заявок
// по id соревнования, чтобы проверить запрет уменьшать лимит ниже него.
type fakeStore struct {
	items    map[int64]Competition
	approved map[int64]int
}

func newFakeStore(items ...Competition) *fakeStore {
	f := &fakeStore{items: map[int64]Competition{}, approved: map[int64]int{}}
	for _, c := range items {
		f.items[c.ID] = c
	}
	return f
}

func (f *fakeStore) List(context.Context, Filter) ([]Competition, error) { return nil, nil }

func (f *fakeStore) Get(_ context.Context, id int64) (Competition, error) {
	c, ok := f.items[id]
	if !ok {
		return Competition{}, ErrNotFound
	}
	return c, nil
}

func (f *fakeStore) Create(_ context.Context, c Competition) (Competition, error) {
	c.ID = int64(len(f.items) + 1)
	f.items[c.ID] = c
	return c, nil
}

func (f *fakeStore) Update(_ context.Context, c Competition) (Competition, error) {
	if _, ok := f.items[c.ID]; !ok {
		return Competition{}, ErrNotFound
	}
	if c.Capacity < f.approved[c.ID] {
		return Competition{}, ErrCapacityBelowApproved
	}
	f.items[c.ID] = c
	return c, nil
}

func (f *fakeStore) Delete(_ context.Context, id int64) error {
	if _, ok := f.items[id]; !ok {
		return ErrNotFound
	}
	delete(f.items, id)
	return nil
}

func newTestService(store Store) *Service {
	svc := NewService(store, nil, CacheTTL{})
	svc.now = func() time.Time { return now }
	return svc
}

func ptr[T any](v T) *T { return &v }

func TestService_Create(t *testing.T) {
	svc := newTestService(newFakeStore())
	created, err := svc.Create(context.Background(), Input{
		Title: ptr("  Кубок ЗГУ  "), Sport: ptr("Волейбол"), Location: ptr("Норильск"),
		StartsOn: ptr("2026-11-01"), Capacity: ptr(12),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Title != "Кубок ЗГУ" || created.Description != "" {
		t.Errorf("создано %+v: пробелы по краям должны обрезаться, описание — необязательное", created)
	}

	_, err = svc.Create(context.Background(), Input{Title: ptr("Кубок")})
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation || len(appErr.Fields) != 4 {
		t.Errorf("без обязательных полей: %v, ожидались 4 ошибки (sport, location, starts_on, capacity)", err)
	}
}

func TestService_PatchAndReplace(t *testing.T) {
	past := Competition{ID: 1, Title: "Прошедший турнир", Sport: "Шахматы", Location: "Норильск",
		StartsOn: "2026-01-10", Capacity: 8, Description: "Итоги опубликованы"}
	store := newFakeStore(past)
	svc := newTestService(store)
	ctx := context.Background()

	// PATCH меняет только переданное поле; прошедшая дата не мешает, если её не меняют
	patched, err := svc.Patch(ctx, 1, Input{Capacity: ptr(10)})
	if err != nil {
		t.Fatal(err)
	}
	if patched.Capacity != 10 || patched.Title != past.Title || patched.Description != past.Description {
		t.Errorf("PATCH изменил лишнее: %+v", patched)
	}

	// перенос на прошедшую дату запрещён
	_, err = svc.Patch(ctx, 1, Input{StartsOn: ptr("2026-02-01")})
	if appErr, ok := apperr.As(err); !ok || appErr.Fields[0].Field != "starts_on" {
		t.Errorf("перенос в прошлое: %v", err)
	}

	// PUT заменяет целиком: непереданное описание становится пустым
	replaced, err := svc.Replace(ctx, 1, Input{Title: ptr("Новый турнир"), Sport: ptr("Шахматы"),
		Location: ptr("Норильск"), StartsOn: ptr("2026-12-01"), Capacity: ptr(8)})
	if err != nil {
		t.Fatal(err)
	}
	if replaced.Description != "" {
		t.Errorf("PUT сохранил старое описание: %q", replaced.Description)
	}
}

func TestService_ErrorsMapping(t *testing.T) {
	c := validCompetition()
	c.ID = 1
	store := newFakeStore(c)
	store.approved[1] = 5
	svc := newTestService(store)
	ctx := context.Background()

	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{"чтение несуществующего", func() error { _, _, err := svc.Get(ctx, 404); return err }(), "NOT_FOUND"},
		{"удаление несуществующего", svc.Delete(ctx, 404), "NOT_FOUND"},
		{"лимит меньше одобренных", func() error { _, err := svc.Patch(ctx, 1, Input{Capacity: ptr(4)}); return err }(), "CAPACITY_BELOW_APPROVED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr, ok := apperr.As(tt.err)
			if !ok || appErr.Code != tt.wantCode {
				t.Errorf("ошибка %v, ожидался код %s", tt.err, tt.wantCode)
			}
		})
	}

	// неизвестные ошибки хранилища не превращаются в «понятные» клиенту
	if err := mapErr(errors.New("connection reset")); err == nil {
		t.Error("ошибка потеряна")
	} else if _, ok := apperr.As(err); ok {
		t.Errorf("внутренняя ошибка выдана как %v", err)
	}
}

func TestListKey_Normalized(t *testing.T) {
	a := listKey(3, Filter{Query: " Кубок ", Sport: "ФУТБОЛ"})
	b := listKey(3, Filter{Query: "кубок", Sport: "футбол"})
	if a != b {
		t.Errorf("ключи для одного запроса в разном регистре различаются: %q и %q", a, b)
	}
	if listKey(4, Filter{Query: "кубок"}) == listKey(3, Filter{Query: "кубок"}) {
		t.Error("смена версии не меняет ключ — инвалидация списков не сработает")
	}
	// разделитель «:» в тексте поиска не должен давать совпадение с другим фильтром
	if listKey(1, Filter{Query: "a:sport=b"}) == listKey(1, Filter{Query: "a", Sport: "b"}) {
		t.Error("разные фильтры дали один ключ")
	}
}
