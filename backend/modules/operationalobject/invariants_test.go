package operationalobject_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/Dmitriy-495/giga-watt/backend/internal/testdb"
)

// orgChain создаёт валидную цепочку Учреждение → Филиал → ЖКС → ПУ и
// возвращает id всех четырёх узлов.
func orgChain(ctx context.Context, t *testing.T, tx pgx.Tx) (inst, branch, jks, pu int64) {
	t.Helper()

	mustScan := func(sql string, args ...any) int64 {
		var id int64
		if err := tx.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("setup: %v", err)
		}
		return id
	}

	inst = mustScan(`INSERT INTO organization_units (type, name, location, address)
		VALUES ('institution', 'Учреждение', 'L', 'Ad') RETURNING id`)
	branch = mustScan(`INSERT INTO organization_units (type, name, location, address, parent_id)
		VALUES ('branch', 'Филиал', 'L', 'Ad', $1) RETURNING id`, inst)
	jks = mustScan(`INSERT INTO organization_units (type, name, location, address, parent_id)
		VALUES ('jks', 'ЖКС', 'L', 'Ad', $1) RETURNING id`, branch)
	pu = mustScan(`INSERT INTO organization_units (type, name, location, address, parent_id)
		VALUES ('production_unit', 'ПУ', 'L', 'Ad', $1) RETURNING id`, jks)

	return inst, branch, jks, pu
}

func mustInsertObjectType(ctx context.Context, t *testing.T, tx pgx.Tx, name string) int64 {
	t.Helper()

	var id int64

	if err := tx.QueryRow(ctx, `
		INSERT INTO object_types (name) VALUES ($1) RETURNING id`, name,
	).Scan(&id); err != nil {
		t.Fatalf("setup: insert object_type: %v", err)
	}

	return id
}

func mustInsertObjectPurpose(ctx context.Context, t *testing.T, tx pgx.Tx, objectTypeID int64, name string) int64 {
	t.Helper()

	var id int64

	if err := tx.QueryRow(ctx, `
		INSERT INTO object_purposes (object_type_id, name) VALUES ($1, $2) RETURNING id`,
		objectTypeID, name,
	).Scan(&id); err != nil {
		t.Fatalf("setup: insert object_purpose: %v", err)
	}

	return id
}

// TestOperationalObject_OwnedByProductionUnit_Succeeds: объект
// эксплуатации, принадлежащий ПУ, создаётся без ошибок.
func TestOperationalObject_OwnedByProductionUnit_Succeeds(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	_, _, _, pu := orgChain(ctx, t, tx)
	objType := mustInsertObjectType(ctx, t, tx, "Коммунальный объект (тест)")

	if _, err := tx.Exec(ctx, `
		INSERT INTO operational_objects (organization_unit_id, object_type_id, name)
		VALUES ($1, $2, 'Котельная №1')`, pu, objType,
	); err != nil {
		t.Fatalf("expected object owned by production_unit to succeed, got error: %v", err)
	}
}

// TestOperationalObject_OwnedByNonProductionUnit_Fails: объект
// эксплуатации не может принадлежать Учреждению, Филиалу или ЖКС —
// только ПУ (см. validate_operational_object_owner).
func TestOperationalObject_OwnedByNonProductionUnit_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	inst, branch, jks, _ := orgChain(ctx, t, tx)
	objType := mustInsertObjectType(ctx, t, tx, "Коммунальный объект 2")

	cases := []struct {
		name   string
		unitID int64
	}{
		{"institution", inst},
		{"branch", branch},
		{"jks", jks},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testdb.Try(ctx, tx, func() error {
				_, err := tx.Exec(ctx, `
					INSERT INTO operational_objects (organization_unit_id, object_type_id, name)
					VALUES ($1, $2, 'Незаконный объект')`, tc.unitID, objType,
				)
				return err
			})

			if err == nil {
				t.Fatalf("expected error owning an operational object by %s, got none", tc.name)
			}
		})
	}
}

// TestOperationalObject_ValidTypePurpose_Succeeds: назначение,
// принадлежащее тому же типу, что указан у объекта, принимается.
func TestOperationalObject_ValidTypePurpose_Succeeds(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	_, _, _, pu := orgChain(ctx, t, tx)
	objType := mustInsertObjectType(ctx, t, tx, "Коммунальный объект 3")
	purpose := mustInsertObjectPurpose(ctx, t, tx, objType, "Котельная")

	if _, err := tx.Exec(ctx, `
		INSERT INTO operational_objects (organization_unit_id, object_type_id, object_purpose_id, name)
		VALUES ($1, $2, $3, 'Котельная №2')`, pu, objType, purpose,
	); err != nil {
		t.Fatalf("expected matching type/purpose to succeed, got error: %v", err)
	}
}

// TestOperationalObject_InvalidTypePurpose_Fails: назначение,
// принадлежащее ДРУГОМУ типу, чем указан у объекта, отклоняется составным
// внешним ключом fk_operational_objects_type_purpose.
func TestOperationalObject_InvalidTypePurpose_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	_, _, _, pu := orgChain(ctx, t, tx)

	typeA := mustInsertObjectType(ctx, t, tx, "Коммунальный объект 4")
	typeB := mustInsertObjectType(ctx, t, tx, "Объект ЖФ 4")
	purposeForB := mustInsertObjectPurpose(ctx, t, tx, typeB, "Многоквартирный дом")

	err := testdb.Try(ctx, tx, func() error {
		_, err := tx.Exec(ctx, `
			INSERT INTO operational_objects (organization_unit_id, object_type_id, object_purpose_id, name)
			VALUES ($1, $2, $3, 'Несоответствие')`, pu, typeA, purposeForB,
		)
		return err
	})

	if err == nil {
		t.Fatal("expected error for object_type/object_purpose mismatch, got none")
	}
}

// TestOperationalObject_InvalidReferences_Fail: несуществующие
// organization_unit_id/object_type_id отклоняются внешними ключами.
func TestOperationalObject_InvalidReferences_Fail(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	_, _, _, pu := orgChain(ctx, t, tx)
	objType := mustInsertObjectType(ctx, t, tx, "Коммунальный объект 5")

	const nonExistentID = 999999999

	cases := []struct {
		name   string
		unitID int64
		typeID int64
	}{
		{"unknown organization_unit_id", nonExistentID, objType},
		{"unknown object_type_id", pu, nonExistentID},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testdb.Try(ctx, tx, func() error {
				_, err := tx.Exec(ctx, `
					INSERT INTO operational_objects (organization_unit_id, object_type_id, name)
					VALUES ($1, $2, 'x')`, tc.unitID, tc.typeID,
				)
				return err
			})

			if err == nil {
				t.Fatalf("expected error for %s, got none", tc.name)
			}
		})
	}
}
