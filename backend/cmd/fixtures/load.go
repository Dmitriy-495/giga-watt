package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"

	"github.com/Dmitriy-495/giga-watt/backend/modules/employee"
)

// orgTypeDB переводит человеческую подпись уровня иерархии (как её
// заполняет человек в Excel) в код типа, который ожидает БД.
var orgTypeDB = map[string]string{
	"Учреждение": "institution",
	"Филиал":     "branch",
	"ЖКС":        "jks",
	"ПУ":         "production_unit",
}

// orgTypeLoadOrder — порядок загрузки уровней иерархии: родитель всегда
// должен существовать раньше потомка (см. триггер
// validate_organization_unit_hierarchy).
var orgTypeLoadOrder = []string{"institution", "branch", "jks", "production_unit"}

// assignmentTypeDB переводит человеческую подпись типа кадрового
// назначения в код, который ожидает БД (см. chk_employee_assignments_type).
var assignmentTypeDB = map[string]string{
	"основное":          "primary",
	"совместительство":  "part_time",
	"временный перевод": "temporary_transfer",
	"совмещение":        "combination",
}

// loader держит транзакцию и карты "бизнес-код → id в БД", которые
// заполняются по мере загрузки листов и используются для разрешения
// ссылок между ними (см. AGENTS.md, п. 9: технические id генерируются
// при загрузке, а не хранятся в Excel).
type loader struct {
	ctx context.Context
	tx  pgx.Tx

	objectTypeID    map[string]int64
	objectPurposeID map[string]int64
	orgUnitID       map[string]int64
	positionID      map[string]int64
	employeeID      map[string]int64

	// pendingLeaders — строки organization с непустым
	// leader_employee_code, обрабатываются вторым проходом после
	// загрузки employees (руководитель — сотрудник, а сотрудники грузятся
	// после организационной структуры).
	pendingLeaders []orgRow

	counts map[string]int
}

func newLoader(ctx context.Context, tx pgx.Tx) *loader {
	return &loader{
		ctx:             ctx,
		tx:              tx,
		objectTypeID:    map[string]int64{},
		objectPurposeID: map[string]int64{},
		orgUnitID:       map[string]int64{},
		positionID:      map[string]int64{},
		employeeID:      map[string]int64{},
		counts:          map[string]int{},
	}
}

// load читает все листы книги и загружает их в БД в порядке,
// продиктованном внешними ключами схемы (см.
// docs/DEVELOPMENT.md, раздел "Fixtures").
func load(ctx context.Context, tx pgx.Tx, f *excelize.File) error {
	l := newLoader(ctx, tx)

	steps := []struct {
		sheet    string
		required []string
		fn       func(*sheet) error
	}{
		{"object_types", []string{"code", "name"}, l.loadObjectTypes},
		{"object_purposes", []string{"code", "object_type_code", "name"}, l.loadObjectPurposes},
		{"positions", []string{"code", "name"}, l.loadPositions},
		{"organization", []string{"code", "type", "parent_code", "name", "location", "address"}, l.loadOrganization},
		{"employees", []string{"code", "last_name", "first_name", "middle_name"}, l.loadEmployees},
		{"employee_assignments", []string{"employee_code", "organization_code", "position_code", "assignment_type", "starts_at"}, l.loadEmployeeAssignments},
		{"employee_phones", []string{"employee_code", "phone"}, l.loadEmployeePhones},
		{"employee_emails", []string{"employee_code", "email"}, l.loadEmployeeEmails},
		{"staff_positions", []string{"organization_code", "position_code", "quantity"}, l.loadStaffPositions},
		{"operational_objects", []string{"code", "organization_code", "object_type_code", "name"}, l.loadOperationalObjects},
	}

	for _, step := range steps {
		s, err := readSheet(f, step.sheet, step.required)
		if err != nil {
			return err
		}

		if err := step.fn(s); err != nil {
			return err
		}

		l.counts[step.sheet] = len(s.rows)
	}

	if err := l.applyOrganizationLeaders(); err != nil {
		return err
	}

	printSummary(l.counts)

	return nil
}

