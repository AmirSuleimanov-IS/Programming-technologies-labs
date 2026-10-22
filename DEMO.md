# Как показать лабораторные работы (в том числе после перезапуска компьютера)

После перезапуска контейнеры сами не стартуют: нужно запустить Docker Desktop и проект. Это делает один скрипт.

## Запуск (2–3 минуты)

1. Откройте PowerShell в папке `D:\Projects\technologies` (в Проводнике: адресная строка → `powershell` → Enter).
2. Выполните:

   ```powershell
   powershell -ExecutionPolicy Bypass -File demo.ps1 -Open
   ```

   Скрипт запускает Docker Desktop (если он не запущен) и ждёт его, поднимает все контейнеры, ждёт готовности API, создаёт демо-пользователей и открывает в браузере приложение, Grafana и Prometheus. Перед показом дашборда Grafana (лабы 2 и 3) добавьте ключ `-Traffic`: 2 минуты будет идти нагрузка, и на графиках появятся данные.

| Что        | Адрес                            | Вход          |
| ---------- | -------------------------------- | ------------- |
| Приложение | http://localhost:8091            | см. ниже      |
| API        | http://localhost:8090/api/health | —             |
| Метрики    | http://localhost:8090/metrics    | —             |
| Prometheus | http://localhost:9091/targets    | —             |
| Grafana    | http://localhost:3001            | admin / admin |

| Роль  | Email                    | Пароль     |
| ----- | ------------------------ | ---------- |
| ADMIN | admin@sportreg.local     | Admin12345 |
| USER  | ivan.petrov@mail.ru      | Sport2026  |
| USER  | olga.smirnova@mail.ru    | Chess2026  |
| USER  | maria.kuznetsova@mail.ru | Volley2026 |

Все команды ниже выполняются в PowerShell из папки `D:\Projects\technologies`, если не сказано иное. Скрипты `.sh` запускаются через Git Bash: `& "C:\Program Files\Git\bin\bash.exe" путь/к/скрипту.sh`. Просто `bash` в PowerShell запускает WSL, а там нет Go и Node.js.

Сценарии `tools/steps/demo_*.json` можно запускать сколько угодно раз: каждый сам создаёт нужное соревнование и удаляет его в конце.

## Лабораторная 1 — первичная настройка проекта

1. **Структура.** `backend/` и `frontend/` раздельны, в каждой свой README, `.gitignore`, `Dockerfile`, линтер, `.env`. Общие только `docker-compose.yml` и `.pre-commit-config.yaml`. Удобно показать на GitHub: https://github.com/AmirSuleimanov-IS/Programming-technologies-labs
2. **Линтеры и форматтеры:**

   ```powershell
   cd backend; golangci-lint run ./...; cd ..
   cd frontend; npm run lint; npm run format:check; npm run typecheck; npm run knip; cd ..
   ```

3. **Pre-commit по всем файлам:** `tools\.venv\Scripts\pre-commit run --all-files` — все проверки Passed.
4. **Коммит не проходит при нарушениях:**

   ```powershell
   & "C:\Program Files\Git\bin\bash.exe" tools/precommit-demo.sh
   ```

   На временной ветке по очереди делаются коммиты: файл 684 КБ, маркеры конфликта, неотформатированный код и неиспользуемая функция, имена `CompetitionCard.tsx` и `competitionCache.go`. Каждый коммит останавливается. Последний случай: Prettier сам исправил файл и остановил коммит, разработчик смотрит `git diff`, добавляет файл, и повторный коммит проходит. Между случаями нажимайте Enter. В конце скрипт удаляет временную ветку. Перед запуском в рабочей копии не должно быть незакоммиченных изменений.

5. **REST API:** `node tools/api-demo.mjs tools/steps/demo_lab1.json` — GET, POST (201 + Location), PATCH, PUT, DELETE (204), затем ошибки 404, 422, 400, 405.
6. **Клиент:** http://localhost:8091 → войти администратором → «Администрирование» → вкладка «Соревнования»: «Добавить соревнование» с пустыми полями (ошибки под полями), затем добавить, изменить, удалить.
7. **Переменные окружения:** `backend/.env.development`, `backend/.env.production.example`, `frontend/.env.development`, `frontend/.env.production`. В Docker (production) в шапке написано «SportReg». Режим разработки: `cd frontend; npm run dev -- --host 127.0.0.1`, затем откройте http://127.0.0.1:5173 — там «SportReg (dev)» (Ctrl+C — остановить).
8. **Docker:** `docker compose ps`, файлы `backend/Dockerfile`, `frontend/Dockerfile`, `docker-compose.yml`.

## Лабораторная 2 — авторизация, роли, логирование, мониторинг

1. **API и роли:**

   ```powershell
   node tools/api-demo.mjs tools/steps/demo_lab2.json       # 401, 409, 422, 401, 200, 403, 403, 401, 200
   node tools/api-demo.mjs tools/steps/demo_lab2_apps.json  # заявки: 201, 409, 403, 200, 409 (мест нет), отмена, удаление
   ```

