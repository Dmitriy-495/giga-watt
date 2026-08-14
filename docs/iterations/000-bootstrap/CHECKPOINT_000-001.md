# CHECKPOINT 000-001 — Bootstrap

## Назначение итерации

Создать минимальную рабочую основу Giga Watt, пригодную для дальнейшей
разработки: структуру репозитория, базовую документацию, архитектурные
решения (ADR) и работающий сквозной стек backend → PostgreSQL, frontend →
backend.

Bootstrap намеренно не реализует предметную область предприятия.

---

## Выполненные работы

- Инициализирован репозиторий, ветки `main`/`develop`.
- Создана структура каталогов: `backend`, `frontend`, `docs`, `migrations`,
  `scripts`.
- Написана базовая документация: README, CHANGELOG, CONSTITUTION, ROADMAP,
  VERSION, DEVELOPMENT, SYSTEM Architecture.
- Приняты ADR-0001..ADR-0004 (SQL-first, Modular Monolith, Iteration Memory,
  Standard Library HTTP Bootstrap).
- Backend инициализирован на Go:
  - конфигурация (`backend/config`, cleanenv, `.env`);
  - логирование (`backend/platform/logger`, `slog`);
  - подключение к PostgreSQL (`backend/platform/database`, pgx);
  - HTTP-сервер на стандартной библиотеке (`net/http`, без Gin, см.
    ADR-0004);
  - graceful shutdown по `SIGINT`/`SIGTERM`;
  - `GET /api/ping`.
- Frontend инициализирован на Nuxt 4 (`frontend/`), страница `app.vue`
  проверяет подключение к backend через `/api/ping`.
- Настроен golang-migrate, создана первая техническая миграция
  `000001_bootstrap.up.sql` (таблица `bootstrap_healthcheck`, не является
  частью предметной модели).

---

## Созданные файлы

- `backend/cmd/server/main.go`
- `backend/config/config.go`
- `backend/platform/logger/logger.go`
- `backend/platform/database/database.go`
- `frontend/app.vue`, `frontend/nuxt.config.ts`, `frontend/package.json`
- `migrations/000001_bootstrap.up.sql`
- `docs/adr/ADR-0001..ADR-0004`
- `docs/iterations/000-bootstrap/{ITERATION,CHECKLIST,NOTES}.md`

---

## Принятые архитектурные решения

- SQL-first, без ORM (ADR-0001).
- Модульный монолит (ADR-0002).
- Каждая итерация ведёт чек-лист и завершается CHECKPOINT-документом
  (ADR-0003).
- HTTP-бутстрап на стандартной библиотеке `net/http`, Gin не используется
  (ADR-0004).

---

## Текущее состояние проекта

Сквозной стек проверен вручную на локальном окружении:

```text
Nuxt (frontend, :3000) → GET /api/ping → Go/net/http (backend, :8080)
                                              │
                                              └── PostgreSQL (giga_watt)
```

- `go build ./...` в `backend/` проходит без ошибок.
- Backend поднимается, подключается к PostgreSQL, отвечает на
  `GET /api/ping` (`200 OK`, `{"status":"ok"}`), корректно завершает
  работу по сигналу остановки.
- Frontend поднимается, страница отображает `Backend status: ok`.
- Локальная БД `giga_watt` пересоздана с нуля, миграция
  `000001_bootstrap` применена и проверена через `schema_migrations`.

---

## Известные ограничения

- Down-миграции отсутствуют.
- Нет docker-compose/Makefile для воспроизводимого локального окружения:
  `.env` для backend создаётся вручную (шаблон — `.env.example`), файл
  должен физически находиться рядом с `backend/go.mod`, так как
  `cleanenv.ReadConfig(".env", ...)` резолвит путь относительно текущей
  рабочей директории процесса.
- Нет CI-проверок (`.github/` пуст, кроме `.gitkeep`).
- Bootstrap-миграция (`bootstrap_healthcheck`) технически не удалена и
  остаётся в истории миграций — по решению из самой миграции она не
  является частью предметной модели и не требует отдельной очистки на
  этом этапе.

---

## Инструкции для следующего этапа

Iteration 001 — Foundation уже начата (см.
`docs/iterations/001-foundation/`):

- миграция `000002_foundation.up.sql` реализует предметную модель
  (организационная структура, сотрудники, кадровые назначения, штатные
  единицы, объекты эксплуатации) и проверена на чистой БД;
- backend-модули для Foundation (`backend/modules`) ещё не реализованы —
  только заготовка каталога;
- frontend-интерфейс Foundation не реализован — только страница проверки
  связи с backend.

Продолжать в `docs/iterations/001-foundation/` согласно `ITERATION.md` и
`NOTES.md` этой итерации.
