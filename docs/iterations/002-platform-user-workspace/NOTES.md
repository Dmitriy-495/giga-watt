# Iteration 002 — Notes

## Domain Decisions

### Platform User

User является сущностью Платформы.

Employee является сущностью предприятия.

Эти сущности не должны смешиваться.

### Employee Without User

Employee может существовать без User.

Это нормальное состояние.

Не каждый сотрудник предприятия обязан иметь доступ к Платформе.

### User Without Employee

User в нормальном состоянии должен соответствовать Employee.

Однако отсутствие Employee при создании User не блокирует создание.

Система выдаёт warning.

### Email Matching

Источник истины для поиска Employee:

    employee_emails

Алгоритм:

    User.email
        ↓
    employee_emails.email
        ↓
    employee_emails.employee_id
        ↓
    Employee

employees.email не используется как источник истины.

### Initial Data

Users импортируются после Employee Emails.

Начальный пользователь:

    tda

имеет:

    is_superuser = true

## Frontend Decisions

### Operator First

Стартовый экран представляет рабочее пространство оператора.

Оператор начинает работу из своего организационного контекста.

### Breadcrumbs

Breadcrumbs расположены непосредственно под Header.

Пример:

    Учреждение → Филиал → ЖКС → ПУ

Breadcrumbs являются одновременно:

- индикатором положения;
- навигацией по организационной иерархии.

### Current Context

Контекст определяется через Employee Assignment.

Если у сотрудника несколько действующих назначений, алгоритм выбора должен быть определён отдельно.

## Out of Scope

Не реализуются:

- authentication;
- login/password;
- sessions;
- JWT;
- OAuth/OIDC;
- roles;
- permissions;
- registration;
- password reset;
- authorization framework.

## Open Questions

### Multiple Assignments

Как определяется текущее назначение сотрудника при наличии нескольких действующих назначений?

До принятия правила произвольный выбор запрещён.

### Email Normalization

Необходимо определить единое правило нормализации email при сопоставлении User и Employee Email.

### Current User

Необходимо определить временный механизм определения текущего пользователя до появления полноценной authentication.

## Decision Log

2026-08-15 — введена минимальная модель Platform User.

2026-08-15 — User.email является обязательным и уникальным.

2026-08-15 — is_superuser имеет DEFAULT false.

2026-08-15 — tda является начальным superuser.

2026-08-15 — User → Employee определяется через employee_emails.

2026-08-15 — Employee может существовать без User.

2026-08-15 — отсутствие Employee при создании User является warning.

2026-08-15 — initial-data Excel должен поддерживать Users.

2026-08-15 — authentication, roles и permissions отложены.

2026-08-15 — стартовый UX строится вокруг организационного контекста оператора.

2026-08-15 — организационная иерархия отображается breadcrumbs под Header.
