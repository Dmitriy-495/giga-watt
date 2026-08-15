package organization_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/Dmitriy-495/giga-watt/backend/internal/testdb"
)

// insertUnit создаёт организационную единицу в переданной транзакции и
// возвращает её id. Обёрнуто в testdb.Try (SAVEPOINT), чтобы ожидаемая
// ошибка в негативных тестах не переводила всю транзакцию в состояние
// aborted.
func insertUnit(ctx context.Context, tx pgx.Tx, typ, name string, parentID *int64) (int64, error) {
	var id int64

	err := testdb.Try(ctx, tx, func() error {
		return tx.QueryRow(ctx, `
			INSERT INTO organization_units (type, name, location, address, parent_id)
			VALUES ($1, $2, 'L', 'Ad', $3)
			RETURNING id`,
			typ, name, parentID,
		).Scan(&id)
	})

	return id, err
}

// updateUnit меняет type и parent_id существующей организационной
// единицы. См. insertUnit про testdb.Try.
func updateUnit(ctx context.Context, tx pgx.Tx, id int64, typ string, parentID *int64) error {
	return testdb.Try(ctx, tx, func() error {
		_, err := tx.Exec(ctx, `
			UPDATE organization_units SET type = $1, parent_id = $2 WHERE id = $3`,
			typ, parentID, id,
		)
		return err
	})
}

func ptr(v int64) *int64 { return &v }

// chain создаёт валидную цепочку Учреждение → Филиал → ЖКС → ПУ и
// возвращает id всех четырёх узлов по порядку.
func chain(ctx context.Context, t *testing.T, tx pgx.Tx) (inst, branch, jks, pu int64) {
	t.Helper()

	var err error

	inst, err = insertUnit(ctx, tx, "institution", "Учреждение", nil)
	if err != nil {
		t.Fatalf("setup institution: %v", err)
	}

	branch, err = insertUnit(ctx, tx, "branch", "Филиал", ptr(inst))
	if err != nil {
		t.Fatalf("setup branch: %v", err)
	}

	jks, err = insertUnit(ctx, tx, "jks", "ЖКС", ptr(branch))
	if err != nil {
		t.Fatalf("setup jks: %v", err)
	}

	pu, err = insertUnit(ctx, tx, "production_unit", "ПУ", ptr(jks))
	if err != nil {
		t.Fatalf("setup production_unit: %v", err)
	}

	return inst, branch, jks, pu
}

// TestHierarchy_ValidChain проверяет, что полная валидная цепочка
// Учреждение → Филиал → ЖКС → ПУ создаётся без ошибок.
func TestHierarchy_ValidChain(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	chain(ctx, t, tx)
}

// TestHierarchy_InvalidLevels проверяет, что запрещённые сочетания
// уровня и родителя отклоняются БД, используя валидную цепочку как
// источник родителей нужного уровня.
func TestHierarchy_InvalidLevels(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	inst, branch, jks, pu := chain(ctx, t, tx)

	cases := []struct {
		name      string
		childType string
		parentID  int64
	}{
		{"institution -> jks (skip branch)", "jks", inst},
		{"institution -> production_unit (skip branch,jks)", "production_unit", inst},
		{"branch -> production_unit (skip jks)", "production_unit", branch},
		{"jks -> branch (reversed)", "branch", jks},
		{"production_unit -> anything (leaf)", "production_unit", pu},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := insertUnit(ctx, tx, tc.childType, "Ребёнок", ptr(tc.parentID)); err == nil {
				t.Fatalf("expected error inserting %s under parent id=%d, got none", tc.childType, tc.parentID)
			}
		})
	}
}

