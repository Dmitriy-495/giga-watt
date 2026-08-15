// Package operationalobject реализует объекты эксплуатации Giga Watt.
//
// Здание, сооружение, сеть, котельная, скважина и т.д. не выделяются в
// отдельные сущности — все они являются объектами эксплуатации, а их
// конкретная природа отражается наименованием, типом и назначением (см.
// ITERATION.md).
//
// Объект эксплуатации принадлежит только ПУ — это проверяется триггером
// БД (validate_operational_object_owner), а не в этом пакете.
package operationalobject

import "time"

// Object — объект эксплуатации.
type Object struct {
	ID                 int64     `json:"id"`
	OrganizationUnitID int64     `json:"organization_unit_id"`
	ObjectTypeID       int64     `json:"object_type_id"`
	ObjectPurposeID    *int64    `json:"object_purpose_id,omitempty"`
	Name               string    `json:"name"`
	Address            *string   `json:"address,omitempty"`
	Latitude           *float64  `json:"latitude,omitempty"`
	Longitude          *float64  `json:"longitude,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}