2. **Интерфейс:** войти как `ivan.petrov@mail.ru` — в шапке только «Соревнования». Подать заявку, затем вручную открыть http://localhost:8091/admin — страница 403. Выйти и войти администратором: видны обе страницы, во вкладке «Заявки» одобрить заявку Ивана. Чтобы показать обе роли одновременно, откройте второе окно в режиме InPrivate (Ctrl+Shift+N).
3. **Логи:** `docker compose logs backend --tail 30` — JSON со временем, уровнем, методом, путём, статусом, `request_id`, пользователем; входы, отказы 401/403 — WARN.
4. **В production уровень не ниже INFO:**

   ```powershell
   docker compose run -d --no-deps --name loglevel-demo -e LOG_LEVEL=debug backend
   Start-Sleep 3; docker logs loglevel-demo; docker rm -f loglevel-demo
   ```

   В выводе: предупреждение, что DEBUG в production не разрешён, и `"log_level":"INFO"`.

5. **Метрики:** http://localhost:8090/metrics (Ctrl+F: `http_requests_total`, `auth_attempts_total`).
6. **Prometheus:** http://localhost:9091/targets — цель `sportreg-backend` в состоянии UP.
7. **Grafana:** http://localhost:3001 (admin / admin) → Dashboards → SportReg → «SportReg — API»: запросы, ошибки 4xx/5xx, время ответа, успешные и неуспешные входы. Если графики пустые, запустите нагрузку `node tools/traffic.mjs 120` и подождите минуту.

## Лабораторная 3 — кэширование

1. **Интерфейс:** войти как Иван. Над каталогом — «Источник данных». Первая загрузка — сервер (MISS или HIT). Введите что-нибудь в поиск и сотрите — «кэш браузера, сохранено N с назад». «Обновить с сервера» — «сервер, кэш Redis (HIT)». F12 → «Консоль»: строки `[cache] miss/set/hit/invalidate`.
2. **API:** `node tools/api-demo.mjs tools/steps/demo_lab3.json` — MISS → HIT → изменение → MISS с новыми данными → HIT, у списка — смена версии и нормализация регистра (заголовок `X-Cache` и поле `meta.cache`).
3. **Redis:**

   ```powershell
   docker compose exec redis redis-cli --scan
   docker compose exec redis redis-cli TTL competition:1     # после открытия каталога; около 300 с ± 10%
   docker compose exec redis redis-cli CONFIG GET maxmemory-policy
   ```

4. **Логи кэша:** `docker compose logs backend --tail 100 | Select-String '"cache '` — hit, miss, set, invalidate с ключом и результатом.
5. **Отказ Redis — сервис продолжает работать:**

   ```powershell
   docker compose stop redis
   node tools/api-demo.mjs tools/steps/demo_lab3.json      # X-Cache: BYPASS, ответы 200
   docker compose start redis
   ```

6. **Grafana:** внизу дашборда — «Доля попаданий в кэш» и «Операции с кэшем по результату».

## Лабораторная 4 — тестирование

Для integration-тестов нужен запущенный Docker: тесты сами поднимают отдельные PostgreSQL и Redis и удаляют их после себя.

```powershell
cd backend
go test ./internal/...                                            # unit-тесты, секунды
go test -v ./internal/auth -run TestCheckPassword                 # граничные значения пароля
go test -v ./internal/application -run "TestCheckTransition$"     # таблица решений для статусов
go test -tags=integration ./tests/integration/...                 # integration, ~20 с
go test -tags=integration -v -run "^TestSmoke_" ./tests/integration/...
go test -tags=integration -v -run "^TestScenario_" ./tests/integration/...
go test -cover ./internal/...                                     # покрытие
cd ..\frontend; npm test; cd ..
```

- Группы одной командой: `& "C:\Program Files\Git\bin\bash.exe" backend/scripts/test.sh smoke` (также `unit`, `integration`, `scenario`, `regression`, `validation`, `access`, `auth`, `competitions`, `applications`, `cache`).
- Пока идут integration-тесты, во втором окне `docker ps`: тестовые `postgres:17-alpine` и `redis:7.4-alpine` на случайных портах, отдельно от `sportreg-db-1`. Основная БД не используется.
- План покрытия — `docs/test-plan.md`. CI — вкладка Actions на GitHub, все запуски зелёные.
- Мутационная проверка вживую: в `backend/internal/auth/validation.go` замените `MinPasswordLen   = 8` на `7`, запустите `go test ./internal/auth` — тест упадёт на границе 7 символов. Верните файл: `git checkout -- backend/internal/auth/validation.go`.

## После показа

`docker compose stop` — остановить контейнеры (данные сохраняются). `docker compose down -v` — удалить и данные. При следующем запуске база создастся заново с демо-соревнованиями, а `demo.ps1` снова создаст пользователей.

## Если что-то не так

| Симптом                                       | Что делать                                                                                                                                                     |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `port is already allocated` при запуске       | порт занят другим контейнером: `docker ps`, остановите лишний (`docker stop <имя>`). Процессы `com.docker.backend` и `wslrelay` не завершайте — это сам Docker |
| Docker Desktop не запускается                 | откройте его вручную и подождите. Если не помогло: закройте Docker Desktop, выполните `wsl --shutdown`, запустите снова                                        |
| страница не открывается                       | `docker compose ps` — все контейнеры должны быть Up; `docker compose logs backend --tail 20`                                                                   |
| не входит пользователь Иван (после `down -v`) | `node tools/api-demo.mjs tools/steps/demo_users.json`                                                                                                          |
| в Grafana пустые графики                      | `node tools/traffic.mjs 120`, подождать минуту; период справа сверху — «Last 15 minutes»                                                                       |
| `npm run dev` не открывается по localhost     | открывайте http://127.0.0.1:5173 (на этом компьютере IPv6-адрес `::1` заблокирован)                                                                            |
