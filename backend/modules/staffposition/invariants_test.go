package staffposition_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/Dmitriy-495/giga-watt/backend/internal/testdb"
)

func mustInsertOrgUnit(ctx context.Context, t *testing.T, tx pgx.Tx, name string) int64 {
	t.Helper()

	var id int64

	err := tx.QueryRow(ctx, `
		INSERT INTO organization_units (type, name, location, address)
		VALUES ('institution', $1, 'L', 'Ad')
		RETURNING id`, name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("setup: insert organization_unit: %v", err)
	}

	return id
}

func mustInsertPosition(ctx context.Context, t *testing.T, tx pgx.Tx, name string) int64 {
	t.Helper()

	var id int64

	err := tx.QueryRow(ctx, `
		INSERT INTO positions (name) VALUES ($1) RETURNING id`, name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("setup: insert position: %v", err)
	}

	return id
}

// TestStaffPosition_Valid_Succeeds: штатная единица с существующими
// organization_unit и position и положительным количеством создаётся без
// ошибок.
func TestStaffPosition_Valid_Succeeds(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	unit := mustInsertOrgUnit(ctx, t, tx, "Организация")
	pos := mustInsertPosition(ctx, t, tx, "Должность")

	if _, err := tx.Exec(ctx, `
		INSERT INTO staff_positions (organization_unit_id, position_id, quantity)
		VALUES ($1, $2, 2.5)`, unit, pos,
	); err != nil {
		t.Fatalf("expected valid staff position to succeed, got error: %v", err)
	}
}

// TestStaffPosition_InvalidQuantity_Fails: quantity <= 0 отклоняется
// CHECK-ограничением chk_staff_positions_quantity.
func TestStaffPosition_InvalidQuantity_Fails(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	unit := mustInsertOrgUnit(ctx, t, tx, "Организация2")
	pos := mustInsertPosition(ctx, t, tx, "Должность2")

	cases := []float64{0, -1}

	for _, qty := range cases {
		err := testdb.Try(ctx, tx, func() error {
			_, err := tx.Exec(ctx, `
				INSERT INTO staff_positions (organization_unit_id, position_id, quantity)
				VALUES ($1, $2, $3)`, unit, pos, qty,
			)
			return err
		})

		if err == nil {
			t.Fatalf("expected error for quantity=%v, got none", qty)
		}
	}
}

// TestStaffPosition_InvalidReferences_Fail: несуществующие
// organization_unit_id/position_id отклоняются внешними ключами.
func TestStaffPosition_InvalidReferences_Fail(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	unit := mustInsertOrgUnit(ctx, t, tx, "Организация3")
	pos := mustInsertPosition(ctx, t, tx, "Должность3")

	const nonExistentID = 999999999

	cases := []struct {
		name   string
		unitID int64
		posID  int64
	}{
		{"unknown organization_unit_id", nonExistentID, pos},
		{"unknown position_id", unit, nonExistentID},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testdb.Try(ctx, tx, func() error {
				_, err := tx.Exec(ctx, `
					INSERT INTO staff_positions (organization_unit_id, position_id, quantity)
					VALUES ($1, $2, 1)`, tc.unitID, tc.posID,
				)
				return err
			})

			if err == nil {
				t.Fatalf("expected error for %s, got none", tc.name)
			}
		})
	}
}

// TestStaffPosition_DeletePosition_BlockedWhileInUse проверяет, что
// должность, используемая штатной единицей, не может быть удалена
// (внешний ключ staff_positions.position_id без ON DELETE CASCADE).
func TestStaffPosition_DeletePosition_BlockedWhileInUse(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	unit := mustInsertOrgUnit(ctx, t, tx, "Организация4")
	pos := mustInsertPosition(ctx, t, tx, "Должность4")

	if _, err := tx.Exec(ctx, `
		INSERT INTO staff_positions (organization_unit_id, position_id, quantity)
		VALUES ($1, $2, 1)`, unit, pos,
	); err != nil {
		t.Fatalf("setup: staff position: %v", err)
	}

	err := testdb.Try(ctx, tx, func() error {
		_, err := tx.Exec(ctx, `DELETE FROM positions WHERE id = $1`, pos)
		return err
	})

	if err == nil {
		t.Fatal("expected error deleting a position still referenced by a staff position, got none")
	}
}
