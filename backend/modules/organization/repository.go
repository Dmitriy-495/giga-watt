package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда организационная единица не найдена.
var ErrNotFound = errors.New("organization unit not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть (например, нарушение иерархии, проверяемое триггером в БД, или
// ссылка на несуществующий parent_id/leader_employee_id).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к organization_units в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const unitColumns = `
	id, parent_id, type, name, location, address,
	latitude, longitude, phone, email, leader_employee_id, created_at`

func scanUnit(row pgx.Row) (*Unit, error) {
	var u Unit

	err := row.Scan(
		&u.ID, &u.ParentID, &u.Type, &u.Name, &u.Location, &u.Address,
		&u.Latitude, &u.Longitude, &u.Phone, &u.Email, &u.LeaderEmployeeID,
		&u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// List возвращает все организационные единицы, упорядоченные по id.
func (r *Repository) List(ctx context.Context) ([]Unit, error) {
	rows, err := r.db.Query(ctx, `SELECT `+unitColumns+`
		FROM organization_units
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("organization: list: %w", err)
	}
	defer rows.Close()

	units := make([]Unit, 0)

	for rows.Next() {
		u, err := scanUnit(rows)
		if err != nil {
			return nil, fmt.Errorf("organization: list: scan: %w", err)
		}

		units = append(units, *u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("organization: list: %w", err)
	}

	return units, nil
}

// Get возвращает организационную единицу по id.
//
// Если единица не найдена, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*Unit, error) {
	row := r.db.QueryRow(ctx, `SELECT `+unitColumns+`
		FROM organization_units
		WHERE id = $1`, id)

	u, err := scanUnit(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("organization: get: %w", err)
	}

	return u, nil
}

// Input — данные для создания или обновления организационной единицы.
type Input struct {
	ParentID         *int64
	Type             string
	Name             string
	Location         string
	Address          string
	Latitude         *float64
	Longitude        *float64
	Phone            *string
	Email            *string
	LeaderEmployeeID *int64
}

// Create создаёт организационную единицу.
//
// Соответствие иерархии проверяется триггером в БД
// (validate_organization_unit_hierarchy); при нарушении вернётся
// *ValidationError.
func (r *Repository) Create(ctx context.Context, in Input) (*Unit, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO organization_units
			(parent_id, type, name, location, address,
			 latitude, longitude, phone, email, leader_employee_id)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING `+unitColumns,
		in.ParentID, in.Type, in.Name, in.Location, in.Address,
		in.Latitude, in.Longitude, in.Phone, in.Email, in.LeaderEmployeeID,
	)

	u, err := scanUnit(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return u, nil
}

// Update обновляет организационную единицу по id.
//
// Если единица не найдена, возвращается ErrNotFound.
func (r *Repository) Update(ctx context.Context, id int64, in Input) (*Unit, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE organization_units SET
			parent_id = $1,
			type = $2,
			name = $3,
			location = $4,
			address = $5,
			latitude = $6,
			longitude = $7,
			phone = $8,
			email = $9,
			leader_employee_id = $10
		WHERE id = $11
		RETURNING `+unitColumns,
		in.ParentID, in.Type, in.Name, in.Location, in.Address,
		in.Latitude, in.Longitude, in.Phone, in.Email, in.LeaderEmployeeID,
		id,
	)

	u, err := scanUnit(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return u, nil
}

// wrapWriteError превращает ошибки Postgres (триггер иерархии, внешние
// ключи, CHECK-ограничения) в ValidationError с понятным сообщением,
// остальные ошибки оборачиваются как внутренние.
func wrapWriteError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "P0001": // raise_exception — сработал триггер иерархии
			return &ValidationError{Message: pgErr.Message}
		case "23503": // foreign_key_violation
			return &ValidationError{
				Message: "указан несуществующий parent_id или leader_employee_id",
			}
		case "23514": // check_violation (например, недопустимый type)
			return &ValidationError{Message: pgErr.Message}
		}
	}

	return fmt.Errorf("organization: %s: %w", op, err)
}
