# CHECKPOINT 001-002 — Foundation Hardening + Fixtures

## Назначение

Технический проход между Iteration 001 (Foundation) и Iteration 002: не
расширять предметную область, а укрепить уже реализованный Foundation
(найти и закрыть реальные дефекты инвариантов, закрепить их
regression-тестами) и обеспечить воспроизводимое создание
демонстрационной/тестовой БД из реалистичных данных (Excel fixtures).

---

## Что было найдено

### 🔴 Критично: дефект триггера иерархии `organization_units`

`validate_organization_unit_hierarchy` (миграция `000003`) проверял
новую/изменённую строку только относительно её **непосредственного
родителя** в момент записи. Не проверялись:

1. существующие **дети** узла при смене его `type`;
2. **циклы** в цепочке `parent_id`.

**Воспроизведено** (напрямую через SQL, в обход API):

```sql
INSERT INTO organization_units (type, name, location, address)
  VALUES ('institution', 'A', 'L', 'Ad');                    -- id=1
INSERT INTO organization_units (type, name, location, address, parent_id)
  VALUES ('branch', 'B', 'L', 'Ad', 1);                       -- id=2, parent=1

UPDATE organization_units SET type='jks', parent_id=2 WHERE id=1;
-- ДО ФИКСА: проходило. Результат: id=1(type=jks,parent=2),
-- id=2(type=branch,parent=1) — цикл A<->B, и при этом B стал невалиден
-- (branch требует родителя institution, а не jks), но триггер для B не
-- перепроверялся.
```

Также воспроизведён 3-узловой цикл тем же способом (A→B→C, затем A
переводится в потомка C).

### 🟡 Расхождения документации с кодом

- `README.md` / `docs/DEVELOPMENT.md` утверждали, что frontend использует
  Nuxt UI, Tailwind CSS, Pinia — по факту (`frontend/package.json`)
  установлены только `nuxt`, `vue`, `vue-router`.
- `docs/VERSION.md` не обновлялся с Iteration 000 (`0.0.0-bootstrap`),
  несмотря на завершённые Iteration 000 и 001; нигде не была
  зафиксирована сама схема версионирования.
- `CHANGELOG.md` не содержал записи об Iteration 001.
- `ITERATION.md` перечислял "Должность" как "обязательный реквизит"
  сотрудника, что противоречило штатной модели того же документа
  (сотрудник может существовать без штатной единицы/назначения).
- `organization_units` не имеет `DELETE` (в отличие от остальных 6
  модулей) — решение существовало неявно, нигде не зафиксировано.
- `ITERATION.md` требовал seed-значения типов объектов (Коммунальный
  объект/Объект ЖФ/Объект КЖФ) — миграции их не создавали; "Объект КЖФ"
  не был протестирован ни разу.

---

## Что исправлено

### Иерархия организационных единиц

**Severity:** Critical (порча данных / потенциальный бесконечный цикл при
любом коде, рекурсивно обходящем предков/потомков).

**Root cause:** триггер проверял только "новая строка относительно её
родителя", не учитывая (a) существующих детей узла при смене типа, (b)
цикличность цепочки.

**Fix** (миграция `000006_organization_hierarchy_hardening.up.sql`,
`CREATE OR REPLACE FUNCTION`, старая `000003` не редактировалась):

1. При `UPDATE`: если у узла уже есть дочерние организационные единицы,
   смена его `type` запрещена. `parent_id` менять по-прежнему можно —
   доказано (см. ниже), что чистая перепривязка без смены типа не может
   ни нарушить валидность детей, ни создать цикл в этой 4-уровневой
   схеме типов.
2. Независимо от (1) — явная защита от циклов: перед сохранением
   проверяется, что цепочка `parent_id` от нового родителя не
   возвращается к самому узлу (до 10 шагов; глубина модели — 4 уровня).

**Почему не "запретить менять узел с детьми" целиком (type и parent_id
вместе)**: для этой конкретной схемы (каждый уровень требует родителя
строго другого, уникального типа) доказано, что узел физически не может
быть перепривязан к одному из своих потомков без одновременной смены
собственного типа — поэтому запрет только на `type` достаточен и не
ограничивает легитимную операцию "перенести ЖКС в другой Филиал".
Regression-тест `TestHierarchy_ReparentWithoutTypeChange_Succeeds`
закрепляет, что это по-прежнему разрешено.