func printSummary(counts map[string]int) {
	fmt.Println("fixtures: загружено строк по листам:")

	order := []string{
		"object_types", "object_purposes", "positions", "organization",
		"employees", "employee_assignments", "employee_phones",
		"employee_emails", "staff_positions", "operational_objects",
	}

	for _, name := range order {
		fmt.Printf("  %-22s %d\n", name, counts[name])
	}
}

func (l *loader) loadObjectTypes(s *sheet) error {
	for i, row := range s.rows {
		code := s.get(row, "code")
		name := s.get(row, "name")

		if code == "" {
			return fmt.Errorf("object_types строка %d: пустой code", s.rowNumber(i))
		}

		if name == "" {
			return fmt.Errorf("object_types строка %d (%s): пустой name", s.rowNumber(i), code)
		}

		if _, dup := l.objectTypeID[code]; dup {
			return fmt.Errorf("object_types строка %d: дублирующийся code %q", s.rowNumber(i), code)
		}

		var id int64

		err := l.tx.QueryRow(l.ctx, `
			INSERT INTO object_types (name) VALUES ($1) RETURNING id`, name,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("object_types строка %d (%s): %w", s.rowNumber(i), code, err)
		}

		l.objectTypeID[code] = id
	}

	return nil
}

func (l *loader) loadObjectPurposes(s *sheet) error {
	for i, row := range s.rows {
		code := s.get(row, "code")
		typeCode := s.get(row, "object_type_code")
		name := s.get(row, "name")

		if code == "" {
			return fmt.Errorf("object_purposes строка %d: пустой code", s.rowNumber(i))
		}

		if name == "" {
			return fmt.Errorf("object_purposes строка %d (%s): пустой name", s.rowNumber(i), code)
		}

		if _, dup := l.objectPurposeID[code]; dup {
			return fmt.Errorf("object_purposes строка %d: дублирующийся code %q", s.rowNumber(i), code)
		}

		typeID, ok := l.objectTypeID[typeCode]
		if !ok {
			return fmt.Errorf("object_purposes строка %d (%s): неизвестный object_type_code %q",
				s.rowNumber(i), code, typeCode)
		}

		var id int64

		err := l.tx.QueryRow(l.ctx, `
			INSERT INTO object_purposes (object_type_id, name) VALUES ($1, $2) RETURNING id`,
			typeID, name,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("object_purposes строка %d (%s): %w", s.rowNumber(i), code, err)
		}

		l.objectPurposeID[code] = id
	}

	return nil
}

func (l *loader) loadPositions(s *sheet) error {
	for i, row := range s.rows {
		code := s.get(row, "code")
		name := s.get(row, "name")

		if code == "" {
			return fmt.Errorf("positions строка %d: пустой code", s.rowNumber(i))
		}

		if name == "" {
			return fmt.Errorf("positions строка %d (%s): пустой name", s.rowNumber(i), code)
		}

		if _, dup := l.positionID[code]; dup {
			return fmt.Errorf("positions строка %d: дублирующийся code %q", s.rowNumber(i), code)
		}

		var id int64

		err := l.tx.QueryRow(l.ctx, `
			INSERT INTO positions (name) VALUES ($1) RETURNING id`, name,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("positions строка %d (%s): %w", s.rowNumber(i), code, err)
		}

		l.positionID[code] = id
	}

	return nil
}

// orgRow — разобранная (провалидированная) строка листа organization,
// ожидающая вставки в БД в правильном порядке по уровню иерархии.
type orgRow struct {
	rowNum     int
	code       string
	dbType     string
	parentCode string
	name       string
	location   string
	address    string
	latitude   *float64
	longitude  *float64
	phone      *string
	email      *string
	leaderCode string
}

