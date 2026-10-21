#!/usr/bin/env bash
# Рисунки отчёта 3: код кэша и вывод терминала → reports/lab_3/img
set -e
cd "$(dirname "$0")/.."
R=raw; O=../reports/lab_3/img; B=../backend; F=../frontend
cs() { node codeshot.mjs "$@" > /dev/null; }

# Бэкенд
cs $B/internal/cache/cache.go $O/code_cache_load.png --title "backend/internal/cache/cache.go" --from 73 --to 116
cs $B/internal/cache/cache.go $O/code_cache_breaker.png --title "backend/internal/cache/cache.go" --from 29 --to 72
cs $B/internal/cache/cache.go $O/code_cache_invalidate.png --title "backend/internal/cache/cache.go" --from 130 --to 176
cs $B/internal/competition/cache.go $O/code_competition_cache.png --title "backend/internal/competition/cache.go"
cs $B/internal/application/service.go $O/code_app_invalidate.png --title "backend/internal/application/service.go" --from 98 --to 113
cs $B/internal/competition/handler.go $O/code_with_source.png --title "backend/internal/competition/handler.go" --from 115 --to 128
cs $B/cmd/api/main.go $O/code_new_cache.png --title "backend/cmd/api/main.go" --from 115 --to 136
sed -n '/^  # Кэш/,/^  backend:/p' ../docker-compose.yml | sed '$d' | cs - $O/code_compose_redis.png --lang yaml --title "docker-compose.yml (Redis)" --nonum --width 1300

# Фронтенд
cs $F/src/utils/query-cache.ts $O/code_query_cache.png --lang typescript --title "frontend/src/utils/query-cache.ts" --from 1 --to 56
cs $F/src/api/competitions-api.ts $O/code_fe_fetch.png --lang typescript --title "frontend/src/api/competitions-api.ts" --from 1 --to 49
cs $F/src/components/ui/source-badge.tsx $O/code_source_badge.png --lang typescript --title "frontend/src/components/ui/source-badge.tsx"

# Терминал
cs $R/l3_cache.txt $O/term_cache.png --terminal --title "Попадания, промахи и инвалидация кэша" --width 1150
cs $R/l3_redis.txt $O/term_redis.png --terminal --title "Содержимое Redis" --width 1150
cs $R/l3_ttl.txt $O/term_ttl.png --terminal --title "Истечение TTL (CACHE_ITEM_TTL=5s)" --width 1000
cs $R/l3_bypass.txt $O/term_bypass.png --terminal --title "Отказ Redis: fail-open" --width 1150
cs $R/l3_bench.txt $O/term_bench.png --terminal --title "Время ответа: Redis и БД" --width 1000
cs $R/l3_logs.txt $O/logs_cache.png --terminal --title "Логи операций с кэшем" --width 1500 --wrap
cs $R/l3_console.txt $O/console_cache.png --terminal --title "Консоль браузера: журнал кэша клиента" --width 1000
echo "lab3 shots done"
