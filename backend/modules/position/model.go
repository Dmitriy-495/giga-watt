// Package position реализует справочник должностей Giga Watt.
//
// Должность — отдельный редактируемый справочник (см.
// docs/iterations/001-foundation/ITERATION.md). Сотрудник не хранит
// название должности напрямую — оно приходит через кадровое назначение.
package position

import "time"

// Position — должность из справочника.
type Position struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
