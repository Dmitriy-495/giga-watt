// Package objecttype реализует справочник типов объектов эксплуатации
// Giga Watt (Коммунальный объект, Объект ЖФ, Объект КЖФ и т.д.).
//
// Список не является закрытым — пользователь может добавлять и
// редактировать типы объектов (см. ITERATION.md).
package objecttype

import "time"

// Type — тип объекта эксплуатации из справочника.
type Type struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
