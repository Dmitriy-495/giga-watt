package employee

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const phoneColumns = `id, employee_id, phone, is_primary, created_at`

func scanPhone(row pgx.Row) (*Phone, error) {
	var p Phone

	if err := row.Scan(&p.ID, &p.EmployeeID, &p.Phone, &p.IsPrimary, &p.CreatedAt); err != nil {
		return nil, err
	}

	return &p, nil
}

// ListPhones возвращает телефоны сотрудника (основной — первым).
func (r *Repository) ListPhones(ctx context.Context, employeeID int64) ([]Phone, error) {
	rows, err := r.db.Query(ctx, `SELECT `+phoneColumns+`
		FROM employee_phones
		WHERE employee_id = $1
		ORDER BY is_primary DESC, id`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee: list phones: %w", err)
	}
	defer rows.Close()

	phones := make([]Phone, 0)

	for rows.Next() {
		p, err := scanPhone(rows)
		if err != nil {
			return nil, fmt.Errorf("employee: list phones: scan: %w", err)
		}

		phones = append(phones, *p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("employee: list phones: %w", err)
	}

	return phones, nil
}

// AddPhone добавляет телефон сотруднику. Если isPrimary=true, ранее
// основной телефон этого сотрудника (если был) снимается с основного
// в той же транзакции.
func (r *Repository) AddPhone(ctx context.Context, employeeID int64, phone string, isPrimary bool) (*Phone, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("employee: add phone: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if isPrimary {
		if _, err := tx.Exec(ctx, `
			UPDATE employee_phones SET is_primary = FALSE
			WHERE employee_id = $1 AND is_primary`, employeeID); err != nil {
			return nil, fmt.Errorf("employee: add phone: unset primary: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO employee_phones (employee_id, phone, is_primary)
		VALUES ($1, $2, $3)
		RETURNING `+phoneColumns,
		employeeID, phone, isPrimary,
	)

	p, err := scanPhone(row)
	if err != nil {
		return nil, wrapWriteError("add phone", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("employee: add phone: commit: %w", err)
	}

	return p, nil
}

// UpdatePhone обновляет номер и признак основного телефона.
//
// Если телефон не найден (или принадлежит другому сотруднику),
// возвращается ErrNotFound.
func (r *Repository) UpdatePhone(ctx context.Context, employeeID, phoneID int64, phone string, isPrimary bool) (*Phone, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("employee: update phone: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if isPrimary {
		if _, err := tx.Exec(ctx, `
			UPDATE employee_phones SET is_primary = FALSE
			WHERE employee_id = $1 AND is_primary AND id <> $2`, employeeID, phoneID); err != nil {
			return nil, fmt.Errorf("employee: update phone: unset primary: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		UPDATE employee_phones SET phone = $1, is_primary = $2
		WHERE id = $3 AND employee_id = $4
		RETURNING `+phoneColumns,
		phone, isPrimary, phoneID, employeeID,
	)

	p, err := scanPhone(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update phone", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("employee: update phone: commit: %w", err)
	}

	return p, nil
}

// DeletePhone удаляет телефон сотрудника.
func (r *Repository) DeletePhone(ctx context.Context, employeeID, phoneID int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM employee_phones
		WHERE id = $1 AND employee_id = $2`, phoneID, employeeID)
	if err != nil {
		return fmt.Errorf("employee: delete phone: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

const emailColumns = `id, employee_id, email, is_primary, created_at`

func scanEmail(row pgx.Row) (*Email, error) {
	var e Email

	if err := row.Scan(&e.ID, &e.EmployeeID, &e.Email, &e.IsPrimary, &e.CreatedAt); err != nil {
		return nil, err
	}

	return &e, nil
}

// ListEmails возвращает e-mail сотрудника (основной — первым).
func (r *Repository) ListEmails(ctx context.Context, employeeID int64) ([]Email, error) {
	rows, err := r.db.Query(ctx, `SELECT `+emailColumns+`
		FROM employee_emails
		WHERE employee_id = $1
		ORDER BY is_primary DESC, id`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee: list emails: %w", err)
	}
	defer rows.Close()

	emails := make([]Email, 0)

	for rows.Next() {
		e, err := scanEmail(rows)
		if err != nil {
			return nil, fmt.Errorf("employee: list emails: scan: %w", err)
		}

		emails = append(emails, *e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("employee: list emails: %w", err)
	}

	return emails, nil
}

// AddEmail добавляет e-mail сотруднику. Если isPrimary=true, ранее
// основной e-mail этого сотрудника (если был) снимается с основного
// в той же транзакции.
func (r *Repository) AddEmail(ctx context.Context, employeeID int64, email string, isPrimary bool) (*Email, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("employee: add email: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if isPrimary {
		if _, err := tx.Exec(ctx, `
			UPDATE employee_emails SET is_primary = FALSE
			WHERE employee_id = $1 AND is_primary`, employeeID); err != nil {
			return nil, fmt.Errorf("employee: add email: unset primary: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO employee_emails (employee_id, email, is_primary)
		VALUES ($1, $2, $3)
		RETURNING `+emailColumns,
		employeeID, email, isPrimary,
	)

	e, err := scanEmail(row)
	if err != nil {
		return nil, wrapWriteError("add email", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("employee: add email: commit: %w", err)
	}

	return e, nil
}

// UpdateEmail обновляет адрес и признак основного e-mail.
//
// Если e-mail не найден (или принадлежит другому сотруднику),
// возвращается ErrNotFound.
func (r *Repository) UpdateEmail(ctx context.Context, employeeID, emailID int64, email string, isPrimary bool) (*Email, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("employee: update email: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if isPrimary {
		if _, err := tx.Exec(ctx, `
			UPDATE employee_emails SET is_primary = FALSE
			WHERE employee_id = $1 AND is_primary AND id <> $2`, employeeID, emailID); err != nil {
			return nil, fmt.Errorf("employee: update email: unset primary: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		UPDATE employee_emails SET email = $1, is_primary = $2
		WHERE id = $3 AND employee_id = $4
		RETURNING `+emailColumns,
		email, isPrimary, emailID, employeeID,
	)

	e, err := scanEmail(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update email", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("employee: update email: commit: %w", err)
	}

	return e, nil
}

// DeleteEmail удаляет e-mail сотрудника.
func (r *Repository) DeleteEmail(ctx context.Context, employeeID, emailID int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM employee_emails
		WHERE id = $1 AND employee_id = $2`, emailID, employeeID)
	if err != nil {
		return fmt.Errorf("employee: delete email: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
