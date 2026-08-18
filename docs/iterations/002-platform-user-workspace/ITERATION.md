# Iteration 002 — Platform User & Operator Workspace

## Mission

Создать первый рабочий пользовательский интерфейс Giga Watt, в котором оператор начинает работу непосредственно из своего организационного контекста внутри цифрового двойника предприятия.

Итерация связывает:

User → Employee → Employee Email → Assignment → Organization Unit → Workspace.

Основной UX-принцип:

> Оператор начинает работу из своего рабочего контекста, а не с общего дерева предприятия.

---

## Scope

### Platform User

Ввести минимальную сущность пользователя Платформы:

- id
- email
- is_superuser
- created_at
- updated_at

Требования:

- email — NOT NULL;
- email — UNIQUE;
- is_superuser — NOT NULL;
- is_superuser — DEFAULT false.

Создать начального пользователя:

- идентификатор: `tda`;
- is_superuser: true.

Полноценная authentication в Iteration 002 не реализуется.

---

## User → Employee

User является сущностью Платформы.

Employee является сущностью предприятия.

Сотрудник может существовать без User.

User в нормальном состоянии должен соответствовать Employee.

Сопоставление выполняется исключительно через:

    employee_emails

Алгоритм:

    users.email
        ↓
    employee_emails.email
        ↓
    employee_emails.employee_id
        ↓
    employees.id

Поле employees.email не используется как источник истины для данного сопоставления.

---

## User Creation

При создании User выполняется поиск сотрудника по employee_emails.

### Employee найден

User может считаться связанным с найденным Employee.

### Employee не найден

User может быть создан.

Система выдаёт предупреждение:

> Сотрудник с указанным email не найден.

Это warning, а не блокирующая ошибка.

### Найдено несколько сотрудников

Состояние считается неоднозначным и должно быть диагностировано.

---

## Initial Data / Excel

Initial data должна поддерживать импорт Users.

Users импортируются после Employee Emails.

Минимальный порядок:

1. Organization Units
2. Positions
3. Employees
4. Employee Emails
5. Employee Assignments
6. Operational Objects
7. Users

При импорте User выполняется то же правило сопоставления через employee_emails.

Начальный `tda` импортируется как superuser.

---

## Current Operator

Authentication, passwords, sessions, JWT, OAuth/OIDC, roles и permissions находятся вне scope Iteration 002.

На текущем этапе используется минимальный механизм определения текущего пользователя.

Frontend не должен самостоятельно вычислять организационный контекст.

Backend должен предоставить минимальный API-контракт текущего оператора.

---

## Operator Context

Рабочий контекст определяется через:

    User
      ↓
    Employee
      ↓
    Employee Assignment
      ↓
    Organization Unit

Если Employee имеет несколько назначений, правило выбора текущего контекста должно быть определено отдельно.

Нельзя вводить произвольный приоритет без предметного решения.

---

## Application Shell

Основная структура интерфейса:

    Header
    Breadcrumbs
    Workspace

Breadcrumbs размещаются непосредственно под Header.

---

## Breadcrumbs

Breadcrumbs отображают организационный путь:

    Учреждение → Филиал → ЖКС → ПУ

Breadcrumbs являются навигацией.

Родительские уровни позволяют перейти в соответствующий организационный контекст.

Полное дерево предприятия не является обязательным элементом стартового экрана.

---

## Operator Workspace

Стартовый экран показывает рабочее пространство текущего подразделения.

Workspace не должен быть универсальным dashboard framework.

На первом этапе должны быть доступны существующие Foundation-области:

- сотрудники;
- штатная структура;
- объекты эксплуатации.

---

## Frontend

Используется существующий frontend stack проекта.

Не добавлять новые зависимости без подтверждённой необходимости.

Не создавать:

- универсальный CRUD framework;
- универсальный entity framework;
- универсальный dashboard framework;
- избыточный navigation framework.

---

## Out of Scope

В Iteration 002 не реализуются:

- полноценная authentication;
- login/password;
- sessions;
- JWT;
- OAuth/OIDC;
- roles;
- permissions;
- registration;
- password reset;
- полноценный authorization framework;
- полноценный HR-контур.

---

## Completion Criteria

Iteration 002 считается завершённой, когда:

1. Создана модель Platform User.
2. Реализованы email и is_superuser.
3. Создан начальный пользователь tda.
4. User → Employee определяется через employee_emails.
5. Отсутствующий Employee диагностируется warning.
6. Неоднозначное соответствие диагностируется.
7. Initial-data Excel поддерживает Users.
8. Current User API реализован.
9. Frontend получает текущего оператора.
10. Frontend получает его организационный контекст.
11. Реализован Application Shell.
12. Breadcrumbs отображают организационную иерархию.
13. Breadcrumbs поддерживают навигацию.
14. Реализован стартовый Workspace подразделения.
15. Foundation-сущности доступны из Workspace.
16. Backend запускается локально.
17. Frontend запускается локально.
18. Fixtures загружаются воспроизводимо.
19. Regression tests проходят.
20. Frontend build проходит.
21. Документация соответствует реализации.
22. Создан CHECKPOINT итерации.