func (l *loader) loadOrganization(s *sheet) error {
	parsed := make([]orgRow, 0, len(s.rows))
	seen := make(map[string]bool, len(s.rows))

	for i, row := range s.rows {
		code := s.get(row, "code")
		label := s.get(row, "type")

		if code == "" {
			return fmt.Errorf("organization строка %d: пустой code", s.rowNumber(i))
		}

		if seen[code] {
			return fmt.Errorf("organization строка %d: дублирующийся code %q", s.rowNumber(i), code)
		}

		seen[code] = true

		dbType, ok := orgTypeDB[label]
		if !ok {
			return fmt.Errorf(
				"organization строка %d (%s): неизвестный type %q (ожидается Учреждение/Филиал/ЖКС/ПУ)",
				s.rowNumber(i), code, label,
			)
		}

		name := s.get(row, "name")
		location := s.get(row, "location")
		address := s.get(row, "address")

		if name == "" || location == "" || address == "" {
			return fmt.Errorf("organization строка %d (%s): name/location/address обязательны",
				s.rowNumber(i), code)
		}

		lat, err := parseFloatPtr(s.get(row, "latitude"))
		if err != nil {
			return fmt.Errorf("organization строка %d (%s): latitude: %w", s.rowNumber(i), code, err)
		}

		lon, err := parseFloatPtr(s.get(row, "longitude"))
		if err != nil {
			return fmt.Errorf("organization строка %d (%s): longitude: %w", s.rowNumber(i), code, err)
		}

		parsed = append(parsed, orgRow{
			rowNum:     s.rowNumber(i),
			code:       code,
			dbType:     dbType,
			parentCode: s.get(row, "parent_code"),
			name:       name,
			location:   location,
			address:    address,
			latitude:   lat,
			longitude:  lon,
			phone:      strPtr(s.get(row, "phone")),
			email:      strPtr(s.get(row, "email")),
			leaderCode: s.get(row, "leader_employee_code"),
		})
	}

	for _, dbType := range orgTypeLoadOrder {
		for _, r := range parsed {
			if r.dbType != dbType {
				continue
			}

			var parentID *int64

			if r.parentCode != "" {
				pid, ok := l.orgUnitID[r.parentCode]
				if !ok {
					return fmt.Errorf("organization строка %d (%s): неизвестный parent_code %q",
						r.rowNum, r.code, r.parentCode)
				}

				parentID = &pid
			}

			var id int64

			err := l.tx.QueryRow(l.ctx, `
				INSERT INTO organization_units
					(type, parent_id, name, location, address, latitude, longitude, phone, email)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id`,
				r.dbType, parentID, r.name, r.location, r.address,
				r.latitude, r.longitude, r.phone, r.email,
			).Scan(&id)
			if err != nil {
				return fmt.Errorf("organization строка %d (%s): %w", r.rowNum, r.code, err)
			}

			l.orgUnitID[r.code] = id
		}
	}

	l.pendingLeaders = parsed

	return nil
}

// applyOrganizationLeaders выставляет leader_employee_id для строк
// organization, у которых была указана leader_employee_code —
// отдельным проходом после загрузки employees (руководитель — это
// сотрудник, а сотрудники грузятся позже organization).
func (l *loader) applyOrganizationLeaders() error {
	for _, r := range l.pendingLeaders {
		if r.leaderCode == "" {
			continue
		}

		empID, ok := l.employeeID[r.leaderCode]
		if !ok {
			return fmt.Errorf("organization (%s): неизвестный leader_employee_code %q", r.code, r.leaderCode)
		}

		orgID := l.orgUnitID[r.code]

		if _, err := l.tx.Exec(l.ctx, `
			UPDATE organization_units SET leader_employee_id = $1 WHERE id = $2`,
			empID, orgID,
		); err != nil {
			return fmt.Errorf("organization (%s): назначение руководителя: %w", r.code, err)
		}
	}

	return nil
}

func (l *loader) loadEmployees(s *sheet) error {
	for i, row := range s.rows {
		code := s.get(row, "code")
		lastName := s.get(row, "last_name")
		firstName := s.get(row, "first_name")
		middleName := s.get(row, "middle_name")

		if code == "" {
			return fmt.Errorf("employees строка %d: пустой code", s.rowNumber(i))
		}

		if _, dup := l.employeeID[code]; dup {
			return fmt.Errorf("employees строка %d: дублирующийся code %q", s.rowNumber(i), code)
		}

		if lastName == "" || firstName == "" || middleName == "" {
			return fmt.Errorf("employees строка %d (%s): last_name/first_name/middle_name обязательны",
				s.rowNumber(i), code)
		}

		birthDate, err := parseDate(s.get(row, "birth_date"))
		if err != nil {
			return fmt.Errorf("employees строка %d (%s): birth_date: %w", s.rowNumber(i), code, err)
		}

		gender, err := parseGender(s.get(row, "gender"))
		if err != nil {
			return fmt.Errorf("employees строка %d (%s): %w", s.rowNumber(i), code, err)
		}

		shortName := employee.BuildShortName(lastName, firstName, middleName)

		var id int64

		err = l.tx.QueryRow(l.ctx, `
			INSERT INTO employees (last_name, first_name, middle_name, short_name, birth_date, gender)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`,
			lastName, firstName, middleName, shortName, birthDate, gender,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("employees строка %d (%s): %w", s.rowNumber(i), code, err)
		}

		l.employeeID[code] = id
	}

	return nil
}

