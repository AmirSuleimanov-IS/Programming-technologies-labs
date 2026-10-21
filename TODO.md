# TODO — «Технологии программирования», лабораторные 1–4 (SportReg)

Легенда: [x] сделано, [ ] не сделано, [~] сделано частично (см. NOTES.md)

## 0. Подготовка

- [x] Задания и лекции из репозитория курса прочитаны
- [x] NOTES.md (дерево решений), CONTEXT.md (глоссарий), ADR 0001
- [x] Инструменты: pre-commit, golangci-lint, python-docx, Playwright + Edge, генератор отчётов
- [x] git init, .gitattributes (LF), корневой .gitignore

## Лаба 1 — первичная настройка (ветка `lab1-initial-setup`)

- [x] Бэкенд Go: CRUD соревнований (GET, POST, PUT, PATCH, DELETE), единый формат ответов и ошибок, PostgreSQL, миграции
- [x] Фронтенд React+TS+Vite: страница соревнований (список, создание, изменение, удаление)
- [x] Линтеры и форматтеры: golangci-lint + gofmt, ESLint + Prettier + knip, ignore-файлы
- [x] pre-commit: большие файлы, конфликты, линтер/форматтер, неиспользуемый код, имена файлов
- [x] .env development/production для обеих частей
- [x] Dockerfile ×2, docker-compose (db, backend, frontend)
- [x] README: корень, backend, frontend (структура, запуск, технологии со ссылками, env, соглашения, неочевидные решения)
- [x] Проверка: сборка, линтеры, docker compose up, ручной прогон API, демонстрация срабатывания каждого хука
- [x] Скриншоты и отчёт lab_1

## Лаба 2 — авторизация, роли, логи, мониторинг (ветка `lab2-auth-monitoring`)

- [x] Пользователи: регистрация, вход, bcrypt, JWT, администратор из env
- [x] Заявки: подача, отмена, одобрение, отклонение (бизнес-правила из NOTES Q16)
- [x] Роли ADMIN/USER: middleware на бэкенде, защищённые маршруты на фронтенде
- [x] Страницы: вход, регистрация, страница пользователя, страница администратора, 403
- [x] Логирование slog: уровни, обязательные события, production ≥ INFO
- [x] Prometheus /metrics: запросы, время, ошибки, авторизации
- [x] Grafana: provisioning источника и дашборда
- [x] docker-compose: + prometheus, grafana
- [x] Проверка, скриншоты, отчёт lab_2

## Лаба 3 — кэширование (ветка `lab3-caching`)

- [x] Redis в compose, модуль кэша с логами и метриками
- [x] Cache-aside для соревнования и списка, TTL с джиттером, инвалидация (delete on write + versioned keys), singleflight, fail-open
- [x] X-Cache и meta.cache в ответах
- [x] Кэш на фронтенде с TTL и инвалидацией, индикатор источника данных
- [x] Панели кэша в Grafana
- [x] Проверка (hit/miss/invalidate/TTL/отказ Redis), скриншоты, отчёт lab_3

## Лаба 4 — тестирование (ветка `lab4-testing`)

- [x] План тестового покрытия docs/test-plan.md
- [x] Unit-тесты: валидация, статусы заявок, пароль, JWT, права, кэш, TTL
- [x] Integration-тесты: testcontainers (PostgreSQL, Redis), позитивные и негативные, сквозной сценарий подачи заявки
- [x] Группы тестов и скрипт запуска, Vitest на фронтенде
- [x] CI GitHub Actions
- [x] Проверка, скриншоты, отчёт lab_4

## Финал

- [x] Полный прогон: линтеры, тесты, сборка, docker compose
- [x] Code review и исправления
- [x] Репозиторий на GitHub, push всех веток, проверка CI
- [x] Итоговый отчёт пользователю, обновить память
