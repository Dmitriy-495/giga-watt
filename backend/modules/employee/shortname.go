package employee

import "strings"

// BuildShortName формирует сокращённое ФИО из фамилии и инициалов имени
// и отчества, например: "Иванов И. И." (см. ITERATION.md: "Сокращённое
// ФИО формируется автоматически из фамилии и инициалов").
func BuildShortName(lastName, firstName, middleName string) string {
	short := strings.TrimSpace(lastName)

	if initial := firstInitial(firstName); initial != "" {
		short += " " + initial + "."
	}

	if initial := firstInitial(middleName); initial != "" {
		short += " " + initial + "."
	}

	return short
}

// firstInitial возвращает первую букву имени (по рунам, корректно
// работает с кириллицей).
func firstInitial(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	runes := []rune(name)

	return strings.ToUpper(string(runes[0:1]))
}
