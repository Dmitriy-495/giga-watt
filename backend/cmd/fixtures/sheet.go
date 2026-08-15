package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// sheet — обёртка над листом Excel: заголовок разобран в карту
// имя-колонки → индекс, чтобы порядок колонок в файле не имел значения.
type sheet struct {
	name string
	rows [][]string
	col  map[string]int
}

// readSheet читает лист по имени и проверяет наличие обязательных
// колонок. Отсутствие листа или обязательной колонки — явная,
// диагностируемая ошибка (см. AGENTS.md: loader должен понятно
// сообщать о проблеме, а не падать неочевидно).
func readSheet(f *excelize.File, name string, required []string) (*sheet, error) {
	rows, err := f.GetRows(name)
	if err != nil {
		return nil, fmt.Errorf("лист %q: %w", name, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("лист %q пуст (нет строки заголовка)", name)
	}

	col := make(map[string]int, len(rows[0]))
	for i, h := range rows[0] {
		h = strings.TrimSpace(h)
		if h != "" {
			col[h] = i
		}
	}

	var missing []string
	for _, r := range required {
		if _, ok := col[r]; !ok {
			missing = append(missing, r)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("лист %q: отсутствуют обязательные колонки: %s",
			name, strings.Join(missing, ", "))
	}

	dataRows := rows[1:]

	// Отбрасываем полностью пустые строки (частая ситуация в конце
	// диапазона, который Excel/LibreOffice считает "использованным").
	nonEmpty := dataRows[:0]

	for _, row := range dataRows {
		empty := true

		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				empty = false
				break
			}
		}

		if !empty {
			nonEmpty = append(nonEmpty, row)
		}
	}

	return &sheet{name: name, rows: nonEmpty, col: col}, nil
}

// get возвращает значение колонки name для строки row (или "", если
// колонки нет в файле — актуально для необязательных колонок).
func (s *sheet) get(row []string, name string) string {
	idx, ok := s.col[name]
	if !ok || idx >= len(row) {
		return ""
	}

	return strings.TrimSpace(row[idx])
}

// rowNumber возвращает номер строки в Excel (с учётом заголовка и
// единицы, а не нуля) — для диагностических сообщений.
func (s *sheet) rowNumber(i int) int {
	return i + 2
}

// parseBool разбирает "человеческие" обозначения true/false, которые
// реально появляются в .xlsx (excelize возвращает булевы ячейки как
// строки "TRUE"/"FALSE", но лист может редактироваться руками).
func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes", "да", "истина":
		return true
	default:
		return false
	}
}

const dateLayout = "2006-01-02"

// parseDate разбирает дату в формате YYYY-MM-DD. Пустая строка — это
// отсутствие значения (ptr == nil), а не ошибка.
func parseDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	t, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil, fmt.Errorf("ожидался формат даты YYYY-MM-DD, получено %q: %w", raw, err)
	}

	return &t, nil
}

// parseFloatPtr разбирает необязательное число с плавающей точкой
// (например, координаты). Пустая строка — nil, а не ошибка.
func parseFloatPtr(raw string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("ожидалось число, получено %q: %w", raw, err)
	}

	return &v, nil
}

// strPtr возвращает nil для пустой строки и указатель на значение
// иначе — для необязательных текстовых полей (address, phone, email).
func strPtr(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	return &raw
}
