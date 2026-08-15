// Package organization реализует организационную структуру Giga Watt:
// Учреждение → Филиал → ЖКС → ПУ.
//
// Иерархия жёсткая (1:N на каждом уровне) и проверяется триггером
// в PostgreSQL (см. migrations/000003_organization_hierarchy.up.sql),
// поэтому backend не дублирует эту проверку, а полагается на БД.
package organization

import "time"

// Допустимые типы организационных единиц (см. chk_organization_units_type).
const (
	TypeInstitution    = "institution"
	TypeBranch         = "branch"
	TypeJKS            = "jks"
	TypeProductionUnit = "production_unit"
)

// Unit — организационная единица (Учреждение, Филиал, ЖКС или ПУ).
type Unit struct {
	ID               int64     `json:"id"`
	ParentID         *int64    `json:"parent_id,omitempty"`
	Type             string    `json:"type"`
	Name             string    `json:"name"`
	Location         string    `json:"location"`
	Address          string    `json:"address"`
	Latitude         *float64  `json:"latitude,omitempty"`
	Longitude        *float64  `json:"longitude,omitempty"`
	Phone            *string   `json:"phone,omitempty"`
	Email            *string   `json:"email,omitempty"`
	LeaderEmployeeID *int64    `json:"leader_employee_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
