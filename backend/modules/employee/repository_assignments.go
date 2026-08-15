package employee

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const assignmentColumns = `
	id, employee_id, organization_unit_id, position_id,
	assignment_type, starts_at, ends_at, created_at`

func scanAssignment(row pgx.Row) (*Assignment, error) {
	var a Assignment

	err := row.Scan(
		&a.ID, &a.EmployeeID, &a.OrganizationUnitID, &a.PositionID,
		&a.AssignmentType, &a.StartsAt, &a.EndsAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// ListAssignments возвращает кадровые назначения сотрудника (последние
// по дате начала — первыми).
func (r *Repository) ListAssignments(ctx context.Context, employeeID int64) ([]Assignment, error) {
	rows, err := r.db.Query(ctx, `SELECT `+assignmentColumns+`
		FROM employee_assignments
		WHERE employee_id = $1
		ORDER BY starts_at DESC, id DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee: list assignments: %w", err)
	}
	defer rows.Close()

	assignments := make([]Assignment, 0)

	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("employee: list assignments: scan: %w", err)
		}

		assignments = append(assignments, *a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("employee: list assignments: %w", err)
	}

	return assignments, nil
}

// AssignmentInput — данные для создания или обновления кадрового
// назначения.
type AssignmentInput struct {
	OrganizationUnitID int64
	PositionID         int64
	AssignmentType     string
	StartsAt           time.Time
	EndsAt             *time.Time
}

// AddAssignment создаёт кадровое назначение сотрудника.
//
// Соответствие organization_unit_id/position_id существующим записям,
// допустимость assignment_type и корректность дат проверяются
// ограничениями БД; при нарушении вернётся *ValidationError.
func (r *Repository) AddAssignment(ctx context.Context, employeeID int64, in AssignmentInput) (*Assignment, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO employee_assignments
			(employee_id, organization_unit_id, position_id,
			 assignment_type, starts_at, ends_at)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING `+assignmentColumns,
		employeeID, in.OrganizationUnitID, in.PositionID,
		in.AssignmentType, in.StartsAt, in.EndsAt,
	)

	a, err := scanAssignment(row)
	if err != nil {
		return nil, wrapAssignmentError("add assignment", err)
	}

	return a, nil
}

// UpdateAssignment обновляет кадровое назначение сотрудника.
//
// Если назначение не найдено (или принадлежит другому сотруднику),
// возвращается ErrNotFound.
func (r *Repository) UpdateAssignment(ctx context.Context, employeeID, assignmentID int64, in AssignmentInput) (*Assignment, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE employee_assignments SET
			organization_unit_id = $1,
			position_id = $2,
			assignment_type = $3,
			starts_at = $4,
			ends_at = $5
		WHERE id = $6 AND employee_id = $7
		RETURNING `+assignmentColumns,
		in.OrganizationUnitID, in.PositionID, in.AssignmentType,
		in.StartsAt, in.EndsAt, assignmentID, employeeID,
	)

	a, err := scanAssignment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapAssignmentError("update assignment", err)
	}

	return a, nil
}

// DeleteAssignment удаляет кадровое назначение сотрудника.
func (r *Repository) DeleteAssignment(ctx context.Context, employeeID, assignmentID int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM employee_assignments
		WHERE id = $1 AND employee_id = $2`, assignmentID, employeeID)
	if err != nil {
		return fmt.Errorf("employee: delete assignment: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// wrapAssignmentError превращает ошибки Postgres при записи кадрового
// назначения в понятные ValidationError, используя имя нарушенного
// ограничения, а не общее сообщение об ошибке employee-репозитория.
func wrapAssignmentError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "employee_assignments_organization_unit_id_fkey":
			return &ValidationError{Message: "указан несуществующий organization_unit_id"}
		case "employee_assignments_position_id_fkey":
			return &ValidationError{Message: "указан несуществующий position_id"}
		case "employee_assignments_employee_id_fkey":
			return &ValidationError{Message: "сотрудник не найден"}
		case "chk_employee_assignments_dates":
			return &ValidationError{Message: "ends_at должен быть не раньше starts_at"}
		case "chk_employee_assignments_type":
			return &ValidationError{Message: "assignment_type: недопустимое значение"}
		}

		switch pgErr.Code {
		case "23503":
			return &ValidationError{Message: "указан несуществующий organization_unit_id или position_id"}
		case "23514":
			return &ValidationError{Message: pgErr.Message}
		}
	}

	return fmt.Errorf("employee: %s: %w", op, err)
}
