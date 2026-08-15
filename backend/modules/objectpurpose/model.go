// Package objectpurpose реализует справочник назначений объектов
// эксплуатации Giga Watt (котельная, скважина, тепловые сети и т.д.).
//
// Каждое назначение связано с типом объекта (object_type_id), что
// ограничивает выбор назначений при создании/редактировании объекта
// выбранным типом (см. ITERATION.md).
package objectpurpose

import "time"

// Purpose — назначение объекта эксплуатации из справочника.
type Purpose struct {
	ID           int64     `json:"id"`
	ObjectTypeID int64     `json:"object_type_id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}
