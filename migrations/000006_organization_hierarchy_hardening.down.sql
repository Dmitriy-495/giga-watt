-- Откатывает validate_organization_unit_hierarchy к виду, в котором она
-- была после 000003 (без защиты от смены type у узла с детьми и без
-- явной проверки циклов). Полное удаление функции/триггера — это уже
-- ответственность down-миграции 000003, а не этой.

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
