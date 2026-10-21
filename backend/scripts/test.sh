#!/usr/bin/env bash
# Запуск тестов по группам. Использование: scripts/test.sh <группа>
#
#   unit          быстрые тесты функций и сервисов без БД (секунды) — при каждом изменении
#   integration   API + PostgreSQL + Redis в контейнерах (нужен Docker) — перед слиянием
#   smoke         основной счастливый путь: здоровье сервиса, вход, каталог — после сборки
#   scenario      законченный пользовательский сценарий «успешная заявка»
#   regression    всё вместе — перед сдачей и релизом
#   validation    проверки входных данных (поля, пароль, JSON, параметры пути)
#   access        права доступа (роли, токены, статусная модель)
#   auth | competitions | applications | cache   тесты одной сущности (unit + integration)
#   coverage      unit-тесты с отчётом о покрытии по пакетам
set -euo pipefail
cd "$(dirname "$0")/.."

it() { go test -tags=integration -count=1 ./tests/integration/... "$@"; }

case "${1:-}" in
  unit)         go test ./internal/... ;;
  integration)  it ;;
  smoke)        it -run '^TestSmoke_' -v ;;
  scenario)     it -run '^TestScenario_' -v ;;
  regression)   go test -count=1 ./internal/... && it ;;
  validation)   go test ./internal/... -run 'Validate|CheckPassword|DecodeJSON|PathID' -v ;;
  access)       go test ./internal/... -run 'Guard|Tokens|Transition' -v && it -run 'AccessMatrix' -v ;;
  auth)         go test ./internal/auth/... && it -run '^TestAuth_' ;;
  competitions) go test ./internal/competition/... && it -run '^TestCompetitions_' ;;
  applications) go test ./internal/application/... && it -run '^TestApplications_|^TestScenario_' ;;
  cache)        go test ./internal/cache/... && it -run '^TestCache_' ;;
  coverage)     go test -cover ./internal/... ;;
  *)
    sed -n '2,15p' "$0"
    exit 2
    ;;
esac
