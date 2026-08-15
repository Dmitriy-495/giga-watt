-- Giga Watt
-- Iteration 001 — Foundation
--
-- Организационная структура имеет фиксированную иерархию:
--
--     Учреждение
--         └── Филиал
--               └── ЖКС
--                     └── ПУ
--
-- Отношение между уровнями — 1:N (например, у одного Филиала может быть
-- несколько ЖКС, но ЖКС не может подчиняться напрямую Учреждению или
-- быть родителем другого ЖКС/Филиала).
--
-- Эта миграция добавляет self-reference parent_id и триггер, который
-- проверяет, что тип родителя соответствует ожидаемому уровню иерархии.
-- Проверка сделана в БД (а не только в backend), чтобы структура
-- оставалась целостной независимо от того, что её пишет.

ALTER TABLE organization_units
    ADD COLUMN parent_id BIGINT
        REFERENCES organization_units(id);

CREATE INDEX idx_organization_units_parent_id
    ON organization_units(parent_id);

CREATE OR REPLACE FUNCTION validate_organization_unit_hierarchy()
RETURNS TRIGGER AS $$
DECLARE
    parent_type VARCHAR(20);
BEGIN
    IF NEW.type = 'institution' THEN
        IF NEW.parent_id IS NOT NULL THEN
            RAISE EXCEPTION
                'organization_units: institution must not have a parent';
        END IF;

        RETURN NEW;
    END IF;

    IF NEW.parent_id IS NULL THEN
        RAISE EXCEPTION
            'organization_units: % must have a parent', NEW.type;
    END IF;

    SELECT type INTO parent_type
    FROM organization_units
    WHERE id = NEW.parent_id;

    IF parent_type IS NULL THEN
        RAISE EXCEPTION
            'organization_units: parent_id % not found', NEW.parent_id;
    END IF;

    IF (NEW.type = 'branch' AND parent_type <> 'institution')
        OR (NEW.type = 'jks' AND parent_type <> 'branch')
        OR (NEW.type = 'production_unit' AND parent_type <> 'jks')
    THEN
        RAISE EXCEPTION
            'organization_units: % cannot have parent of type %',
            NEW.type, parent_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_validate_organization_unit_hierarchy
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION validate_organization_unit_hierarchy();
