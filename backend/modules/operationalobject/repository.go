package operationalobject

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound возвращается, когда объект эксплуатации не найден.
var ErrNotFound = errors.New("operational object not found")

// ValidationError — предметная ошибка, которую нужно вернуть клиенту как
// есть: организационная единица не является ПУ, несуществующий тип или
// назначение, назначение не соответствует типу объекта.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Repository — доступ к operational_objects в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создаёт Repository поверх пула соединений pgx.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const objectColumns = `
	id, organization_unit_id, object_type_id, object_purpose_id,
	name, address, latitude, longitude, created_at`

func scanObject(row pgx.Row) (*Object, error) {
	var o Object

	err := row.Scan(
		&o.ID, &o.OrganizationUnitID, &o.ObjectTypeID, &o.ObjectPurposeID,
		&o.Name, &o.Address, &o.Latitude, &o.Longitude, &o.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

// List возвращает объекты эксплуатации.
//
// Если organizationUnitID > 0, список ограничивается этим ПУ.
func (r *Repository) List(ctx context.Context, organizationUnitID int64) ([]Object, error) {
	var rows pgx.Rows
	var err error

	if organizationUnitID > 0 {
		rows, err = r.db.Query(ctx, `SELECT `+objectColumns+`
			FROM operational_objects
			WHERE organization_unit_id = $1
			ORDER BY name`, organizationUnitID)
	} else {
		rows, err = r.db.Query(ctx, `SELECT `+objectColumns+`
			FROM operational_objects
			ORDER BY name`)
	}

	if err != nil {
		return nil, fmt.Errorf("operationalobject: list: %w", err)
	}
	defer rows.Close()

	objects := make([]Object, 0)

	for rows.Next() {
		o, err := scanObject(rows)
		if err != nil {
			return nil, fmt.Errorf("operationalobject: list: scan: %w", err)
		}

		objects = append(objects, *o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("operationalobject: list: %w", err)
	}

	return objects, nil
}

// Get возвращает объект эксплуатации по id.
//
// Если объект не найден, возвращается ErrNotFound.
func (r *Repository) Get(ctx context.Context, id int64) (*Object, error) {
	row := r.db.QueryRow(ctx, `SELECT `+objectColumns+`
		FROM operational_objects
		WHERE id = $1`, id)

	o, err := scanObject(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("operationalobject: get: %w", err)
	}

	return o, nil
}

// Input — данные для создания или обновления объекта эксплуатации.
type Input struct {
	OrganizationUnitID int64
	ObjectTypeID       int64
	ObjectPurposeID    *int64
	Name               string
	Address            *string
	Latitude           *float64
	Longitude          *float64
}

// Create создаёт объект эксплуатации.
//
// organization_unit_id должен ссылаться на ПУ (проверяется триггером
// validate_operational_object_owner), а object_purpose_id — совпадать
// по типу с object_type_id (проверяется составным внешним ключом
// fk_operational_objects_type_purpose). При нарушении вернётся
// *ValidationError.
func (r *Repository) Create(ctx context.Context, in Input) (*Object, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO operational_objects
			(organization_unit_id, object_type_id, object_purpose_id,
			 name, address, latitude, longitude)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+objectColumns,
		in.OrganizationUnitID, in.ObjectTypeID, in.ObjectPurposeID,
		in.Name, in.Address, in.Latitude, in.Longitude,
	)

	o, err := scanObject(row)
	if err != nil {
		return nil, wrapWriteError("create", err)
	}

	return o, nil
}

// Update обновляет объект эксплуатации по id.
//
// Если объект не найден, возвращается ErrNotFound.
func (r *Repository) Update(ctx context.Context, id int64, in Input) (*Object, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE operational_objects SET
			organization_unit_id = $1,
			object_type_id = $2,
			object_purpose_id = $3,
			name = $4,
			address = $5,
			latitude = $6,
			longitude = $7
		WHERE id = $8
		RETURNING `+objectColumns,
		in.OrganizationUnitID, in.ObjectTypeID, in.ObjectPurposeID,
		in.Name, in.Address, in.Latitude, in.Longitude,
		id,
	)

	o, err := scanObject(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, wrapWriteError("update", err)
	}

	return o, nil
}

// Delete удаляет объект эксплуатации по id.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM operational_objects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("operationalobject: delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// wrapWriteError превращает ошибки Postgres (триггер владения ПУ,
// внешние ключи типа/назначения, несоответствие назначения типу) в
// ValidationError с понятным сообщением, остальные ошибки оборачиваются
// как внутренние.
func wrapWriteError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "P0001": // raise_exception — сработал триггер владения ПУ
			return &ValidationError{Message: pgErr.Message}
		}

		switch pgErr.ConstraintName {
		case "operational_objects_organization_unit_id_fkey":
			return &ValidationError{Message: "указан несуществующий organization_unit_id"}
		case "operational_objects_object_type_id_fkey":
			return &ValidationError{Message: "указан несуществующий object_type_id"}
		case "operational_objects_object_purpose_id_fkey":
			return &ValidationError{Message: "указан несуществующий object_purpose_id"}
		case "fk_operational_objects_type_purpose":
			return &ValidationError{Message: "object_purpose_id не соответствует указанному object_type_id"}
		}

		switch pgErr.Code {
		case "23503":
			return &ValidationError{Message: "указана несуществующая ссылка (organization_unit_id/object_type_id/object_purpose_id)"}
		}
	}

	return fmt.Errorf("operationalobject: %s: %w", op, err)
}
