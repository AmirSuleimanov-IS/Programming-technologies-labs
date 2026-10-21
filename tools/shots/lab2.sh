#!/usr/bin/env bash
# Рисунки отчёта 2: код, логи и вывод терминала → reports/lab_2/img
set -e
cd "$(dirname "$0")/.."
R=raw; O=../reports/lab_2/img; B=../backend; F=../frontend; M=../monitoring
cs() { node codeshot.mjs "$@" > /dev/null; }

# Аутентификация и роли
{ cat $B/internal/database/migrations/0002_users.sql; echo; cat $B/internal/database/migrations/0003_applications.sql; } |
  cs - $O/code_migrations.png --lang sql --title "migrations/0002_users.sql, 0003_applications.sql" --nonum
cs $B/internal/auth/validation.go $O/code_password.png --title "backend/internal/auth/validation.go" --from 58 --to 77
cs $B/internal/auth/tokens.go $O/code_tokens.png --title "backend/internal/auth/tokens.go" --from 30 --to 66
cs $B/internal/auth/service.go $O/code_login.png --title "backend/internal/auth/service.go" --from 70 --to 96
cs $B/internal/auth/guard.go $O/code_guard.png --title "backend/internal/auth/guard.go" --from 23 --to 62
cs $B/internal/competition/handler.go $O/code_routes_roles.png --title "backend/internal/competition/handler.go" --from 22 --to 32
cs $B/internal/application/rules.go $O/code_rules.png --title "backend/internal/application/rules.go" --from 22 --to 57
cs $B/internal/application/service.go $O/code_submit.png --title "backend/internal/application/service.go" --from 38 --to 67

# Фронтенд
cs $F/src/constants/routes.ts $O/code_fe_routes.png --lang typescript --title "frontend/src/constants/routes.ts" --width 900
cs $F/src/components/elements/protected-route.tsx $O/code_protected_route.png --lang typescript --title "frontend/src/components/elements/protected-route.tsx" --width 900
cs $F/src/app.tsx $O/code_app_routes.png --lang typescript --title "frontend/src/app.tsx" --from 13 --to 37
cs $F/src/api/http-client.ts $O/code_fe_token.png --lang typescript --title "frontend/src/api/http-client.ts" --from 19 --to 62

# Логирование и метрики
cs $B/internal/observability/logger.go $O/code_logger.png --title "backend/internal/observability/logger.go" --from 41 --to 71
cs $B/internal/observability/middleware.go $O/code_telemetry.png --title "backend/internal/observability/middleware.go" --from 30 --to 77
cs $B/internal/observability/metrics.go $O/code_metrics.png --title "backend/internal/observability/metrics.go"
cs $R/l2_logs_dev.txt $O/logs_dev.png --terminal --title "Логи в режиме development (LOG_LEVEL=debug)" --width 1500 --wrap
cs $R/l2_logs_prod.txt $O/logs_prod.png --terminal --title "docker compose logs backend (production)" --width 1500 --wrap
cs $R/l2_prod_loglevel.txt $O/logs_prod_level.png --terminal --title "production: LOG_LEVEL=debug поднимается до info" --width 1500 --wrap
cs $R/l2_metrics.txt $O/term_metrics.png --terminal --title "GET /metrics (фрагмент)" --width 1300

# Мониторинг
cs $M/prometheus/prometheus.yml $O/code_prometheus.png --lang yaml --title "monitoring/prometheus/prometheus.yml" --width 800
{ echo "# monitoring/grafana/provisioning/datasources/prometheus.yml"; cat $M/grafana/provisioning/datasources/prometheus.yml; echo; echo "# monitoring/grafana/provisioning/dashboards/sportreg.yml"; cat $M/grafana/provisioning/dashboards/sportreg.yml; } |
  cs - $O/code_grafana_provisioning.png --lang yaml --title "Provisioning Grafana" --nonum --width 900
sed -n '/^  prometheus:/,/^volumes:/p' ../docker-compose.yml | sed '$d' | cs - $O/code_compose_monitoring.png --lang yaml --title "docker-compose.yml (мониторинг)" --nonum --width 900
cs $R/l2_docker.txt $O/term_docker.png --terminal --title "docker compose ps" --width 1300

# Запросы к API
cs $R/l2_auth.txt $O/term_auth.png --terminal --title "Регистрация и вход" --width 1300
cs $R/l2_roles.txt $O/term_roles.png --terminal --title "Проверка ролей на бэкенде" --width 1300
cs $R/l2_apps.txt $O/term_apps.png --terminal --title "Заявки: статусная модель и лимит мест" --width 1300
echo "lab2 shots done"
