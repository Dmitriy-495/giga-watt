-- Giga Watt
-- Iteration 001 — Foundation
--
-- Базовая предметная модель:
-- Учреждение → Филиал → ЖКС → ПУ
--
-- ПУ является одновременно организационной,
-- территориальной и эксплуатационной единицей.
--
-- Объекты эксплуатации принадлежат только ПУ.

-- ============================================================
-- Организационная структура
-- ============================================================

CREATE TABLE organization_units (
    id BIGSERIAL PRIMARY KEY,

    type VARCHAR(20) NOT NULL,

    name VARCHAR(255) NOT NULL,

    location VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,

    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),

    phone VARCHAR(50),
    email VARCHAR(255),

    leader_employee_id BIGINT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_organization_units_type
    ON organization_units(type);

CREATE INDEX idx_organization_units_leader_employee_id
    ON organization_units(leader_employee_id);

-- ============================================================
-- Должности
-- ============================================================

CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,

    name VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_positions_name UNIQUE (name)
);

-- ============================================================
-- Сотрудники
-- ============================================================

CREATE TYPE employee_gender AS ENUM (
    'male',
    'female'
);

CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,

    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100) NOT NULL,

    short_name VARCHAR(255) NOT NULL,

    birth_date DATE,

    gender employee_gender,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_employees_last_name
    ON employees(last_name);

-- ============================================================
-- Кадровые назначения
--
-- Один сотрудник может иметь несколько назначений:
-- - основное;
-- - совместительство;
-- - временный перевод;
-- - совмещение;
-- - историческое назначение.
--
-- Штатная единица пока необязательна.
-- Это позволяет начать работу без полноценного штатного
-- расписания и развить его позднее без изменения модели
-- сотрудника.
-- ============================================================

CREATE TABLE employee_assignments (
    id BIGSERIAL PRIMARY KEY,

    employee_id BIGINT NOT NULL
        REFERENCES employees(id),

    organization_unit_id BIGINT NOT NULL
        REFERENCES organization_units(id),

    position_id BIGINT NOT NULL
        REFERENCES positions(id),

    assignment_type VARCHAR(30) NOT NULL DEFAULT 'primary',

    starts_at DATE NOT NULL,
    ends_at DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_employee_assignments_dates
        CHECK (ends_at IS NULL OR ends_at >= starts_at),

    CONSTRAINT chk_employee_assignments_type
        CHECK (
            assignment_type IN (
                'primary',
                'part_time',
                'temporary_transfer',
                'combination'
            )
        )
);

CREATE INDEX idx_employee_assignments_employee_id
    ON employee_assignments(employee_id);

CREATE INDEX idx_employee_assignments_organization_unit_id
    ON employee_assignments(organization_unit_id);

CREATE INDEX idx_employee_assignments_position_id
    ON employee_assignments(position_id);

-- ============================================================
-- Телефоны сотрудников
-- ============================================================

