package objecttype

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда тип объекта не найден.
var ErrNotFound = errors.New("object type not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть (дублирующееся наименование, попытка удалить используемый тип).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к object_types в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const typeColumns = `id, name, created_at`

func scanType(row pgx.Row) (*Type, error) {
	var t Type

	if err := row.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
		return nil, err
	}

	return &t, nil
}

// List возвращает все типы объектов, упорядоченные по наименованию.
func (r *Repository) List(ctx context.Context) ([]Type, error) {
	rows, err := r.db.Query(ctx, `SELECT `+typeColumns+`
		FROM object_types
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("objecttype: list: %w", err)
	}
	defer rows.Close()

	types := make([]Type, 0)

	for rows.Next() {
		t, err := scanType(rows)
		if err != nil {
			return nil, fmt.Errorf("objecttype: list: scan: %w", err)
		}

		types = append(types, *t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("objecttype: list: %w", err)
	}

	return types, nil
}

// Get возвращает тип объекта по id.
//
// Если тип не найден, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*Type, error) {
	row := r.db.QueryRow(ctx, `SELECT `+typeColumns+`
		FROM object_types
		WHERE id = $1`, id)

	t, err := scanType(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("objecttype: get: %w", err)
	}

	return t, nil
}

// Create создаёт тип объекта.
//
// Если тип с таким наименованием уже существует, возвращается
// *ValidationError.
func (r *Repository) Create(ctx context.Context, name string) (*Type, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO object_types (name)
		VALUES ($1)
		RETURNING `+typeColumns,
		name,
	)

	t, err := scanType(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return t, nil
}

// Update переименовывает тип объекта по id.
//
// Если тип не найден, возвращается ErrNotFound. Если наименование уже
// занято другим типом, возвращается *ValidationError.
func (r *Repository) Update(ctx context.Context, id int64, name string) (*Type, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE object_types SET name = $1
		WHERE id = $2
		RETURNING `+typeColumns,
		name, id,
	)

	t, err := scanType(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return t, nil
}

// Delete удаляет тип объекта по id.
//
// Если тип не найден, возвращается ErrNotFound. Если тип используется в
// назначениях (object_purposes) или объектах эксплуатации, возвращается
// *ValidationError.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM object_types WHERE id = $1`, id)
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
			return &ValidationError{Message: "тип объекта с таким наименованием уже существует"}
		case "23503": // foreign_key_violation
			return &ValidationError{
				Message: "тип объекта используется в назначениях или объектах эксплуатации и не может быть удалён",
			}
		}
	}

	return fmt.Errorf("objecttype: %s: %w", op, err)
}
