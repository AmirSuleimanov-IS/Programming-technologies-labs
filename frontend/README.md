# SportReg — frontend

Клиент системы записи на спортивные соревнования: SPA на React. Страница «Соревнования» показывает каталог с поиском и позволяет создавать, изменять и удалять соревнования через REST API бэкенда.

## Структура

Структура `src` повторяет схему из лекции «Инженерный старт проекта».

```
frontend/
├── src/
│   ├── main.tsx                 # точка входа
│   ├── app.tsx                  # маршруты
│   ├── api/                     # запросы к REST API
│   │   ├── http-client.ts       #   fetch + разбор общего формата ответов и ошибок
│   │   └── competitions-api.ts
│   ├── assets/styles.css        # стили
│   ├── components/
│   │   ├── ui/                  # базовые элементы (поле формы)
│   │   └── elements/            # составные блоки (шапка, таблица, форма соревнования)
│   ├── constants/               # постоянные значения (подсказки видов спорта)
│   ├── hooks/requests/          # хуки загрузки данных
│   ├── pages/                   # страницы
│   ├── types/                   # типы данных API
│   └── utils/                   # вспомогательные функции (даты)
├── public/                      # статические файлы (иконка)
├── .env.development, .env.production
├── eslint.config.js, .prettierrc.json, .prettierignore, knip.json
├── nginx.conf, Dockerfile       # production-образ
└── vite.config.ts
```

## Установка и запуск

Нужен Node.js 24+.

```bash
npm install
npm run dev          # http://localhost:5173, запросы /api проксируются на http://localhost:8090
```

Бэкенд должен быть запущен (см. `backend/README.md` или `docker compose up` в корне).

| Команда                           | Что делает                                            |
| --------------------------------- | ----------------------------------------------------- |
| `npm run build`                   | проверка типов и production-сборка в `dist/`          |
| `npm run preview`                 | просмотр production-сборки                            |
| `npm run lint` / `lint:fix`       | ESLint (без предупреждений) / с автоисправлением      |
| `npm run format` / `format:check` | Prettier: отформатировать / проверить                 |
| `npm run typecheck`               | TypeScript без сборки                                 |
| `npm run knip`                    | поиск неиспользуемых файлов, экспортов и зависимостей |

## Технологии

| Технология                    | Зачем                                   | Документация                                                    |
| ----------------------------- | --------------------------------------- | --------------------------------------------------------------- |
| React 19                      | интерфейс                               | https://react.dev/                                              |
| TypeScript 6                  | типизация                               | https://www.typescriptlang.org/docs/                            |
| Vite 8                        | сервер разработки, прокси, сборка       | https://vite.dev/guide/                                         |
| React Router 8                | маршрутизация SPA                       | https://reactrouter.com/                                        |
| ESLint 10 + typescript-eslint | линтер                                  | https://eslint.org/docs/latest/ , https://typescript-eslint.io/ |
| eslint-plugin-unused-imports  | удаление неиспользуемых импортов        | https://github.com/sweepline/eslint-plugin-unused-imports       |
| Prettier 3                    | форматтер                               | https://prettier.io/docs/                                       |
| knip                          | поиск мёртвого кода                     | https://knip.dev/                                               |
| nginx                         | раздача сборки и прокси `/api` в Docker | https://nginx.org/ru/docs/                                      |

## Переменные окружения

Vite подставляет переменные с префиксом `VITE_` в код при сборке: `npm run dev` берёт их из `.env.development`, `npm run build` — из `.env.production`. Секретов в них нет, оба файла лежат в git. Локальные переопределения — в `.env.*.local` (в `.gitignore`).

| Переменная       | development             | production | Описание                                                                         |
| ---------------- | ----------------------- | ---------- | -------------------------------------------------------------------------------- |
| `VITE_APP_TITLE` | `SportReg (dev)`        | `SportReg` | название в шапке                                                                 |
| `VITE_API_URL`   | `/api`                  | `/api`     | префикс запросов к API                                                           |
| `BACKEND_URL`    | `http://localhost:8090` | —          | куда Vite проксирует `/api` (только для `vite.config.ts`, в браузер не попадает) |

## Соглашения о коде

- Файлы и папки в `src` — **kebab-case** (`competition-form.tsx`, `use-competitions.ts`). Проверяет хук pre-commit `file-names-frontend`.
- Компоненты и типы — `PascalCase`, функции, переменные и хуки — `camelCase` (хуки начинаются с `use`), константы модуля — `UPPER_SNAKE_CASE`. Проверяет правило ESLint `@typescript-eslint/naming-convention`.
- Именованные экспорты, без `export default` (кроме конфигов инструментов).
- Форматирование — Prettier: одинарные кавычки, без точек с запятой, ширина 100.
- Поля ответов API остаются как в JSON (`starts_on`), чтобы не переименовывать их при каждом запросе.

## Неочевидные решения

- **Прокси вместо CORS.** В разработке `/api` проксирует Vite, в Docker — nginx. Браузер видит один источник, поэтому бэкенду не нужны CORS-заголовки.
- **PATCH при изменении.** Форма отправляет только изменённые поля (`changedFields` в `competitions-page.tsx`), поэтому правка одного поля не перезаписывает чужие одновременные изменения других полей.
- **Ошибки валидации у полей.** Бэкенд возвращает 422 со списком `details[]`, и форма показывает каждое сообщение под своим полем.
- **Lock-файл и Linux.** `package-lock.json` создан заново после удаления `node_modules`. Lock-файл, выросший из шаблона на Windows, не содержал необязательных пакетов для Linux, и `npm ci` в Docker падал.
