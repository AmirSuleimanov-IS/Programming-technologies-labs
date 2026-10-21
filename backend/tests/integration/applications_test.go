//go:build integration

package integration

import (
	"net/http"
	"sync"
	"testing"
)

// Негативный сценарий: вторая активная заявка на то же соревнование отклоняется,
// а после отмены первой подать снова можно
func TestApplications_DuplicateAndReapply(t *testing.T) {
	reset(t)
	c := newCompetition(t, adminToken(t), "Турнир", 5)
	user := newUser(t, "anna")

	first := expect[applicationDTO](t, submit(t, user, c.ID), http.StatusCreated)
	if r := submit(t, user, c.ID); r.Status != http.StatusConflict || errorCode(t, r) != "DUPLICATE_APPLICATION" {
		t.Fatalf("повторная заявка: %d %s", r.Status, r.Body)
	}
	expect[any](t, setStatus(t, user, first.ID, "cancelled"), http.StatusOK)
	expect[any](t, submit(t, user, c.ID), http.StatusCreated)

	if n := countRows(t, `SELECT count(*) FROM applications WHERE status IN ('pending', 'approved')`); n != 1 {
		t.Errorf("активных заявок %d, ожидалась 1", n)
	}
}

func TestApplications_NoFreePlaces(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Шахматы", 1)
	app := expect[applicationDTO](t, submit(t, newUser(t, "anna"), c.ID), http.StatusCreated)
	expect[any](t, setStatus(t, admin, app.ID, "approved"), http.StatusOK)

	if r := submit(t, newUser(t, "boris"), c.ID); r.Status != http.StatusConflict || errorCode(t, r) != "NO_FREE_PLACES" {
		t.Fatalf("заявка без мест: %d %s", r.Status, r.Body)
	}
}

// Гонка за места: 6 заявок на 2 места одобряются одновременно. Блокировка строки
// соревнования (SELECT … FOR UPDATE) должна пропустить ровно 2 одобрения.
func TestApplications_ConcurrentApprovalsDoNotExceedCapacity(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Финал", 2)
	var ids []int64
	for _, name := range []string{"u1", "u2", "u3", "u4", "u5", "u6"} {
		ids = append(ids, expect[applicationDTO](t, submit(t, newUser(t, name), c.ID), http.StatusCreated).ID)
	}

	statuses := make([]int, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses[i] = setStatus(t, admin, id, "approved").Status
		}()
	}
	wg.Wait()

	ok, conflict := 0, 0
	for _, s := range statuses {
		switch s {
		case http.StatusOK:
			ok++
		case http.StatusConflict:
			conflict++
		}
	}
	if ok != 2 || conflict != 4 {
		t.Errorf("одобрено %d, отказов %d (коды %v), ожидалось 2 и 4", ok, conflict, statuses)
	}
	if n := countRows(t, `SELECT count(*) FROM applications WHERE status = 'approved'`); n != 2 {
		t.Errorf("в БД одобрено %d заявок при лимите 2", n)
	}
}
