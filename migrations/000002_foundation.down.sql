-- Обратный порядок относительно 000002_foundation.up.sql: сначала
-- таблицы, ссылающиеся на другие (внешние ключи), затем то, на что они
-- ссылались. Индексы и ограничения удаляются автоматически вместе с
-- таблицей/колонкой.

DROP TABLE IF EXISTS operational_objects;
DROP TABLE IF EXISTS object_purposes;
DROP TABLE IF EXISTS object_types;
DROP TABLE IF EXISTS staff_positions;
DROP TABLE IF EXISTS employee_emails;
DROP TABLE IF EXISTS employee_phones;
DROP TABLE IF EXISTS employee_assignments;

-- organization_units.leader_employee_id ссылается на employees, поэтому
-- organization_units должна быть удалена раньше employees.
DROP TABLE IF EXISTS organization_units;

DROP TABLE IF EXISTS employees;
DROP TYPE IF EXISTS employee_gender;

DROP TABLE IF EXISTS positions;
