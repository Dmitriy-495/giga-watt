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

Основной стек:

- Nuxt 3
- TypeScript
- Nuxt UI
- Tailwind CSS
- Pinia

---

# Database

Изменения структуры базы данных выполняются только через миграции.

Прямое изменение схемы базы данных запрещено.

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
