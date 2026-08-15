// Package staffposition реализует штатные единицы Giga Watt — связь
// организационной единицы, должности и количества единиц.
//
// На начальном этапе сотрудники могут существовать без привязки к
// конкретной штатной единице (см. ITERATION.md).
package staffposition

import "time"

// StaffPosition — штатная единица.
type StaffPosition struct {
	ID                 int64     `json:"id"`
	OrganizationUnitID int64     `json:"organization_unit_id"`
	PositionID         int64     `json:"position_id"`
	Quantity           float64   `json:"quantity"`
	CreatedAt          time.Time `json:"created_at"`
}
