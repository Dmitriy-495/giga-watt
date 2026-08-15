# Giga Watt

## Enterprise Digital Twin Platform

Giga Watt — платформа цифрового двойника предприятия коммунальной инфраструктуры.

Проект предназначен для создания цифровой модели предприятия, его структуры, объектов, процессов и событий.

---

## Current Iteration

Iteration 002 (тема согласовывается)

Текущая версия: `0.1.1-foundation` (см. `docs/VERSION.md`, ADR-0005).

Завершённые итерации / технические проходы:

- Iteration 000 — Bootstrap, см.
  [CHECKPOINT_000-001](docs/iterations/000-bootstrap/CHECKPOINT_000-001.md).
- Iteration 001 — Foundation, см.
  [CHECKPOINT_001-001](docs/iterations/001-foundation/CHECKPOINT_001-001.md).
- Foundation Hardening + Fixtures, см.
  [CHECKPOINT_001-002](docs/iterations/001-foundation/CHECKPOINT_001-002.md).

---

## Philosophy

Основные принципы проекта:

1. Предметная область важнее технологий.
2. Простые решения предпочтительнее сложных.
3. Архитектура должна расширяться, а не усложняться.
4. Каждая итерация должна завершаться работающим результатом.
5. Архитектурные решения фиксируются документально.

---

## Technology Stack

### Backend

- Go
- net/http (см. ADR-0004)
- PostgreSQL
- pgx
- SQL-first
- golang-migrate
- slog
- cleanenv

### Frontend

- Nuxt 4
- Vue 3
- TypeScript

(Nuxt UI / Tailwind CSS / Pinia не подключены — см. `docs/DEVELOPMENT.md`)

---

## Development

Подробная информация находится в каталоге:


docs/


---

## Status

Foundation domain model (organizational structure, employees, staff
positions, operational objects) implemented and verified end-to-end,
hardened against a real hierarchy-corruption defect (see
CHECKPOINT_001-002), and reproducible via `fixtures/initial-data.xlsx`
(see `docs/DEVELOPMENT.md`).
