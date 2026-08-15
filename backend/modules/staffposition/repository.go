package staffposition

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда штатная единица не найдена.
var ErrNotFound = errors.New("staff position not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть (несуществующие organization_unit_id/position_id, недопустимое
// количество единиц).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к staff_positions в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const staffPositionColumns = `id, organization_unit_id, position_id, quantity, created_at`

func scanStaffPosition(row pgx.Row) (*StaffPosition, error) {
	var s StaffPosition

	err := row.Scan(&s.ID, &s.OrganizationUnitID, &s.PositionID, &s.Quantity, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// List возвращает штатные единицы.
//
// Если organizationUnitID > 0, список ограничивается этой
// организационной единицей.
func (r *Repository) List(ctx context.Context, organizationUnitID int64) ([]StaffPosition, error) {
	var rows pgx.Rows
	var err error

	if organizationUnitID > 0 {
		rows, err = r.db.Query(ctx, `SELECT `+staffPositionColumns+`
			FROM staff_positions
			WHERE organization_unit_id = $1
			ORDER BY id`, organizationUnitID)
	} else {
		rows, err = r.db.Query(ctx, `SELECT `+staffPositionColumns+`
			FROM staff_positions
			ORDER BY id`)
	}

	if err != nil {
		return nil, fmt.Errorf("staffposition: list: %w", err)
	}
	defer rows.Close()

	items := make([]StaffPosition, 0)

	for rows.Next() {
		s, err := scanStaffPosition(rows)
		if err != nil {
			return nil, fmt.Errorf("staffposition: list: scan: %w", err)
		}

		items = append(items, *s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("staffposition: list: %w", err)
	}

	return items, nil
}

// Get возвращает штатную единицу по id.
//
// Если штатная единица не найдена, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*StaffPosition, error) {
	row := r.db.QueryRow(ctx, `SELECT `+staffPositionColumns+`
		FROM staff_positions
		WHERE id = $1`, id)

	s, err := scanStaffPosition(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("staffposition: get: %w", err)
	}

	return s, nil
}

// Input — данные для создания или обновления штатной единицы.
type Input struct {
	OrganizationUnitID int64
	PositionID         int64
	Quantity           float64
}

// Create создаёт штатную единицу.
func (r *Repository) Create(ctx context.Context, in Input) (*StaffPosition, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO staff_positions (organization_unit_id, position_id, quantity)
		VALUES ($1, $2, $3)
		RETURNING `+staffPositionColumns,
		in.OrganizationUnitID, in.PositionID, in.Quantity,
	)

	s, err := scanStaffPosition(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return s, nil
}

// Update обновляет штатную единицу по id.
//
// Если штатная единица не найдена, возвращается ErrNotFound.
func (r *Repository) Update(ctx context.Context, id int64, in Input) (*StaffPosition, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE staff_positions SET
			organization_unit_id = $1,
			position_id = $2,
			quantity = $3
		WHERE id = $4
		RETURNING `+staffPositionColumns,
		in.OrganizationUnitID, in.PositionID, in.Quantity, id,
	)

	s, err := scanStaffPosition(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return s, nil
}

// Delete удаляет штатную единицу по id.
//
// Если штатная единица не найдена, возвращается ErrNotFound.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM staff_positions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("staffposition: delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// wrapWriteError превращает ошибки Postgres (внешние ключи,
// CHECK-ограничение на количество) в ValidationError с понятным
// сообщением, остальные ошибки оборачиваются как внутренние.
func wrapWriteError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "staff_positions_organization_unit_id_fkey":
			return &ValidationError{Message: "указан несуществующий organization_unit_id"}
		case "staff_positions_position_id_fkey":
			return &ValidationError{Message: "указан несуществующий position_id"}
		case "chk_staff_positions_quantity":
			return &ValidationError{Message: "quantity: количество единиц должно быть больше нуля"}
		}

		switch pgErr.Code {
		case "23503":
			return &ValidationError{Message: "указан несуществующий organization_unit_id или position_id"}
		case "23514":
			return &ValidationError{Message: pgErr.Message}
		}
	}

	return fmt.Errorf("staffposition: %s: %w", op, err)
}
