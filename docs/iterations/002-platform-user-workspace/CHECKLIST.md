# Iteration 002 — Checklist

## 1. Platform User

- [ ] Создана таблица users
- [ ] Добавлено поле email
- [ ] users.email NOT NULL
- [ ] users.email UNIQUE
- [ ] Добавлено поле is_superuser
- [ ] users.is_superuser NOT NULL DEFAULT false
- [ ] Создан пользователь tda
- [ ] tda.is_superuser = true
- [ ] Добавлены timestamps
- [ ] Migration имеет up/down
- [ ] Реализован минимальный User API

## 2. User → Employee

- [ ] Сопоставление выполняется через employee_emails
- [ ] employees.email не используется как источник истины
- [ ] Найден один Employee — связь определена
- [ ] Employee не найден — warning
- [ ] Найдено несколько Employee — диагностируется неоднозначность
- [ ] Правила покрыты тестами
- [ ] Не введён обязательный FK User → Employee
- [ ] Employee без User остаётся допустимым состоянием

## 3. Initial Data / Excel

- [ ] Excel поддерживает Users
- [ ] В initial data присутствует tda
- [ ] tda импортируется как superuser
- [ ] Employee Emails загружаются до Users
- [ ] Loader выполняет User → Employee matching
- [ ] Missing Employee выдаётся как warning
- [ ] Ambiguous Employee диагностируется
- [ ] Импорт остаётся воспроизводимым

## 4. Current Operator API

- [ ] Определён контракт текущего пользователя
- [ ] Backend предоставляет текущего User
- [ ] Backend предоставляет Employee при наличии
- [ ] Backend предоставляет организационный контекст
- [ ] API покрыт тестами

## 5. Frontend Shell

- [ ] Реализован основной layout
- [ ] Реализован Header
- [ ] Реализованы Breadcrumbs
- [ ] Реализован Workspace
- [ ] Реализован loading state
- [ ] Реализован error state
- [ ] Реализован необходимый empty state

## 6. Operator Context

- [ ] Frontend получает User
- [ ] Frontend получает Employee
- [ ] Frontend получает Organization Context
- [ ] Контекст не захардкожен во frontend
- [ ] Определено поведение нескольких назначений
- [ ] Определено поведение отсутствующего Employee

## 7. Breadcrumbs

- [ ] Отображается текущий организационный путь
- [ ] Поддерживается Учреждение → Филиал → ЖКС → ПУ
- [ ] Родительские уровни кликабельны
- [ ] Текущий уровень визуально обозначен
- [ ] Breadcrumbs расположены непосредственно под Header
- [ ] Полное дерево предприятия не дублируется

## 8. Workspace

- [ ] Реализован стартовый экран подразделения
- [ ] Отображается название подразделения
- [ ] Доступны сотрудники
- [ ] Доступна штатная структура
- [ ] Доступны объекты эксплуатации
- [ ] Не создан универсальный dashboard framework

## 9. Frontend Architecture

- [ ] Сохранён существующий frontend stack
- [ ] Нет ненужных новых dependencies
- [ ] Нет универсального CRUD framework
- [ ] Нет универсального entity framework
- [ ] Нет избыточного state management

## 10. Tests

- [ ] User tests
- [ ] Email uniqueness test
- [ ] User → Employee matching test
- [ ] Missing Employee warning test
- [ ] Ambiguous matching test
- [ ] Fixture loader tests
- [ ] Current User API tests
- [ ] Frontend checks/tests
- [ ] go test ./...
- [ ] go vet ./...
- [ ] gofmt
- [ ] Frontend build

## 11. Documentation

- [ ] ITERATION.md актуален
- [ ] CHECKLIST.md актуален
- [ ] NOTES.md актуален
- [ ] ADR Platform User опубликован
- [ ] ADR Operator Workspace опубликован
- [ ] ROADMAP.md обновлён
- [ ] README.md проверен
- [ ] CHANGELOG.md обновлён
- [ ] VERSION.md обновлён после завершения
- [ ] CHECKPOINT опубликован после завершения

## Completion

- [ ] Все обязательные пункты выполнены
- [ ] Backend запускается
- [ ] Frontend запускается
- [ ] Initial data загружается
- [ ] tda определяется как текущий User
- [ ] Employee определяется через employee_emails
- [ ] Организационный контекст определяется
- [ ] Workspace отображается
- [ ] Breadcrumb navigation работает
- [ ] Regression tests проходят
- [ ] Документация соответствует реализации
