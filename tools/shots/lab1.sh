#!/usr/bin/env bash
# Рисунки отчёта 1: код и вывод терминала → reports/lab_1/img
set -e
cd "$(dirname "$0")/.."
R=raw; O=../reports/lab_1/img; B=../backend; F=../frontend
cs() { node codeshot.mjs "$@" > /dev/null; }
term() { grep -v "Skipped" "$1" | cs - "$2" --terminal --title "${3:-Терминал}" --width "${4:-1100}"; }

# Структура
cs $R/l1_tree_root.txt $O/tree_root.png --terminal --title "Структура репозитория" --width 600
cs $R/l1_tree_backend.txt $O/tree_backend.png --terminal --title "backend/" --width 600
cs $R/l1_tree_frontend.txt $O/tree_frontend.png --terminal --title "frontend/" --width 600

# Бэкенд
cs $B/internal/competition/handler.go $O/code_routes.png --title "backend/internal/competition/handler.go" --from 20 --to 28
cs $B/internal/httpapi/response.go $O/code_response.png --title "backend/internal/httpapi/response.go" --from 15 --to 67
cs $B/internal/competition/service.go $O/code_service.png --title "backend/internal/competition/service.go" --from 40 --to 88
cs $B/internal/competition/validation.go $O/code_validation.png --title "backend/internal/competition/validation.go" --from 58 --to 87
cs $B/internal/httpapi/middleware.go $O/code_json_errors.png --title "backend/internal/httpapi/middleware.go" --from 66 --to 98
cs $B/internal/config/config.go $O/code_config.png --title "backend/internal/config/config.go" --from 27 --to 51

# Окружение
{ echo "# backend/.env.development"; cat $B/.env.development; echo; echo "# backend/.env.production.example"; cat $B/.env.production.example; echo; echo "# frontend/.env.development"; cat $F/.env.development; echo; echo "# frontend/.env.production"; cat $F/.env.production; } | cs - $O/code_env.png --lang ini --title ".env-файлы" --nonum

# Фронтенд
cs $F/src/api/http-client.ts $O/code_http_client.png --lang typescript --title "frontend/src/api/http-client.ts" --from 19 --to 48
cs $F/src/pages/competitions-page.tsx $O/code_patch.png --lang typescript --title "frontend/src/pages/competitions-page.tsx" --from 9 --to 35

# Инструменты качества
cs $B/.golangci.yml $O/code_golangci.png --lang yaml --title "backend/.golangci.yml"
cs $F/eslint.config.js $O/code_eslint.png --lang javascript --title "frontend/eslint.config.js" --from 10 --to 47
{ echo "// .prettierrc.json"; cat $F/.prettierrc.json; } | cs - $O/code_prettier.png --lang json --title "frontend/.prettierrc.json, .prettierignore" --nonum --width 700
cs ../.pre-commit-config.yaml $O/code_precommit1.png --lang yaml --title ".pre-commit-config.yaml" --from 1 --to 37
cs ../.pre-commit-config.yaml $O/code_precommit2.png --lang yaml --title ".pre-commit-config.yaml" --from 38 --to 94
cs $R/l1_linters.txt $O/term_linters.png --terminal --title "Линтеры и форматтеры" --width 900

# Docker
cs $B/Dockerfile $O/code_dockerfile_backend.png --lang dockerfile --title "backend/Dockerfile" --width 800
cs $F/Dockerfile $O/code_dockerfile_frontend.png --lang dockerfile --title "frontend/Dockerfile" --width 800
cs ../docker-compose.yml $O/code_compose.png --lang yaml --title "docker-compose.yml" --width 1000
cs $R/l1_docker.txt $O/term_docker.png --terminal --title "Запуск через docker compose"

# API
head -64 $R/l1_crud.txt | cs - $O/term_crud1.png --terminal --title "Запросы к API: GET, POST, PATCH" --width 1300
sed -n '65,200p' $R/l1_crud.txt | cs - $O/term_crud2.png --terminal --title "Запросы к API: PUT, DELETE" --width 1300
cs $R/l1_errors.txt $O/term_errors.png --terminal --title "Обработка ошибок" --width 1100

# pre-commit
term $R/l1_pc_large.txt $O/pc_large.png "pre-commit: большой файл"
term $R/l1_pc_conflict.txt $O/pc_conflict.png "pre-commit: маркеры конфликта"
term $R/l1_pc_format.txt $O/pc_format.png "pre-commit: форматирование и мёртвый код"
term $R/l1_pc_unused.txt $O/pc_unused.png "pre-commit: неиспользуемые импорты и переменные, недостижимый код" 1300
cs $R/l1_pc_unused_diff.txt $O/pc_unused_diff.png --terminal --title "git diff — что исправил eslint --fix и prettier" --width 900
term $R/l1_pc_names.txt $O/pc_names.png "pre-commit: имена файлов"
term $R/l1_pc_cycle1.txt $O/pc_cycle1.png "pre-commit: коммит остановлен, файл исправлен"
cs $R/l1_pc_cycle_diff.txt $O/pc_cycle_diff.png --terminal --title "Разработчик проверяет исправление" --width 1000
term $R/l1_pc_cycle2.txt $O/pc_cycle2.png "Повторный коммит"
echo "lab1 shots done"