// TestHierarchy_RootLevels проверяет допустимость/недопустимость
// корневого уровня (без родителя).
func TestHierarchy_RootLevels(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	cases := []struct {
		typ      string
		wantFail bool
	}{
		{"institution", false},
		{"branch", true},
		{"jks", true},
		{"production_unit", true},
	}

	for _, tc := range cases {
		t.Run(tc.typ, func(t *testing.T) {
			_, err := insertUnit(ctx, tx, tc.typ, "Корень", nil)

			if tc.wantFail && err == nil {
				t.Fatalf("expected error creating root-level %s, got none", tc.typ)
			}

			if !tc.wantFail && err != nil {
				t.Fatalf("unexpected error creating root-level %s: %v", tc.typ, err)
			}
		})
	}
}

// TestHierarchy_ChangeTypeWithChildren_Fails воспроизводит ровно
// сценарий из задания: Филиал с существующими ЖКС/ПУ пытаются превратить
// в Учреждение — это не должно проходить, иначе дети окажутся под
// недопустимым родителем.
func TestHierarchy_ChangeTypeWithChildren_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	_, branch, _, _ := chain(ctx, t, tx)

	// Филиал (branch) имеет ребёнка (ЖКС) — попытка сделать его
	// Учреждением (institution, без родителя) должна провалиться.
	if err := updateUnit(ctx, tx, branch, "institution", nil); err == nil {
		t.Fatal("expected error changing type of branch with existing children, got none")
	}
}

// TestHierarchy_ReparentWithoutTypeChange_Succeeds проверяет, что
// перепривязка узла к другому валидному родителю того же уровня — даже
// если у перепривязываемого узла есть собственные дети — разрешена и не
// нарушает целостность иерархии.
func TestHierarchy_ReparentWithoutTypeChange_Succeeds(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	inst, _, jks, _ := chain(ctx, t, tx)

	branchB, err := insertUnit(ctx, tx, "branch", "Филиал Б", ptr(inst))
	if err != nil {
		t.Fatalf("branch B: unexpected error: %v", err)
	}

	// У ЖКС есть собственный ребёнок (ПУ, создан внутри chain()), но мы
	// не меняем его type — только parent_id (переезд в другой Филиал).
	// Это должно быть разрешено.
	if err := updateUnit(ctx, tx, jks, "jks", ptr(branchB)); err != nil {
		t.Fatalf("expected reparenting (same type) to succeed, got error: %v", err)
	}
}

// TestHierarchy_TwoNodeCycle_Fails воспроизводит найденный дефект:
// попытка превратить родителя в потомка собственного ребёнка не должна
// проходить (в противном случае образуется цикл A <-> B).
func TestHierarchy_TwoNodeCycle_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	a, err := insertUnit(ctx, tx, "institution", "A", nil)
	if err != nil {
		t.Fatalf("A: unexpected error: %v", err)
	}

	b, err := insertUnit(ctx, tx, "branch", "B", ptr(a))
	if err != nil {
		t.Fatalf("B: unexpected error: %v", err)
	}

	// Пытаемся: A становится jks с родителем B (у B type=branch, jks
	// требует branch — тип формально подходит), но A — родитель B.
	if err := updateUnit(ctx, tx, a, "jks", ptr(b)); err == nil {
		t.Fatal("expected error creating a 2-node cycle (A <-> B), got none")
	}
}

// TestHierarchy_ThreeNodeCycle_Fails — то же самое для цепочки из трёх
// узлов: A -> B -> C, попытка сделать A потомком C.
func TestHierarchy_ThreeNodeCycle_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	a, err := insertUnit(ctx, tx, "institution", "A", nil)
	if err != nil {
		t.Fatalf("A: unexpected error: %v", err)
	}

	b, err := insertUnit(ctx, tx, "branch", "B", ptr(a))
	if err != nil {
		t.Fatalf("B: unexpected error: %v", err)
	}

	c, err := insertUnit(ctx, tx, "jks", "C", ptr(b))
	if err != nil {
		t.Fatalf("C: unexpected error: %v", err)
	}

	// Пытаемся: A становится production_unit с родителем C (jks) —
	// формально валидный тип-родитель, но замыкает цикл A -> B -> C -> A.
	if err := updateUnit(ctx, tx, a, "production_unit", ptr(c)); err == nil {
		t.Fatal("expected error creating a 3-node cycle (A -> B -> C -> A), got none")
	}
}

