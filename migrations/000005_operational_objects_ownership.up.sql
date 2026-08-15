-- Giga Watt
-- Iteration 001 — Foundation
--
-- "Объект эксплуатации принадлежит ПУ. Другие уровни организационной
-- структуры (Учреждение, Филиал, ЖКС) не имеют собственных объектов
-- эксплуатации" (ITERATION.md).
--
-- Эта миграция добавляет триггер, проверяющий, что
-- operational_objects.organization_unit_id указывает на организационную
-- единицу с типом production_unit — по аналогии с триггером иерархии
-- organization_units (000003).

CREATE OR REPLACE FUNCTION validate_operational_object_owner()
RETURNS TRIGGER AS $$
DECLARE
    owner_type VARCHAR(20);
BEGIN
    SELECT type INTO owner_type
    FROM organization_units
    WHERE id = NEW.organization_unit_id;

    IF owner_type IS NULL THEN
        RAISE EXCEPTION
            'operational_objects: organization_unit_id % not found',
            NEW.organization_unit_id;
    END IF;

    IF owner_type <> 'production_unit' THEN
        RAISE EXCEPTION
            'operational_objects: organization_unit_id must reference a production_unit (ПУ), got %',
            owner_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_validate_operational_object_owner
    BEFORE INSERT OR UPDATE ON operational_objects
    FOR EACH ROW
    EXECUTE FUNCTION validate_operational_object_owner();
