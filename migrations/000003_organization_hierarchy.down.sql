DROP TRIGGER IF EXISTS trg_validate_organization_unit_hierarchy ON organization_units;
DROP FUNCTION IF EXISTS validate_organization_unit_hierarchy();

-- Удаление колонки автоматически удаляет её индекс
-- (idx_organization_units_parent_id).
ALTER TABLE organization_units DROP COLUMN IF EXISTS parent_id;
