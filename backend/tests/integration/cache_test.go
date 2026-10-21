//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestCache_HitAfterMiss(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Кубок", 10)
	path := fmt.Sprintf("/competitions/%d", c.ID)

	for i, want := range []string{"MISS", "HIT"} {
		if got := call(t, http.MethodGet, path, admin, nil).Header.Get("X-Cache"); got != want {
			t.Errorf("запрос %d: X-Cache %q, ожидался %q", i+1, got, want)
		}
	}
	ttl := rdb.TTL(context.Background(), fmt.Sprintf("competition:%d", c.ID)).Val()
	if ttl <= 0 {
		t.Errorf("у ключа нет TTL: %v", ttl)
	}
}

// После изменения соревнования кэш отдаёт новые данные, а не устаревшие
func TestCache_InvalidatedOnUpdate(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Кубок", 10)
	path := fmt.Sprintf("/competitions/%d", c.ID)
	call(t, http.MethodGet, path, admin, nil)
	call(t, http.MethodGet, "/competitions", admin, nil) // список тоже в кэше

	expect[any](t, call(t, http.MethodPatch, path, admin, map[string]string{"title": "Кубок ЗГУ"}), http.StatusOK)

	r := call(t, http.MethodGet, path, admin, nil)
	if got := expect[competitionDTO](t, r, http.StatusOK); got.Title != "Кубок ЗГУ" || r.Header.Get("X-Cache") != "MISS" {
		t.Errorf("после изменения: %q, X-Cache %s", got.Title, r.Header.Get("X-Cache"))
	}
	list := expect[[]competitionDTO](t, call(t, http.MethodGet, "/competitions", admin, nil), http.StatusOK)
	if len(list) != 1 || list[0].Title != "Кубок ЗГУ" {
		t.Errorf("список из кэша устарел: %+v", list)
	}
}

// Одобрение заявки меняет свободные места — закэшированное соревнование должно обновиться
func TestCache_InvalidatedOnApproval(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Кубок", 10)
	path := fmt.Sprintf("/competitions/%d", c.ID)
	call(t, http.MethodGet, path, admin, nil)

	app := expect[applicationDTO](t, submit(t, newUser(t, "anna"), c.ID), http.StatusCreated)
	expect[any](t, setStatus(t, admin, app.ID, "approved"), http.StatusOK)

	if got := expect[competitionDTO](t, call(t, http.MethodGet, path, admin, nil), http.StatusOK); got.FreePlaces != 9 {
		t.Errorf("свободных мест %d, ожидалось 9 — кэш не сброшен после одобрения", got.FreePlaces)
	}
}
