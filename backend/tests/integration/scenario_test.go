//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
)

// Smoke-тесты: основной «счастливый путь» работает — сервер, БД, вход, каталог.
// Быстрые, запускаются после сборки и развёртывания.
func TestSmoke_Health(t *testing.T) {
	expect[any](t, call(t, http.MethodGet, "/health", "", nil), http.StatusOK)
}

func TestSmoke_LoginAndCatalog(t *testing.T) {
	reset(t)
	expect[[]competitionDTO](t, call(t, http.MethodGet, "/competitions", adminToken(t), nil), http.StatusOK)
}

// Законченный пользовательский сценарий «успешная заявка»: регистрация → вход → каталог →
// подача заявки → администратор видит её и одобряет → пользователь видит одобрение,
// свободных мест становится меньше.
func TestScenario_SuccessfulApplication(t *testing.T) {
	reset(t)
	admin := adminToken(t)
	competition := newCompetition(t, admin, "Первенство ЗГУ по волейболу", 12)

	// 1. регистрация и вход пользователя
	expect[any](t, call(t, http.MethodPost, "/auth/register", "",
		map[string]string{"email": "maria@test.local", "full_name": "Мария Кузнецова", "password": "Volley2026"}), http.StatusCreated)
	user := login(t, "maria@test.local", "Volley2026")

	// 2. каталог: соревнование видно, 12 свободных мест
	catalog := expect[[]competitionDTO](t, call(t, http.MethodGet, "/competitions", user, nil), http.StatusOK)
	if len(catalog) != 1 || catalog[0].FreePlaces != 12 {
		t.Fatalf("каталог: %+v", catalog)
	}

	// 3. подача заявки — статус «на рассмотрении»
	app := expect[applicationDTO](t, submit(t, user, competition.ID), http.StatusCreated)
	if app.Status != "pending" {
		t.Fatalf("статус новой заявки %q", app.Status)
	}

	// 4. администратор видит заявку среди заявок на рассмотрении и одобряет её
	pending := expect[[]applicationDTO](t, call(t, http.MethodGet, "/applications?status=pending", admin, nil), http.StatusOK)
	if len(pending) != 1 || pending[0].ID != app.ID {
		t.Fatalf("заявки на рассмотрении: %+v", pending)
	}
	expect[any](t, setStatus(t, admin, app.ID, "approved"), http.StatusOK)

	// 5. пользователь видит одобрение, свободных мест стало 11
	mine := expect[[]applicationDTO](t, call(t, http.MethodGet, "/applications/my", user, nil), http.StatusOK)
	if len(mine) != 1 || mine[0].Status != "approved" {
		t.Errorf("мои заявки: %+v", mine)
	}
	got := expect[competitionDTO](t, call(t, http.MethodGet, fmt.Sprintf("/competitions/%d", competition.ID), user, nil), http.StatusOK)
	if got.FreePlaces != 11 {
		t.Errorf("свободных мест %d, ожидалось 11", got.FreePlaces)
	}
}