**Regression tests** (`backend/modules/organization/hierarchy_test.go`):
`TestHierarchy_ValidChain`, `TestHierarchy_InvalidLevels`,
`TestHierarchy_RootLevels`, `TestHierarchy_ChangeTypeWithChildren_Fails`
(ровно сценарий из задания), `TestHierarchy_ReparentWithoutTypeChange_Succeeds`,
`TestHierarchy_TwoNodeCycle_Fails`, `TestHierarchy_ThreeNodeCycle_Fails`,
`TestHierarchy_SelfParent_Fails`.

### Документация

- `docs/adr/ADR-0005-versioning.md` (новый): схема версионирования
  `MAJOR.MINOR.PATCH-<iteration-slug>`.
- `docs/VERSION.md`: `0.0.0-bootstrap` → `0.1.1-foundation`.
- `CHANGELOG.md`: добавлены записи `0.1.0-foundation` (Iteration 001) и
  `0.1.1-foundation` (этот проход).
- `docs/DEVELOPMENT.md`: frontend-стек приведён к факту (Nuxt 4, Vue 3,
  TypeScript; явно отмечено, что UI/Tailwind/Pinia не подключены и не
  добавляются "на будущее"); добавлены разделы "Локальная настройка",
  "Применение миграций", "Сброс БД", "Fixtures", "Автоматические тесты" с
  проверенными командами.
- `docs/iterations/001-foundation/ITERATION.md`: убрано "Должность" из
  обязательных реквизитов сотрудника (должность — через кадровое
  назначение, не хранится в самом сотруднике), добавлено пояснение.
- `docs/iterations/001-foundation/NOTES.md`: Decision Log дополнен тремя
  записями (constraint для primary-контактов, триггер владения ПУ уже
  были реализованы ранее и теперь явно описаны; решение о запрете смены
  `type` у узла с детьми; решение не добавлять `DELETE` для
  `organization_units`).
- `backend/modules/organization/handler.go`: комментарий у
  `RegisterRoutes`, явно объясняющий отсутствие `DELETE`.

---

## Что сознательно не исправлялось

- **Seed-значения типов объектов** (Коммунальный объект/ЖФ/КЖФ) — не
  добавлены в миграцию как хардкод. Вместо этого они являются частью
  Excel fixtures (`fixtures/initial-data.xlsx`), что и демонстрирует
  реальное использование всех трёх типов, включая ранее непротестированный
  "Объект КЖФ" (экспериментальный ПУ №5 использует только его).
- **`organization_units` без `DELETE`** — осталось так же (см. выше),
  решение явно задокументировано, а не "исправлено" добавлением DELETE
  для формальной симметрии с другими модулями.
- **Explicit cycle-check (пункт 2 фикса) на практике недостижим отдельно
  от children-block (пункт 1)** в текущей 4-уровневой схеме — оба найденных
  сценария цикла (2- и 3-узловые) перехватываются именно children-block,
  а не explicit cycle-walk. Explicit cycle-walk оставлен как
  defense-in-depth на случай будущего изменения правил иерархии, а не
  убран как "мёртвый код" — это осознанный компромисс.
- **Down-миграции** по-прежнему отсутствуют (известное ограничение с
  CHECKPOINT_000-001).
- **Frontend остаётся read-only** (списки без форм) — fixtures и API
  дают полноценный CRUD и данные для прототипирования, но
  создание/редактирование через UI не входило в объём этого прохода.

---

## Изменения в архитектуре

Новых архитектурных решений, меняющих ранее принятые ADR, не вводилось.
Добавлен один новый ADR:

- **ADR-0005 — Versioning Scheme**: `MAJOR.MINOR.PATCH-<iteration-slug>`,
  `MINOR` растёт с каждой предметной итерацией, `PATCH` — с техническими
  проходами внутри итерации (как этот).

---

## Изменения в БД

| Миграция | Что делает |
|---|---|
| `000006_organization_hierarchy_hardening.up.sql` | `CREATE OR REPLACE FUNCTION validate_organization_unit_hierarchy`: запрет смены `type` у узла с детьми + явная проверка цикла по цепочке `parent_id` |

