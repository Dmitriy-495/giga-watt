package position

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда должность не найдена.
var ErrNotFound = errors.New("position not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть (дублирующееся наименование, попытка удалить используемую
// должность).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к positions в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const positionColumns = `id, name, created_at`

func scanPosition(row pgx.Row) (*Position, error) {
	var p Position

	if err := row.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
		return nil, err
	}

	return &p, nil
}

// List возвращает все должности, упорядоченные по наименованию.
func (r *Repository) List(ctx context.Context) ([]Position, error) {
	rows, err := r.db.Query(ctx, `SELECT `+positionColumns+`
		FROM positions
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("position: list: %w", err)
	}
	defer rows.Close()

	positions := make([]Position, 0)

	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, fmt.Errorf("position: list: scan: %w", err)
		}

		positions = append(positions, *p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("position: list: %w", err)
	}

	return positions, nil
}

// Get возвращает должность по id.
//
// Если должность не найдена, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*Position, error) {
	row := r.db.QueryRow(ctx, `SELECT `+positionColumns+`
		FROM positions
		WHERE id = $1`, id)

	p, err := scanPosition(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("position: get: %w", err)
	}

	return p, nil
}

// Create создаёт должность.
//
// Если должность с таким наименованием уже существует, возвращается
// *ValidationError.
func (r *Repository) Create(ctx context.Context, name string) (*Position, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO positions (name)
		VALUES ($1)
		RETURNING `+positionColumns,
		name,
	)

	p, err := scanPosition(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return p, nil
}

// Update переименовывает должность по id.
//
// Если должность не найдена, возвращается ErrNotFound. Если наименование
// уже занято другой должностью, возвращается *ValidationError.
func (r *Repository) Update(ctx context.Context, id int64, name string) (*Position, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE positions SET name = $1
		WHERE id = $2
		RETURNING `+positionColumns,
		name, id,
	)

	p, err := scanPosition(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return p, nil
}

// Delete удаляет должность по id.
//
// Если должность не найдена, возвращается ErrNotFound. Если должность
// используется в кадровых назначениях или штатных единицах, возвращается
// *ValidationError.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM positions WHERE id = $1`, id)
	if err != nil {
		return wrapWriteError("delete", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// wrapWriteError превращает ошибки Postgres (уникальность имени, внешние
// ключи) в ValidationError с понятным сообщением, остальные ошибки
// оборачиваются как внутренние.
func wrapWriteError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return &ValidationError{Message: "должность с таким наименованием уже существует"}
		case "23503": // foreign_key_violation
			return &ValidationError{
				Message: "должность используется в назначениях или штатных единицах и не может быть удалена",
			}
		}
	}

	return fmt.Errorf("position: %s: %w", op, err)
}
