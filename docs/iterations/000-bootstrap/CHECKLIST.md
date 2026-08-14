# Iteration 000 — Bootstrap Checklist

## Цель

Создать минимальную рабочую основу Giga Watt, пригодную для дальнейшей разработки.

---

## Repository

- [x] Создан локальный Git репозиторий
- [x] Создан GitHub репозиторий
- [x] Настроены ветки main/develop

---

## Project Structure

- [x] Создана структура каталогов
- [x] Созданы backend
- [x] Созданы frontend
- [x] Созданы docs
- [x] Созданы migrations
- [x] Созданы scripts

---

## Documentation

- [x] README
- [x] CHANGELOG
- [x] CONSTITUTION
- [x] ROADMAP
- [x] VERSION
- [x] DEVELOPMENT
- [x] SYSTEM Architecture

---

## ADR

- [x] ADR-0001 SQL-first
- [x] ADR-0002 Modular Monolith
- [x] ADR-0003 Iteration Memory
- [x] ADR-0004 Standard Library HTTP Bootstrap

---

## Backend

- [x] Go initialized
- [x] Standard library HTTP bootstrap (см. ADR-0004; Gin не используется)
- [x] `GET /api/ping` (возвращает `200 OK` и JSON)
- [x] Configuration (`backend/config`, cleanenv, `.env`)
- [x] Logger (`backend/platform/logger`, slog)
- [x] Graceful shutdown (SIGINT/SIGTERM, `backend/cmd/server/main.go`)
- [x] PostgreSQL connection (`backend/platform/database`, pgx)

---

## Frontend

- [x] Nuxt initialized (Nuxt 4, `frontend/`)
- [x] First page (`frontend/app.vue`, проверка подключения к backend через `/api/ping`)

---

## Database

- [x] PostgreSQL ready (локальная БД `giga_watt`)
- [x] golang-migrate (`migrations/`, `schema_migrations`)
- [x] Initial migration (`000001_bootstrap.up.sql`)

---

## Completion

- [x] CHECKPOINT_000-001 published

---

## Текущий статус

**DONE**

Итоговый результат:

```text
GET /api/ping

HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

Backend (Go, net/http, pgx) поднимается, подключается к PostgreSQL,
корректно завершает работу по SIGTERM/SIGINT (graceful shutdown).

Frontend (Nuxt 4) поднимается и отображает статус подключения к backend
(`Backend status: ok`) через `/api/ping`.

Миграция `000001_bootstrap` проверена на чистой локальной БД.

## Следующий этап

Iteration 001 — Foundation (см. `docs/iterations/001-foundation/`).

## Правило завершения

Последним действием Iteration 000 является публикация
`CHECKPOINT_000-001.md`.
