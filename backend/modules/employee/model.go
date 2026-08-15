// Package employee реализует сотрудников Giga Watt: базовые данные,
// телефоны/e-mail (1:N с признаком основного контакта) и кадровые
// назначения (связь с организационной единицей и должностью).
package employee

import "time"

// Допустимые значения пола сотрудника (см. employee_gender).
const (
	GenderMale   = "male"
	GenderFemale = "female"
)

// Допустимые типы кадрового назначения (см. chk_employee_assignments_type).
const (
	AssignmentPrimary           = "primary"
	AssignmentPartTime          = "part_time"
	AssignmentTemporaryTransfer = "temporary_transfer"
	AssignmentCombination       = "combination"
)

// Employee — сотрудник.
//
// ShortName формируется автоматически из фамилии и инициалов
// (см. shortname.go) и не задаётся клиентом напрямую.
type Employee struct {
	ID         int64      `json:"id"`
	LastName   string     `json:"last_name"`
	FirstName  string     `json:"first_name"`
	MiddleName string     `json:"middle_name"`
	ShortName  string     `json:"short_name"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Gender     *string    `json:"gender,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Phone — телефон сотрудника.
type Phone struct {
	ID         int64     `json:"id"`
	EmployeeID int64     `json:"employee_id"`
	Phone      string    `json:"phone"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
}

// Email — e-mail сотрудника.
type Email struct {
	ID         int64     `json:"id"`
	EmployeeID int64     `json:"employee_id"`
	Email      string    `json:"email"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
}

// Assignment — кадровое назначение сотрудника (связь с организационной
// единицей и должностью).
type Assignment struct {
	ID                 int64      `json:"id"`
	EmployeeID         int64      `json:"employee_id"`
	OrganizationUnitID int64      `json:"organization_unit_id"`
	PositionID         int64      `json:"position_id"`
	AssignmentType     string     `json:"assignment_type"`
	StartsAt           time.Time  `json:"starts_at"`
	EndsAt             *time.Time `json:"ends_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// Detail — сотрудник вместе со связанными контактами и назначениями.
// Используется для ответа на GET /api/employees/{id}.
type Detail struct {
	Employee
	Phones      []Phone      `json:"phones"`
	Emails      []Email      `json:"emails"`
	Assignments []Assignment `json:"assignments"`
}
