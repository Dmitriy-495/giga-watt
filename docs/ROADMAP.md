# Giga Watt Roadmap

## Iteration 000 — Bootstrap

Цель:

Создать рабочую основу проекта.

Результат:

- структура репозитория;
- документация;
- запуск backend;
- запуск frontend;
- подключение базы данных.

---

## Iteration 001 — Foundation

Создание первого слоя цифрового двойника.

Foundation формирует базовые предметные сущности, организационную структуру и API, необходимые для дальнейшего развития системы.

---

## Iteration 002 — Platform User & Operator Workspace

### Goal

Создать первый рабочий пользовательский интерфейс оператора, начинающийся из его организационного контекста внутри цифрового двойника предприятия.

### Scope

- Platform User;
- initial superuser `tda`;
- User → Employee matching через `employee_emails`;
- поддержка Users в initial-data Excel;
- минимальный Current User API;
- определение организационного контекста;
- application shell;
- Header;
- Breadcrumbs;
- Operator Workspace;
- навигация к Foundation-сущностям.

### Key UX Principle

Оператор начинает работу из своего рабочего контекста.

Организационная иерархия отображается через breadcrumbs непосредственно после Header:

```text
Учреждение → Филиал → ЖКС → ПУ
```

Breadcrumbs отвечают на вопрос:

> Где я нахожусь?

Workspace отвечает на вопрос:

> Что я могу здесь делать?

### Explicitly Out of Scope

- authentication;
- passwords;
- sessions;
- JWT;
- OAuth/OIDC;
- roles;
- permissions;
- registration;
- password reset;
- полноценный authorization framework.

### Architectural Decisions

- ADR-0006 — Platform User Model;
- ADR-0007 — Operator Workspace and Organizational Context.

---

## Future Modules

Планируемые направления:

- Structure
- People
- Assets
- Geography
- Documents
- Emergency
- Maintenance