// TestHierarchy_SelfParent_Fails проверяет, что узел не может быть
// собственным родителем.
func TestHierarchy_SelfParent_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	inst, err := insertUnit(ctx, tx, "institution", "Учреждение", nil)
	if err != nil {
		t.Fatalf("institution: unexpected error: %v", err)
	}

	if err := updateUnit(ctx, tx, inst, "institution", ptr(inst)); err == nil {
		t.Fatal("expected error setting institution as its own parent, got none")
	}

	branch, err := insertUnit(ctx, tx, "branch", "Филиал", ptr(inst))
	if err != nil {
		t.Fatalf("branch: unexpected error: %v", err)
	}

	if err := updateUnit(ctx, tx, branch, "branch", ptr(branch)); err == nil {
		t.Fatal("expected error setting branch as its own parent, got none")
	}
}

// TestHierarchy_CycleWalkCatchesPreexistingCycle проверяет явную защиту
// от циклов (шаг 3 в validate_organization_unit_hierarchy) в изоляции от
// правила "нельзя менять type узла с детьми" (шаг 2).
//
// В штатной работе через API оба шага срабатывают вместе — цикл,
// найденный в ходе Foundation Hardening, перехватывается уже шагом 2,
// поэтому шаг 3 для него избыточен (см. CHECKPOINT_001-002.md). Чтобы
// всё же проверить именно шаг 3, тест временно отключает триггер и
// создаёт заведомо циклические данные напрямую (имитация состояния,
// которое могло возникнуть до фикса или при прямом вмешательстве в БД в
// обход приложения), затем включает триггер обратно и убеждается, что
// ЛЮБАЯ следующая запись в эти строки (даже не меняющая type/parent_id)
// отклоняется именно проверкой цикла.
func TestHierarchy_CycleWalkCatchesPreexistingCycle(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	if _, err := tx.Exec(ctx, `
		ALTER TABLE organization_units DISABLE TRIGGER trg_validate_organization_unit_hierarchy`,
	); err != nil {
		t.Fatalf("setup: disable trigger: %v", err)
	}

	var a, b int64

	if err := tx.QueryRow(ctx, `
		INSERT INTO organization_units (type, name, location, address)
		VALUES ('institution', 'A', 'L', 'Ad') RETURNING id`,
	).Scan(&a); err != nil {
		t.Fatalf("setup: insert A: %v", err)
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO organization_units (type, name, location, address, parent_id)
		VALUES ('jks', 'B', 'L', 'Ad', $1) RETURNING id`, a,
	).Scan(&b); err != nil {
		t.Fatalf("setup: insert B: %v", err)
	}

	// Замыкаем цикл A <-> B в обход триггера (недостижимо иначе).
	if _, err := tx.Exec(ctx, `
		UPDATE organization_units SET parent_id = $1 WHERE id = $2`, b, a,
	); err != nil {
		t.Fatalf("setup: force cycle bypassing trigger: %v", err)
	}

	if _, err := tx.Exec(ctx, `
		ALTER TABLE organization_units ENABLE TRIGGER trg_validate_organization_unit_hierarchy`,
	); err != nil {
		t.Fatalf("setup: enable trigger: %v", err)
	}

	// Тривиальное изменение: только name, type и parent_id не трогаем —
	// правило "нельзя менять type узла с детьми" здесь ни при чём.
	// Единственное, что может отклонить эту запись — независимая
	// проверка цикла.
	_, err := tx.Exec(ctx, `UPDATE organization_units SET name = 'A renamed' WHERE id = $1`, a)
	if err == nil {
		t.Fatal("expected the standalone cycle check to reject a write to an already-cyclic row, got none")
	}
}
