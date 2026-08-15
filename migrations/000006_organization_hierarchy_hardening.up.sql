-- Giga Watt
-- Foundation Hardening (после Iteration 001)
--
-- Обнаруженный дефект: validate_organization_unit_hierarchy (000003)
-- проверяет новую/изменённую строку только относительно её
-- НЕПОСРЕДСТВЕННОГО РОДИТЕЛЯ в момент записи. Она не проверяет:
--
--   1. что происходит с уже существующими ДЕТЬМИ узла, если у узла
--      меняется type;
--   2. что цепочка parent_id не образует ЦИКЛ.
--
-- Доказанная репродукция (см. CHECKPOINT Foundation Hardening):
--
--   A: institution (parent = NULL)
--   B: branch (parent = A)                       -- обычная валидная связь
--   UPDATE A SET type = 'jks', parent_id = B      -- проходило триггер:
--                                                  -- B сейчас type=branch,
--                                                  -- jks требует branch
--
-- Результат ДО фикса: A(type=jks, parent=B), B(type=branch, parent=A) —
-- A и B ссылаются друг на друга (цикл), а B при этом молча оказывается
-- невалидным (branch требует родителя institution, а не jks), потому что
-- триггер не перепроверяет B при изменении A.
--
-- Решение (минимальное, не меняющее модель иерархии):
--
--   1. При UPDATE: если у узла уже есть дочерние организационные единицы,
--      запрещается менять его type. (parent_id менять по-прежнему можно —
--      для этой строгой 4-уровневой цепочки типов перепривязка без смены
--      типа не может ни нарушить валидность детей, ни создать цикл: каждый
--      тип требует родителя строго другого, отличного типа, поэтому узел
--      физически не может быть перепривязан к одному из своих потомков без
--      одновременной смены собственного типа.)
--
--   2. Независимо от (1) — явная защита от циклов: перед сохранением
--      строки проверяется, что цепочка parent_id, начиная от нового
--      родителя, не возвращается к самому узлу.
--
-- (1) защищает целостность уже существующих детей узла.
-- (2) — самостоятельная, не полагающаяся на (1), защита от циклов.

CREATE OR REPLACE FUNCTION validate_organization_unit_hierarchy()
RETURNS TRIGGER AS $$
DECLARE
    parent_type VARCHAR(20);
    has_children BOOLEAN;
    ancestor_id BIGINT;
    hops INTEGER := 0;
BEGIN
    -- 1. Тип относительно непосредственного родителя (как в 000003).
    IF NEW.type = 'institution' THEN
        IF NEW.parent_id IS NOT NULL THEN
            RAISE EXCEPTION
                'organization_units: institution must not have a parent';
        END IF;
    ELSE
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
    END IF;

    -- 2. Смена type у узла с уже существующими детьми запрещена: иначе
    --    дети (чья валидность проверялась относительно СТАРОГО type этого
    --    узла) молча становятся невалидными, а триггер для них повторно
    --    не сработает.
    IF TG_OP = 'UPDATE' AND NEW.type <> OLD.type THEN
        SELECT EXISTS (
            SELECT 1 FROM organization_units WHERE parent_id = OLD.id
        ) INTO has_children;

        IF has_children THEN
            RAISE EXCEPTION
                'organization_units: cannot change type of % (id=%) while it has child organization units',
                OLD.type, OLD.id;
        END IF;
    END IF;

    -- 3. Явная защита от циклов: NEW.id не должен встречаться в цепочке
    --    предков NEW.parent_id. Глубина цепочки в этой модели не может
    --    превышать 4 уровня (institution → branch → jks → production_unit),
    --    поэтому ограничение в 10 шагов — с большим запасом, но всё ещё
    --    исключает риск бесконечного цикла даже при повреждённых данных.
    IF NEW.parent_id IS NOT NULL THEN
        ancestor_id := NEW.parent_id;

        WHILE ancestor_id IS NOT NULL AND hops < 10 LOOP
            IF ancestor_id = NEW.id THEN
                RAISE EXCEPTION
                    'organization_units: cyclic parent reference detected for id=%',
                    NEW.id;
            END IF;

            SELECT parent_id INTO ancestor_id
            FROM organization_units
            WHERE id = ancestor_id;

            hops := hops + 1;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
