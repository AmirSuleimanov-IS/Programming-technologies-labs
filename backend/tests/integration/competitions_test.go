//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// Позитивный сценарий: создание, чтение, PATCH, PUT и удаление сохраняются в БД
func TestCompetitions_CRUD(t *testing.T) {
	reset(t)
	admin := adminToken(t)

	created := newCompetition(t, admin, "Кубок ЗГУ", 12)
	if created.ID == 0 || created.FreePlaces != 12 {
		t.Fatalf("создано %+v", created)
	}
	path := fmt.Sprintf("/competitions/%d", created.ID)

	patched := expect[competitionDTO](t, call(t, http.MethodPatch, path, admin, map[string]int{"capacity": 20}), http.StatusOK)
	if patched.Capacity != 20 || patched.Title != "Кубок ЗГУ" {
		t.Errorf("PATCH: %+v", patched)
	}
	expect[any](t, call(t, http.MethodPut, path, admin, map[string]any{
		"title": "Кубок ЗГУ 2026", "sport": "Волейбол", "location": "Норильск",
		"starts_on": time.Now().AddDate(0, 2, 0).Format("2006-01-02"), "capacity": 24,
	}), http.StatusOK)
	if n := countRows(t, `SELECT count(*) FROM competitions WHERE title = 'Кубок ЗГУ 2026' AND capacity = 24`); n != 1 {
		t.Errorf("PUT не сохранён в БД")
	}

	expect[any](t, call(t, http.MethodDelete, path, admin, nil), http.StatusNoContent)
	if r := call(t, http.MethodGet, path, admin, nil); r.Status != http.StatusNotFound {
		t.Errorf("после удаления: %d", r.Status)
	}
	if n := countRows(t, `SELECT count(*) FROM competitions`); n != 0 {
		t.Errorf("в БД осталось %d соревнований", n)
	}
}

// Негативный сценарий: невалидные данные не попадают в БД
func TestCompetitions_InvalidNotSaved(t *testing.T) {
	reset(t)
	r := call(t, http.MethodPost, "/competitions", adminToken(t), map[string]any{
		"title": "Ку", "sport": "Волейбол", "location": "Норильск", "starts_on": "2020-01-01", "capacity": 0,
	})
	if r.Status != http.StatusUnprocessableEntity || errorCode(t, r) != "VALIDATION_ERROR" {
		t.Fatalf("ответ %d %s", r.Status, r.Body)
	}
	if n := countRows(t, `SELECT count(*) FROM competitions`); n != 0 {
		t.Errorf("невалидное соревнование сохранено")
	}
}

// Негативный сценарий: лимит нельзя сделать меньше числа одобренных заявок
func TestCompetitions_CapacityBelowApproved(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Турнир", 2)
	for _, name := range []string{"anna", "boris"} {
		app := expect[applicationDTO](t, submit(t, newUser(t, name), c.ID), http.StatusCreated)
		expect[any](t, setStatus(t, admin, app.ID, "approved"), http.StatusOK)
	}

	r := call(t, http.MethodPatch, fmt.Sprintf("/competitions/%d", c.ID), admin, map[string]int{"capacity": 1})
	if r.Status != http.StatusConflict || errorCode(t, r) != "CAPACITY_BELOW_APPROVED" {
		t.Fatalf("ответ %d %s", r.Status, r.Body)
	}
	if n := countRows(t, `SELECT capacity FROM competitions WHERE id = $1`, c.ID); n != 2 {
		t.Errorf("лимит в БД изменился: %d", n)
	}
}

// Негативный сценарий: соревнование с активными заявками не удаляется; после отмены — удаляется
// вместе с отменёнными заявками
func TestCompetitions_DeleteWithActiveApplications(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	c := newCompetition(t, admin, "Турнир", 5)
	user := newUser(t, "anna")
	app := expect[applicationDTO](t, submit(t, user, c.ID), http.StatusCreated)
	path := fmt.Sprintf("/competitions/%d", c.ID)

	r := call(t, http.MethodDelete, path, admin, nil)
	if r.Status != http.StatusConflict || errorCode(t, r) != "HAS_ACTIVE_APPLICATIONS" {
		t.Fatalf("ответ %d %s", r.Status, r.Body)
	}
	if n := countRows(t, `SELECT count(*) FROM competitions`); n != 1 {
		t.Fatal("соревнование удалено вместе с активной заявкой")
	}

	expect[any](t, setStatus(t, user, app.ID, "cancelled"), http.StatusOK)
	expect[any](t, call(t, http.MethodDelete, path, admin, nil), http.StatusNoContent)
	if n := countRows(t, `SELECT count(*) FROM applications`); n != 0 {
		t.Errorf("отменённые заявки не удалились: %d", n)
	}
}
