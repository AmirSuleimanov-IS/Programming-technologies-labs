#!/usr/bin/env bash
# Сбрасывает демо-данные в БД из docker-compose: пустые таблицы, счётчики id с 1.
# После этого бэкенд при старте снова создаёт администратора и демо-соревнования.
set -euo pipefail
cd "$(dirname "$0")/.."
docker compose exec -T db psql -U sportreg -d sportreg -q -c \
  "TRUNCATE competitions, users, applications RESTART IDENTITY CASCADE"
echo "demo data cleared"
