package employee

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда запись не найдена.
var ErrNotFound = errors.New("not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть (нарушение ограничений БД: внешние ключи, CHECK, уникальность
// основного контакта).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к employees и связанным таблицам в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const employeeColumns = `
	id, last_name, first_name, middle_name, short_name,
	birth_date, gender, created_at`

func scanEmployee(row pgx.Row) (*Employee, error) {
	var e Employee

	err := row.Scan(
		&e.ID, &e.LastName, &e.FirstName, &e.MiddleName, &e.ShortName,
		&e.BirthDate, &e.Gender, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

// List возвращает всех сотрудников, упорядоченных по фамилии.
func (r *Repository) List(ctx context.Context) ([]Employee, error) {
	rows, err := r.db.Query(ctx, `SELECT `+employeeColumns+`
		FROM employees
		ORDER BY last_name, first_name, middle_name`)
	if err != nil {
		return nil, fmt.Errorf("employee: list: %w", err)
	}
	defer rows.Close()

	employees := make([]Employee, 0)

	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, fmt.Errorf("employee: list: scan: %w", err)
		}

		employees = append(employees, *e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("employee: list: %w", err)
	}

	return employees, nil
}

// Get возвращает сотрудника по id.
//
// Если сотрудник не найден, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*Employee, error) {
	row := r.db.QueryRow(ctx, `SELECT `+employeeColumns+`
		FROM employees
		WHERE id = $1`, id)

	e, err := scanEmployee(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("employee: get: %w", err)
	}

	return e, nil
}

// GetDetail возвращает сотрудника вместе с телефонами, e-mail и
// кадровыми назначениями.
func (r *Repository) GetDetail(ctx context.Context, id int64) (*Detail, error) {
	e, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	phones, err := r.ListPhones(ctx, id)
	if err != nil {
		return nil, err
	}

	emails, err := r.ListEmails(ctx, id)
	if err != nil {
		return nil, err
	}

	assignments, err := r.ListAssignments(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Detail{
		Employee:    *e,
		Phones:      phones,
		Emails:      emails,
		Assignments: assignments,
	}, nil
}

// Input — данные для создания или обновления сотрудника.
type Input struct {
	LastName   string
	FirstName  string
	MiddleName string
	BirthDate  *time.Time
	Gender     *string
}

// Create создаёт сотрудника. ShortName формируется автоматически из
// фамилии и инициалов (см. shortname.go).
func (r *Repository) Create(ctx context.Context, in Input) (*Employee, error) {
	shortName := BuildShortName(in.LastName, in.FirstName, in.MiddleName)

	row := r.db.QueryRow(ctx, `
		INSERT INTO employees
			(last_name, first_name, middle_name, short_name, birth_date, gender)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING `+employeeColumns,
		in.LastName, in.FirstName, in.MiddleName, shortName, in.BirthDate, in.Gender,
	)

	e, err := scanEmployee(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return e, nil
}

// Update обновляет сотрудника по id. ShortName пересчитывается заново.
//
// Если сотрудник не найден, возвращается ErrNotFound.
func (r *Repository) Update(ctx context.Context, id int64, in Input) (*Employee, error) {
	shortName := BuildShortName(in.LastName, in.FirstName, in.MiddleName)

	row := r.db.QueryRow(ctx, `
		UPDATE employees SET
			last_name = $1,
			first_name = $2,
			middle_name = $3,
			short_name = $4,
			birth_date = $5,
			gender = $6
		WHERE id = $7
		RETURNING `+employeeColumns,
		in.LastName, in.FirstName, in.MiddleName, shortName, in.BirthDate, in.Gender,
		id,
	)

	e, err := scanEmployee(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return e, nil
}

// Delete удаляет сотрудника по id.
//
// Телефоны и e-mail удаляются каскадно. Если у сотрудника есть кадровые
// назначения или он указан руководителем организационной единицы,
// удаление отклоняется как ValidationError.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		return wrapWriteError("delete", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// wrapWriteError превращает ошибки Postgres (внешние ключи,
// CHECK-ограничения, недопустимое значение gender) в ValidationError с
// понятным сообщением, остальные ошибки оборачиваются как внутренние.
func wrapWriteError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			return &ValidationError{
				Message: "сотрудник используется в назначениях или указан руководителем и не может быть удалён/изменён",
			}
		case "23514", "22P02": // check_violation / invalid enum value
			return &ValidationError{Message: pgErr.Message}
		}
	}

	return fmt.Errorf("employee: %s: %w", op, err)
}
