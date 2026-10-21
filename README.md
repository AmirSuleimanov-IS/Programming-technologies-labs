# SportReg — лабораторные работы по дисциплине «Технологии программирования»

Студент: Сулейманов А.Ш., группа ИС-23, ЗГУ им. Н.М. Федоровского, 2026.
Задания: https://github.com/user-PSU-ISIT/Programming_technologies

**SportReg** — система записи на спортивные соревнования. Администратор ведёт каталог соревнований.

| Лабораторная                   | Ветка                | Отчёт                          |
| ------------------------------ | -------------------- | ------------------------------ |
| 1. Первичная настройка проекта | `lab1-initial-setup` | [reports/lab_1](reports/lab_1) |

`main` — итоговое состояние после последней лабораторной.

## Структура репозитория

```
.
├── backend/                 # REST API на Go (свой README, Dockerfile, линтер, .env)
├── frontend/                # SPA на React + TypeScript + Vite (свой README, Dockerfile, линтер, .env)
├── docs/adr/                # записи об архитектурных решениях
├── reports/lab_N/           # отчёт .docx, исходник отчёта spec.py, скриншоты img/
├── scripts/                 # обёртки для хуков pre-commit
├── tools/                   # генератор отчётов и скрипты скриншотов
├── docker-compose.yml       # локальный запуск всего проекта
├── .pre-commit-config.yaml  # проверки перед коммитом
├── CONTEXT.md               # глоссарий предметной области
├── NOTES.md, TODO.md        # принятые решения и план работ
└── .editorconfig, .gitattributes, .gitignore
```

Бэкенд и фронтенд не зависят друг от друга по коду: у каждой части свои инструкции, инструменты и настройки. Общие только `docker-compose.yml` и pre-commit. Почему один репозиторий, а не два — `docs/adr/0001-monorepo.md`.

## Быстрый запуск (Docker)

```bash
docker compose up -d --build
```

| Сервис           | Адрес                                |
| ---------------- | ------------------------------------ |
| Фронтенд (nginx) | http://localhost:8091                |
| API              | http://localhost:8090/api            |
| PostgreSQL       | localhost:5436 (sportreg / sportreg) |

Порты на хосте сдвинуты от стандартных (5432 → 5436 и т. д.), чтобы не конфликтовать с уже установленными сервисами. Пароль БД можно задать переменной `POSTGRES_PASSWORD` в файле `.env` рядом с `docker-compose.yml`.

Запуск для разработки без Docker — в `backend/README.md` и `frontend/README.md`.

## Pre-commit

```bash
pip install pre-commit      # один раз
pre-commit install          # подключить хук к git
pre-commit run --all-files  # проверить весь проект вручную
```

Коммит не выполнится, если:

| Проблема                                                                      | Хук                                                                                                                                    |
| ----------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| добавлен файл больше 500 КБ (отчёты в `reports/` — больше 5 МБ)               | `check-added-large-files`                                                                                                              |
| в файле остались маркеры конфликта `<<<<<<<`, `=======`, `>>>>>>>`            | `check-merge-conflict`                                                                                                                 |
| код не отформатирован                                                         | `gofmt`, `prettier`, `end-of-file-fixer`, `trailing-whitespace`, `mixed-line-ending`                                                   |
| код нарушает правила линтера                                                  | `golangci-lint`, `eslint`, `tsc`                                                                                                       |
| есть неиспользуемые импорты, переменные, функции, параметры, недостижимый код | Go-компилятор и `unused`/`unparam` в `golangci-lint`; `eslint` (`unused-imports`), `tsc` (`noUnused*`, `allowUnreachableCode`), `knip` |
| имя файла не по стандарту (Go — `snake_case`, фронтенд — `kebab-case`)        | `file-names-go`, `file-names-frontend`                                                                                                 |

Форматтеры и `eslint --fix` исправляют файлы сами, но коммит при этом всё равно останавливается. Разработчик смотрит правки (`git diff`), добавляет их (`git add`) и повторяет коммит.

## Соглашения

- Именование кода и файлов описано в `backend/README.md` и `frontend/README.md`.
- Коммиты атомарные, сообщение — `тип(где): что сделано`, например `feat(backend): REST API соревнований`. Типы: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`.
- На каждую лабораторную работу — своя ветка от предыдущей.
- Отступы, кодировка и переводы строк заданы в `.editorconfig`. В репозитории все текстовые файлы хранятся с LF (`.gitattributes`).

## Отчёты

Отчёт собирается из `reports/lab_N/spec.py` (нужны Python и Microsoft Word для обновления оглавления):

```bash
python -m venv tools/.venv && tools/.venv/Scripts/pip install python-docx pymupdf pillow
sh tools/report/make.sh 1
```
