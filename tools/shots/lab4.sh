#!/usr/bin/env bash
# Рисунки отчёта 4: код тестов и вывод терминала → reports/lab_4/img
set -e
cd "$(dirname "$0")/.."
R=raw; O=../reports/lab_4/img; B=../backend; F=../frontend
cs() { node codeshot.mjs "$@" > /dev/null; }

# Unit-тесты бэкенда
cs $B/internal/auth/validation_test.go $O/code_test_password.png --title "backend/internal/auth/validation_test.go" --from 8 --to 41
cs $B/internal/application/rules_test.go $O/code_test_transition.png --title "backend/internal/application/rules_test.go" --from 10 --to 61
cs $B/internal/application/rules_test.go $O/code_test_submit.png --title "backend/internal/application/rules_test.go" --from 72 --to 98
cs $B/internal/auth/tokens_test.go $O/code_test_tokens.png --title "backend/internal/auth/tokens_test.go" --from 51 --to 87
cs $B/internal/auth/guard_test.go $O/code_test_guard.png --title "backend/internal/auth/guard_test.go" --from 12 --to 44
cs $B/internal/application/service_test.go $O/code_test_fake_store.png --title "backend/internal/application/service_test.go" --from 11 --to 40
cs $B/internal/application/service_test.go $O/code_test_invalidate.png --title "backend/internal/application/service_test.go" --from 106 --to 131
cs $B/internal/cache/cache_test.go $O/code_test_cache.png --title "backend/internal/cache/cache_test.go" --from 86 --to 135

# Integration-тесты
cs $B/tests/integration/main_test.go $O/code_it_main.png --title "backend/tests/integration/main_test.go" --from 45 --to 76
cs $B/tests/integration/helpers_test.go $O/code_it_reset.png --title "backend/tests/integration/helpers_test.go" --from 14 --to 29
cs $B/tests/integration/competitions_test.go $O/code_it_negative.png --title "backend/tests/integration/competitions_test.go" --from 44 --to 76
cs $B/tests/integration/applications_test.go $O/code_it_concurrent.png --title "backend/tests/integration/applications_test.go" --from 42 --to 80
cs $B/tests/integration/scenario_test.go $O/code_it_scenario.png --title "backend/tests/integration/scenario_test.go" --from 22 --to 67
cs $B/tests/integration/auth_test.go $O/code_it_access.png --title "backend/tests/integration/auth_test.go" --from 40 --to 77

# Фронтенд
cs $F/src/utils/query-cache.test.ts $O/code_vitest_cache.png --lang typescript --title "frontend/src/utils/query-cache.test.ts" --from 1 --to 34
cs $F/src/api/competitions-api.test.ts $O/code_vitest_api.png --lang typescript --title "frontend/src/api/competitions-api.test.ts" --from 1 --to 39

# Группы, CI
cs $B/scripts/test.sh $O/code_test_sh.png --lang bash --title "backend/scripts/test.sh"
cs ../.github/workflows/ci.yml $O/code_ci.png --lang yaml --title ".github/workflows/ci.yml"

# Вывод
cs $R/l4_unit.txt $O/term_unit.png --terminal --title "Unit-тесты бэкенда" --width 800
cs $R/l4_password.txt $O/term_password.png --terminal --title "Граничные значения пароля" --width 1000
cs $R/l4_transition.txt $O/term_transition.png --terminal --title "Таблица решений: смена статуса заявки" --width 900
cs $R/l4_integration.txt $O/term_integration.png --terminal --title "Integration-тесты" --width 1000
cs $R/l4_docker_ps.txt $O/term_docker_ps.png --terminal --title "Контейнеры во время integration-тестов" --width 1300
cs $R/l4_groups.txt $O/term_groups.png --terminal --title "Запуск по группам" --width 900
cs $R/l4_coverage.txt $O/term_coverage.png --terminal --title "Покрытие кода" --width 1300
cs $R/l4_vitest.txt $O/term_vitest.png --terminal --title "Vitest" --width 1500
cs $R/l4_mutations.txt $O/term_mutations.png --terminal --title "Мутационная проверка unit-тестов" --width 1300
cs $R/l4_mutation_lock.txt $O/term_mutation_lock.png --terminal --title "Мутационная проверка integration-теста" --width 1200
echo "lab4 shots done"
