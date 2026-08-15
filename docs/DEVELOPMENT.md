# Development Guide

## Overview

Документ описывает правила разработки Giga Watt.

---

# Development Philosophy

Разработка ведётся итерациями.

Каждая итерация должна создавать работающий результат.

---

# Branch Strategy

Основные ветки:


main
develop


main:

- стабильная история проекта.

develop:

- текущая разработка.

---

# Backend

Основной стек:

- Go
- net/http
- PostgreSQL
- pgx
- SQL-first
- golang-migrate
- slog

---

# Frontend

Основной стек (см. `frontend/package.json`):

- Nuxt 4
- Vue 3
- TypeScript

Nuxt UI, Tailwind CSS и Pinia не подключены — текущие страницы
(read-only списки) не создавали в них потребности. Если и когда
появится реальная необходимость (формы, состояние на нескольких
страницах), зависимость добавляется отдельным решением, а не заранее
"на будущее" (см. AGENTS.md, п. 2).

---

# Database

Изменения структуры базы данных выполняются только через миграции
(`migrations/*.up.sql`, инструмент — `golang-migrate`).

Прямое изменение схемы базы данных запрещено.

## Локальная настройка

Требуется локальный PostgreSQL и созданная (пустая) БД, например:

```sh
createdb giga_watt
```

`backend/.env` (не коммитится, см. `.env.example`) должен указывать на
неё, например:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=giga_watt
DB_USER=<ваш пользователь>
DB_PASSWORD=<пароль>
```

## Применение миграций

Из корня репозитория:

```sh
migrate -path migrations \
  -database "postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable" \
  up
```

## Сброс БД (reset)

Для получения чистого, воспроизводимого состояния:

```sh
dropdb giga_watt && createdb giga_watt
migrate -path migrations -database "postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable" up
```

## Fixtures (начальные данные)

`fixtures/initial-data.xlsx` — воспроизводимый Excel-источник
demo/test-данных (не production-импорт, см.
`docs/iterations/001-foundation/CHECKPOINT_001-002.md`). Структура и
формат описаны прямо в книге (листы = таблицы предметной области,
бизнес-коды вместо технических id).

Загрузка (из `backend/`, после применения миграций на **чистую** БД —
loader не идемпотентен и не предназначен для повторной загрузки поверх
уже заполненной БД):

```sh
cd backend
go run ./cmd/fixtures -file ../fixtures/initial-data.xlsx
```

Загрузка выполняется в одной транзакции: любая ошибка (отсутствующий
лист/колонка, неизвестный бизнес-код, нарушение ограничения БД) отменяет
всю загрузку целиком и печатает диагностическое сообщение с листом,
номером строки и кодом строки.

Полный воспроизводимый цикл:

```sh
dropdb giga_watt && createdb giga_watt
migrate -path migrations -database "postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable" up
cd backend && go run ./cmd/fixtures -file ../fixtures/initial-data.xlsx
```

## Автоматические тесты

Regression-тесты Foundation-инвариантов (`go test`) подключаются к
реальной PostgreSQL с уже применёнными миграциями и выполняются в
транзакциях, которые всегда откатываются — тесты не оставляют следов в
БД и могут запускаться на БД с уже загруженными fixtures.

```sh
export TEST_DATABASE_URL="postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable"
cd backend && go test ./...
```

Если `TEST_DATABASE_URL` не задана, тесты, требующие БД, пропускаются
(`SKIP`), а не падают.

---

# Documentation

Архитектурные решения фиксируются в ADR.

История чатов не является источником архитектуры.

---

# Iterations

Каждая итерация должна иметь:

- цель;
- ограничения;
- критерий завершения;
- результат.
