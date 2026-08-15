package objectpurpose

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда назначение объекта не найдено.
var ErrNotFound = errors.New("object purpose not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть (несуществующий object_type_id, дублирующееся наименование в
// рамках типа, попытка удалить используемое назначение).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к object_purposes в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const purposeColumns = `id, object_type_id, name, created_at`

func scanPurpose(row pgx.Row) (*Purpose, error) {
	var p Purpose

	if err := row.Scan(&p.ID, &p.ObjectTypeID, &p.Name, &p.CreatedAt); err != nil {
		return nil, err
	}

	return &p, nil
}

// List возвращает назначения объектов, упорядоченные по наименованию.
//
// Если objectTypeID > 0, список ограничивается этим типом объекта (см.
// ITERATION.md: назначение связано с типом объекта).
func (r *Repository) List(ctx context.Context, objectTypeID int64) ([]Purpose, error) {
	var rows pgx.Rows
	var err error

	if objectTypeID > 0 {
		rows, err = r.db.Query(ctx, `SELECT `+purposeColumns+`
			FROM object_purposes
			WHERE object_type_id = $1
			ORDER BY name`, objectTypeID)
	} else {
		rows, err = r.db.Query(ctx, `SELECT `+purposeColumns+`
			FROM object_purposes
			ORDER BY object_type_id, name`)
	}

	if err != nil {
		return nil, fmt.Errorf("objectpurpose: list: %w", err)
	}
	defer rows.Close()

	purposes := make([]Purpose, 0)

	for rows.Next() {
		p, err := scanPurpose(rows)
		if err != nil {
			return nil, fmt.Errorf("objectpurpose: list: scan: %w", err)
		}

		purposes = append(purposes, *p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("objectpurpose: list: %w", err)
	}

	return purposes, nil
}

// Get возвращает назначение объекта по id.
//
// Если назначение не найдено, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*Purpose, error) {
	row := r.db.QueryRow(ctx, `SELECT `+purposeColumns+`
		FROM object_purposes
		WHERE id = $1`, id)

	p, err := scanPurpose(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("objectpurpose: get: %w", err)
	}

	return p, nil
}

// Create создаёт назначение объекта для указанного типа.
func (r *Repository) Create(ctx context.Context, objectTypeID int64, name string) (*Purpose, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO object_purposes (object_type_id, name)
		VALUES ($1, $2)
		RETURNING `+purposeColumns,
		objectTypeID, name,
	)

	p, err := scanPurpose(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return p, nil
}

// Update обновляет назначение объекта по id.
//
// Если назначение не найдено, возвращается ErrNotFound.
func (r *Repository) Update(ctx context.Context, id, objectTypeID int64, name string) (*Purpose, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE object_purposes SET object_type_id = $1, name = $2
		WHERE id = $3
		RETURNING `+purposeColumns,
		objectTypeID, name, id,
	)

	p, err := scanPurpose(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return p, nil
}

// Delete удаляет назначение объекта по id.
//
// Если назначение не найдено, возвращается ErrNotFound. Если назначение
// используется объектами эксплуатации, возвращается *ValidationError.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM object_purposes WHERE id = $1`, id)
	if err != nil {
		return wrapWriteError("delete", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// wrapWriteError превращает ошибки Postgres (внешние ключи, уникальность
// имени в рамках типа) в ValidationError с понятным сообщением, остальные
// ошибки оборачиваются как внутренние.
func wrapWriteError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "object_purposes_object_type_id_fkey":
			return &ValidationError{Message: "указан несуществующий object_type_id"}
		case "uq_object_purposes_type_name":
			return &ValidationError{Message: "назначение с таким наименованием уже существует для этого типа объекта"}
		}

		switch pgErr.Code {
		case "23505":
			return &ValidationError{Message: "назначение с таким наименованием уже существует для этого типа объекта"}
		case "23503":
			return &ValidationError{
				Message: "назначение используется объектами эксплуатации и не может быть удалено, либо указан несуществующий object_type_id",
			}
		}
	}

	return fmt.Errorf("objectpurpose: %s: %w", op, err)
}