Полный список миграций теперь: `000001_bootstrap` … `000006_organization_hierarchy_hardening`
— применяются на чистой БД без ошибок, воспроизводимо (проверено дважды).

---

## Автоматические тесты

Go-тесты (`go test`) против реальной PostgreSQL, каждый в собственной
транзакции с гарантированным `ROLLBACK` (`backend/internal/testdb`) — не
оставляют следов в БД, могут запускаться на БД с уже загруженными
fixtures.

| Пакет | Тестов | Инварианты |
|---|---|---|
| `modules/organization` | 9 | иерархия: валидная цепочка, недопустимые уровни, корневой уровень, смена type с детьми (FAIL), репаринтинг без смены type (PASS), 2- и 3-узловые циклы (FAIL), self-parent (FAIL) |
| `modules/employee` | 7 (2 с subtests) | сотрудник без назначения (PASS), валидное назначение (PASS), невалидные ссылки employee/organization/position (FAIL), invalid dates (FAIL), несколько телефонов/email (PASS), два primary-телефона/email (FAIL) |
| `modules/staffposition` | 4 (2 с subtests) | валидная штатная единица (PASS), quantity ≤ 0 (FAIL), невалидные ссылки (FAIL), защищённое удаление используемой должности (FAIL) |
| `modules/operationalobject` | 5 (2 с subtests) | владение ПУ (PASS), владение Учреждением/Филиалом/ЖКС (FAIL x3), валидное/невалидное соответствие type/purpose, невалидные ссылки |

Итого 25 тестовых функций, все `PASS`. Запуск:

```sh
export TEST_DATABASE_URL="postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable"
cd backend && go test ./...
```

---

## Excel Fixtures

### Workbook structure

`fixtures/initial-data.xlsx`, 10 листов, бизнес-коды вместо технических
id (см. AGENTS.md п.9 — технические id генерируются при загрузке):

| Лист | Колонки |
|---|---|
| `organization` | code, type (Учреждение/Филиал/ЖКС/ПУ), parent_code, name, location, address, latitude, longitude, phone, email, leader_employee_code |
| `positions` | code, name |
| `object_types` | code, name |
| `object_purposes` | code, object_type_code, name |
| `employees` | code, last_name, first_name, middle_name, birth_date, gender (М/Ж) |
| `employee_phones` | employee_code, phone, is_primary |
| `employee_emails` | employee_code, email, is_primary |
| `employee_assignments` | employee_code, organization_code, position_code, assignment_type (основное/совместительство/временный перевод/совмещение), starts_at, ends_at |
| `staff_positions` | organization_code, position_code, quantity |
| `operational_objects` | code, organization_code, object_type_code, object_purpose_code, name, address, latitude, longitude |

Организационная иерархия — **один** лист `organization` с
самоссылающимся `parent_code` (а не 4 отдельных листа/справочника кодов
на каждый уровень) — это прямее отражает дерево, которое человек
рисует на бумаге, и не дублирует одну и ту же таблицу 4 раза.

### Dataset

Реалистичный срез предприятия ЖКХ Санкт-Петербурга:

- 1 Учреждение → 3 Филиала → 5 ЖКС → 6 ПУ (15 организационных единиц);
- 8 должностей;
- 32 сотрудника (реалистичные русские ФИО), из них 2 — без единого
  кадрового назначения, 2 — с несколькими назначениями (включая
  совместительство/совмещение), 1 — с историческим "временный перевод"
  (закрытым по `ends_at`);
- 15 телефонов и 11 e-mail с корректной primary-логикой;
- 15 штатных единиц;
- 3 типа объектов (Коммунальный/ЖФ/КЖФ — все три, включая ранее
  непротестированный КЖФ) и 12 назначений;
- 24 объекта эксплуатации, включая экспериментальный ПУ №5 (только
  2 объекта КЖФ, 0 коммунальных и 0 ЖФ — в точности правило из
  `ITERATION.md`).

### Loader

`backend/cmd/fixtures` (Go, `github.com/xuri/excelize/v2`) — отдельный
бинарник (`cmd/fixtures`, по аналогии с `cmd/server`), не часть HTTP API.
Переиспользует `backend/config` (та же БД, что и backend) и
`employee.BuildShortName` (та же логика автогенерации short_name, что и
в API — не продублирована).

