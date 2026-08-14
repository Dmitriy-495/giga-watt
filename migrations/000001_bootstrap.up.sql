-- Giga Watt
-- Iteration 000 — Bootstrap
--
-- Bootstrap не реализует предметную область предприятия.
-- Эта миграция существует только для проверки механизма
-- golang-migrate и подключения backend к PostgreSQL.
--
-- Доменная модель (организационная структура, сотрудники,
-- объекты эксплуатации и т.д.) начинается со следующей
-- миграции в рамках Iteration 001 — Foundation.

CREATE TABLE bootstrap_healthcheck (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE,

    checked_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_bootstrap_healthcheck_single_row
        CHECK (id)
);

COMMENT ON TABLE bootstrap_healthcheck IS
    'Техническая таблица Bootstrap для проверки работы миграций. Не является частью предметной модели.';