func parseGender(raw string) (*string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "":
		return nil, nil
	case "М", "M", "MALE":
		v := "male"
		return &v, nil
	case "Ж", "F", "FEMALE":
		v := "female"
		return &v, nil
	default:
		return nil, fmt.Errorf("gender: неизвестное значение %q (ожидается М/Ж)", raw)
	}
}

func (l *loader) loadEmployeeAssignments(s *sheet) error {
	for i, row := range s.rows {
		empCode := s.get(row, "employee_code")
		orgCode := s.get(row, "organization_code")
		posCode := s.get(row, "position_code")
		typeLabel := s.get(row, "assignment_type")

		empID, ok := l.employeeID[empCode]
		if !ok {
			return fmt.Errorf("employee_assignments строка %d: неизвестный employee_code %q",
				s.rowNumber(i), empCode)
		}

		orgID, ok := l.orgUnitID[orgCode]
		if !ok {
			return fmt.Errorf("employee_assignments строка %d (%s): неизвестный organization_code %q",
				s.rowNumber(i), empCode, orgCode)
		}

		posID, ok := l.positionID[posCode]
		if !ok {
			return fmt.Errorf("employee_assignments строка %d (%s): неизвестный position_code %q",
				s.rowNumber(i), empCode, posCode)
		}

		dbType, ok := assignmentTypeDB[typeLabel]
		if !ok {
			return fmt.Errorf(
				"employee_assignments строка %d (%s): неизвестный assignment_type %q (ожидается основное/совместительство/временный перевод/совмещение)",
				s.rowNumber(i), empCode, typeLabel,
			)
		}

		startsAt, err := parseDate(s.get(row, "starts_at"))
		if err != nil {
			return fmt.Errorf("employee_assignments строка %d (%s): starts_at: %w", s.rowNumber(i), empCode, err)
		}

		if startsAt == nil {
			return fmt.Errorf("employee_assignments строка %d (%s): starts_at обязателен", s.rowNumber(i), empCode)
		}

		endsAt, err := parseDate(s.get(row, "ends_at"))
		if err != nil {
			return fmt.Errorf("employee_assignments строка %d (%s): ends_at: %w", s.rowNumber(i), empCode, err)
		}

		_, err = l.tx.Exec(l.ctx, `
			INSERT INTO employee_assignments
				(employee_id, organization_unit_id, position_id, assignment_type, starts_at, ends_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			empID, orgID, posID, dbType, startsAt, endsAt,
		)
		if err != nil {
			return fmt.Errorf("employee_assignments строка %d (%s): %w", s.rowNumber(i), empCode, err)
		}
	}

	return nil
}

func (l *loader) loadEmployeePhones(s *sheet) error {
	for i, row := range s.rows {
		empCode := s.get(row, "employee_code")
		phone := s.get(row, "phone")

		empID, ok := l.employeeID[empCode]
		if !ok {
			return fmt.Errorf("employee_phones строка %d: неизвестный employee_code %q", s.rowNumber(i), empCode)
		}

		if phone == "" {
			return fmt.Errorf("employee_phones строка %d (%s): пустой phone", s.rowNumber(i), empCode)
		}

		isPrimary := parseBool(s.get(row, "is_primary"))

		_, err := l.tx.Exec(l.ctx, `
			INSERT INTO employee_phones (employee_id, phone, is_primary)
			VALUES ($1, $2, $3)`,
			empID, phone, isPrimary,
		)
		if err != nil {
			return fmt.Errorf("employee_phones строка %d (%s): %w", s.rowNumber(i), empCode, err)
		}
	}

	return nil
}

func (l *loader) loadEmployeeEmails(s *sheet) error {
	for i, row := range s.rows {
		empCode := s.get(row, "employee_code")
		email := s.get(row, "email")

		empID, ok := l.employeeID[empCode]
		if !ok {
			return fmt.Errorf("employee_emails строка %d: неизвестный employee_code %q", s.rowNumber(i), empCode)
		}

		if email == "" {
			return fmt.Errorf("employee_emails строка %d (%s): пустой email", s.rowNumber(i), empCode)
		}

		isPrimary := parseBool(s.get(row, "is_primary"))

		_, err := l.tx.Exec(l.ctx, `
			INSERT INTO employee_emails (employee_id, email, is_primary)
			VALUES ($1, $2, $3)`,
			empID, email, isPrimary,
		)
		if err != nil {
			return fmt.Errorf("employee_emails строка %d (%s): %w", s.rowNumber(i), empCode, err)
		}
	}

	return nil
}

func (l *loader) loadStaffPositions(s *sheet) error {
	for i, row := range s.rows {
		orgCode := s.get(row, "organization_code")
		posCode := s.get(row, "position_code")
		qtyRaw := s.get(row, "quantity")

		orgID, ok := l.orgUnitID[orgCode]
		if !ok {
			return fmt.Errorf("staff_positions строка %d: неизвестный organization_code %q", s.rowNumber(i), orgCode)
		}

		posID, ok := l.positionID[posCode]
		if !ok {
			return fmt.Errorf("staff_positions строка %d: неизвестный position_code %q", s.rowNumber(i), posCode)
		}

		qty, err := strconv.ParseFloat(qtyRaw, 64)
		if err != nil {
			return fmt.Errorf("staff_positions строка %d: quantity: ожидалось число, получено %q",
				s.rowNumber(i), qtyRaw)
		}

		_, err = l.tx.Exec(l.ctx, `
			INSERT INTO staff_positions (organization_unit_id, position_id, quantity)
			VALUES ($1, $2, $3)`,
			orgID, posID, qty,
		)
		if err != nil {
			return fmt.Errorf("staff_positions строка %d (%s/%s): %w", s.rowNumber(i), orgCode, posCode, err)
		}
	}

	return nil
}

func (l *loader) loadOperationalObjects(s *sheet) error {
	for i, row := range s.rows {
		code := s.get(row, "code")
		orgCode := s.get(row, "organization_code")
		typeCode := s.get(row, "object_type_code")
		purposeCode := s.get(row, "object_purpose_code")
		name := s.get(row, "name")

		if code == "" {
			return fmt.Errorf("operational_objects строка %d: пустой code", s.rowNumber(i))
		}

		if name == "" {
			return fmt.Errorf("operational_objects строка %d (%s): пустой name", s.rowNumber(i), code)
		}

		orgID, ok := l.orgUnitID[orgCode]
		if !ok {
			return fmt.Errorf("operational_objects строка %d (%s): неизвестный organization_code %q",
				s.rowNumber(i), code, orgCode)
		}

		typeID, ok := l.objectTypeID[typeCode]
		if !ok {
			return fmt.Errorf("operational_objects строка %d (%s): неизвестный object_type_code %q",
				s.rowNumber(i), code, typeCode)
		}

		var purposeID *int64

		if purposeCode != "" {
			pid, ok := l.objectPurposeID[purposeCode]
			if !ok {
				return fmt.Errorf("operational_objects строка %d (%s): неизвестный object_purpose_code %q",
					s.rowNumber(i), code, purposeCode)
			}

			purposeID = &pid
		}

		address := strPtr(s.get(row, "address"))

		lat, err := parseFloatPtr(s.get(row, "latitude"))
		if err != nil {
			return fmt.Errorf("operational_objects строка %d (%s): latitude: %w", s.rowNumber(i), code, err)
		}

		lon, err := parseFloatPtr(s.get(row, "longitude"))
		if err != nil {
			return fmt.Errorf("operational_objects строка %d (%s): longitude: %w", s.rowNumber(i), code, err)
		}

		_, err = l.tx.Exec(l.ctx, `
			INSERT INTO operational_objects
				(organization_unit_id, object_type_id, object_purpose_id, name, address, latitude, longitude)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			orgID, typeID, purposeID, name, address, lat, lon,
		)
		if err != nil {
			return fmt.Errorf("operational_objects строка %d (%s): %w", s.rowNumber(i), code, err)
		}
	}

	return nil
}
