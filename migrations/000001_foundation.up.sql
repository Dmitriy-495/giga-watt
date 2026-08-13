-- Структура предприятия
CREATE TABLE organization_units (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES organization_units(id),
    type VARCHAR(20) NOT NULL,
    code VARCHAR(50),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_organization_units_parent_id
    ON organization_units(parent_id);

-- Должности
CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Штатные единицы
CREATE TABLE staff_positions (
    id BIGSERIAL PRIMARY KEY,
    organization_unit_id BIGINT NOT NULL REFERENCES organization_units(id),
    position_id BIGINT NOT NULL REFERENCES positions(id),
    quantity NUMERIC(8,2) NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Сотрудники
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    staff_position_id BIGINT REFERENCES staff_positions(id),
    personnel_number VARCHAR(50),
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    hired_at DATE,
    fired_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Здания и сооружения
CREATE TABLE buildings (
    id BIGSERIAL PRIMARY KEY,
    organization_unit_id BIGINT REFERENCES organization_units(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    address TEXT,
    inventory_number VARCHAR(100),
    cadastral_number VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