Порядок загрузки (по зависимостям FK, проверено на реальной схеме):
`object_types` → `object_purposes` → `positions` → `organization`
(4 внутренних прохода по уровню иерархии) → `employees` →
`employee_assignments` → `employee_phones` → `employee_emails` →
`staff_positions` → `operational_objects` → отложенное простановление
`leader_employee_id` (руководитель — сотрудник, поэтому после
`employees`).

### Validation

Двухуровневая, как и требовалось:

1. **Excel/loader**: обязательные колонки на лист (явная ошибка при
   отсутствии листа/колонки), обязательные поля в строке, разрешение
   бизнес-кодов в id (явная ошибка "неизвестный code" с указанием листа,
   номера строки и кода), дубликаты code.
2. **PostgreSQL** остаётся финальной защитой: ошибки триггеров/FK/CHECK
   пробрасываются как есть (с указанием листа/строки/кода) — не
   дублируются в loader'е.

Проверено вручную (намеренно "сломанные" копии workbook, не попавшие в
репозиторий): отсутствующий файл, отсутствующий лист, дублирующийся
code, неизвестный `parent_code`, попытка привязать объект эксплуатации
не к ПУ (триггер БД корректно всплывает через loader).

### Transaction

Вся загрузка — одна транзакция (`pgx.Tx`). Любая ошибка (на любом этапе,
включая уже "успешно" загруженные более ранние листы) откатывает всё
целиком — проверено: после ошибки на листе `organization` в БД не
осталось даже ранее вставленных `object_types`/`positions`.

---

## Документация

Обновлены: `README.md`, `CHANGELOG.md`, `docs/VERSION.md`,
`docs/DEVELOPMENT.md`, `docs/adr/ADR-0005-versioning.md`,
`docs/iterations/001-foundation/{ITERATION,NOTES,CHECKLIST}.md`.

---

## Команды запуска

```sh
# 1. Чистая БД
dropdb giga_watt && createdb giga_watt

# 2. Миграции
migrate -path migrations \
  -database "postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable" up

# 3. Fixtures
cd backend
go run ./cmd/fixtures -file ../fixtures/initial-data.xlsx

# 4. Тесты
export TEST_DATABASE_URL="postgres://<user>:<password>@localhost:5432/giga_watt?sslmode=disable"
go test ./...

# 5. Backend
go run ./cmd/server

# 6. Frontend (в отдельном терминале)
cd ../frontend && npm run dev
```

---

## Результаты тестирования

- `go build ./...`, `go vet ./...`, `gofmt -l .` — чисто.
- `go test ./...` (25 тестов) — все `PASS`, дважды подряд, в том числе на
  БД с уже загруженными fixtures.
- Полный цикл reset → migrations → fixtures выполнен дважды подряд с
  идентичными счётчиками строк по всем 10 листам.
- Backend API (`GET /api/{organization-units,employees,positions,
  staff-positions,object-types,object-purposes,operational-objects}`) —
  все `200`, данные из fixtures видны и корректны (иерархия,
  руководители, экспериментальный ПУ №5).
- Frontend (Nuxt dev-сервер) — все страницы `200`, реальные имена вместо
  id (сотрудники, объекты эксплуатации и т.д.) отрисовываются верно.

---

## Что осталось техническим долгом

- Down-миграции.
- Frontend только read-only (без форм создания/редактирования).
- Explicit cycle-check в триггере — defense-in-depth, не имеет отдельного
  regression-теста, который заставил бы сработать именно его (а не
  children-block) в текущей 4-уровневой схеме.
- Нет CI (`.github/` пуст) — тесты и проверки запускаются только вручную
  локально.

---

## Готовность к Iteration 002

Foundation теперь: (а) защищён на уровне БД от найденного дефекта
иерархии и покрыт regression-тестами по основным инвариантам, (б) может
быть воспроизводимо развёрнут с реалистичным демонстрационным набором
данных одной командой после миграций. Это создаёт стабильную,
проверенную базу для перехода к следующей предметной области —
тема Iteration 002 требует отдельного обсуждения и согласования (см.
AGENTS.md, п.10/11) и не определяется этим документом.