CREATE TABLE employee_phones (
    id BIGSERIAL PRIMARY KEY,

    employee_id BIGINT NOT NULL
        REFERENCES employees(id)
        ON DELETE CASCADE,

    phone VARCHAR(50) NOT NULL,

    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_employee_phones_employee_id
    ON employee_phones(employee_id);

-- ============================================================
-- E-mail сотрудников
-- ============================================================

CREATE TABLE employee_emails (
    id BIGSERIAL PRIMARY KEY,

    employee_id BIGINT NOT NULL
        REFERENCES employees(id)
        ON DELETE CASCADE,

    email VARCHAR(255) NOT NULL,

    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_employee_emails_employee_id
    ON employee_emails(employee_id);

-- ============================================================
-- Штатные единицы
--
-- Пока модель допускает работу сотрудников без штатной
-- единицы. Таблица создаётся заранее как отдельная область
-- штатной структуры и может использоваться по мере развития.
-- ============================================================

CREATE TABLE staff_positions (
    id BIGSERIAL PRIMARY KEY,

    organization_unit_id BIGINT NOT NULL
        REFERENCES organization_units(id),

    position_id BIGINT NOT NULL
        REFERENCES positions(id),

    quantity NUMERIC(8,2) NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_staff_positions_quantity
        CHECK (quantity > 0)
);

CREATE INDEX idx_staff_positions_organization_unit_id
    ON staff_positions(organization_unit_id);

CREATE INDEX idx_staff_positions_position_id
    ON staff_positions(position_id);

-- ============================================================
-- Типы объектов эксплуатации
--
-- Это справочник, а не enum.
-- Пользователь может изменять состав типов объектов.
-- ============================================================

CREATE TABLE object_types (
    id BIGSERIAL PRIMARY KEY,

    name VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_object_types_name UNIQUE (name)
);

-- ============================================================
-- Назначения объектов эксплуатации
--
-- Назначение также является редактируемым справочником.
--
-- object_type_id ограничивает набор назначений,
-- доступных для конкретного типа объекта.
-- ============================================================

CREATE TABLE object_purposes (
    id BIGSERIAL PRIMARY KEY,

    object_type_id BIGINT NOT NULL
        REFERENCES object_types(id),

    name VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_object_purposes_type_name
        UNIQUE (object_type_id, name),

    -- Требуется для составного FK operational_objects
    -- (object_purpose_id, object_type_id), который проверяет
    -- соответствие назначения типу объекта.
    CONSTRAINT uq_object_purposes_id_type
        UNIQUE (id, object_type_id)
);

CREATE INDEX idx_object_purposes_object_type_id
    ON object_purposes(object_type_id);

-- ============================================================
-- Объекты эксплуатации
--
-- Здание, сооружение, сеть, котельная, скважина и т.д.
-- являются разновидностями объекта эксплуатации.
--
-- Отдельной сущности "здание" или "сооружение" нет:
-- конкретная физическая природа объекта отражается
-- его наименованием и назначением.
--
-- Объект эксплуатации принадлежит только ПУ.
-- ============================================================

CREATE TABLE operational_objects (
    id BIGSERIAL PRIMARY KEY,

    organization_unit_id BIGINT NOT NULL
        REFERENCES organization_units(id),

    object_type_id BIGINT NOT NULL
        REFERENCES object_types(id),

    object_purpose_id BIGINT
        REFERENCES object_purposes(id),

    name VARCHAR(255) NOT NULL,

    address TEXT,

    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_operational_objects_organization_unit_id
    ON operational_objects(organization_unit_id);

CREATE INDEX idx_operational_objects_object_type_id
    ON operational_objects(object_type_id);

CREATE INDEX idx_operational_objects_object_purpose_id
    ON operational_objects(object_purpose_id);

-- ============================================================
-- Связь руководителя с организационной единицей
--
-- FK добавляется после создания employees,
-- поскольку organization_units и employees ссылаются
-- друг на друга предметно.
-- ============================================================

ALTER TABLE organization_units
    ADD CONSTRAINT fk_organization_units_leader_employee
    FOREIGN KEY (leader_employee_id)
    REFERENCES employees(id);

-- ============================================================
-- Контроль соответствия назначения типу объекта
--
-- Назначение должно принадлежать тому же типу объекта,
-- который указан у operational_objects.
-- ============================================================

ALTER TABLE operational_objects
    ADD CONSTRAINT fk_operational_objects_type_purpose
    FOREIGN KEY (object_purpose_id, object_type_id)
    REFERENCES object_purposes(id, object_type_id);

-- ============================================================
-- Уникальный состав типов организационных единиц
-- ============================================================

ALTER TABLE organization_units
    ADD CONSTRAINT chk_organization_units_type
    CHECK (
        type IN (
            'institution',
            'branch',
            'jks',
            'production_unit'
        )
    );

-- ============================================================
-- Дополнительные индексы
-- ============================================================

CREATE INDEX idx_organization_units_name
    ON organization_units(name);

CREATE INDEX idx_operational_objects_name
    ON operational_objects(name);
