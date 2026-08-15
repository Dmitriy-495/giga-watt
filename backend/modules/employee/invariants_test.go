package employee_test

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

func mustInsertEmployee(ctx context.Context, t *testing.T, tx pgx.Tx, lastName string) int64 {
	t.Helper()

	var id int64

	err := tx.QueryRow(ctx, `
		INSERT INTO employees (last_name, first_name, middle_name, short_name)
		VALUES ($1, 'Имя', 'Отчество', $2)
		RETURNING id`, lastName, lastName+" И.О.",
	).Scan(&id)
	if err != nil {
		t.Fatalf("setup: insert employee: %v", err)
	}

	return id
}

// TestEmployee_WithoutAssignment_Succeeds: сотрудник может существовать
// без кадрового назначения (см. ITERATION.md: штатная единица пока
// необязательна).
func TestEmployee_WithoutAssignment_Succeeds(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	mustInsertEmployee(ctx, t, tx, "БезНазначения")
}

// TestEmployee_ValidAssignment_Succeeds: сотрудник может иметь
// назначение, связанное с существующей организационной единицей и
// должностью.
func TestEmployee_ValidAssignment_Succeeds(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	emp := mustInsertEmployee(ctx, t, tx, "Сотрудник")
	unit := mustInsertOrgUnit(ctx, t, tx, "Организация")
	pos := mustInsertPosition(ctx, t, tx, "Должность")

	_, err := tx.Exec(ctx, `
		INSERT INTO employee_assignments
			(employee_id, organization_unit_id, position_id, starts_at)
		VALUES ($1, $2, $3, '2024-01-01')`,
		emp, unit, pos,
	)
	if err != nil {
		t.Fatalf("expected valid assignment to succeed, got error: %v", err)
	}
}

// TestEmployee_AssignmentInvalidReferences_Fail проверяет, что
// назначение, ссылающееся на несуществующего сотрудника, организационную
// единицу или должность, отклоняется БД.
func TestEmployee_AssignmentInvalidReferences_Fail(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	emp := mustInsertEmployee(ctx, t, tx, "Сотрудник2")
	unit := mustInsertOrgUnit(ctx, t, tx, "Организация2")
	pos := mustInsertPosition(ctx, t, tx, "Должность2")

	const nonExistentID = 999999999

	cases := []struct {
		name       string
		employeeID int64
		unitID     int64
		positionID int64
	}{
		{"unknown employee_id", nonExistentID, unit, pos},
		{"unknown organization_unit_id", emp, nonExistentID, pos},
		{"unknown position_id", emp, unit, nonExistentID},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testdb.Try(ctx, tx, func() error {
				_, err := tx.Exec(ctx, `
					INSERT INTO employee_assignments
						(employee_id, organization_unit_id, position_id, starts_at)
					VALUES ($1, $2, $3, '2024-01-01')`,
					tc.employeeID, tc.unitID, tc.positionID,
				)
				return err
			})

			if err == nil {
				t.Fatalf("expected error for %s, got none", tc.name)
			}
		})
	}
}

// TestEmployee_AssignmentInvalidDates_Fail проверяет, что ends_at раньше
// starts_at отклоняется CHECK-ограничением.
func TestEmployee_AssignmentInvalidDates_Fail(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	emp := mustInsertEmployee(ctx, t, tx, "Сотрудник3")
	unit := mustInsertOrgUnit(ctx, t, tx, "Организация3")
	pos := mustInsertPosition(ctx, t, tx, "Должность3")

	err := testdb.Try(ctx, tx, func() error {
		_, err := tx.Exec(ctx, `
			INSERT INTO employee_assignments
				(employee_id, organization_unit_id, position_id, starts_at, ends_at)
			VALUES ($1, $2, $3, '2024-01-01', '2023-01-01')`,
			emp, unit, pos,
		)
		return err
	})

	if err == nil {
		t.Fatal("expected error for ends_at < starts_at, got none")
	}
}

// TestEmployee_MultiplePhonesAndEmails_Succeed: у сотрудника может быть
// несколько телефонов и e-mail.
func TestEmployee_MultiplePhonesAndEmails_Succeed(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	emp := mustInsertEmployee(ctx, t, tx, "Контактный")

	for _, phone := range []string{"+7 900 000-00-01", "+7 900 000-00-02"} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO employee_phones (employee_id, phone) VALUES ($1, $2)`,
			emp, phone,
		); err != nil {
			t.Fatalf("expected multiple phones to succeed, got error: %v", err)
		}
	}

	for _, email := range []string{"one@example.com", "two@example.com"} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO employee_emails (employee_id, email) VALUES ($1, $2)`,
			emp, email,
		); err != nil {
			t.Fatalf("expected multiple emails to succeed, got error: %v", err)
		}
	}
}

// TestEmployee_TwoPrimaryPhones_Fail проверяет, что частичный уникальный
// индекс uq_employee_phones_primary не даёт создать второй основной
// телефон напрямую (в обход авто-переключения на уровне backend).
func TestEmployee_TwoPrimaryPhones_Fail(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	emp := mustInsertEmployee(ctx, t, tx, "ДваОсновныхТелефона")

	if _, err := tx.Exec(ctx, `
		INSERT INTO employee_phones (employee_id, phone, is_primary)
		VALUES ($1, '+7 900 111-11-11', TRUE)`, emp,
	); err != nil {
		t.Fatalf("setup: first primary phone: %v", err)
	}

	err := testdb.Try(ctx, tx, func() error {
		_, err := tx.Exec(ctx, `
			INSERT INTO employee_phones (employee_id, phone, is_primary)
			VALUES ($1, '+7 900 222-22-22', TRUE)`, emp,
		)
		return err
	})

	if err == nil {
		t.Fatal("expected error inserting a second primary phone, got none")
	}
}

// TestEmployee_TwoPrimaryEmails_Fail — аналогично для e-mail
// (uq_employee_emails_primary).
func TestEmployee_TwoPrimaryEmails_Fail(t *testing.T) {
	pool := testdb.Connect(t)
	tx := testdb.WithTx(t, pool)
	ctx := context.Background()

	emp := mustInsertEmployee(ctx, t, tx, "ДваОсновныхEmail")

	if _, err := tx.Exec(ctx, `
		INSERT INTO employee_emails (employee_id, email, is_primary)
		VALUES ($1, 'first@example.com', TRUE)`, emp,
	); err != nil {
		t.Fatalf("setup: first primary email: %v", err)
	}

	err := testdb.Try(ctx, tx, func() error {
		_, err := tx.Exec(ctx, `
			INSERT INTO employee_emails (employee_id, email, is_primary)
			VALUES ($1, 'second@example.com', TRUE)`, emp,
		)
		return err
	})

	if err == nil {
		t.Fatal("expected error inserting a second primary email, got none")
	}
}
