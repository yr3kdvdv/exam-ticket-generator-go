package lab2

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Student struct {
	LastName  string
	FirstName string
}

func (s Student) Label() string { return s.LastName + " " + s.FirstName }

type Groups struct {
	Names    []string
	Students map[string][]Student
}

func LoadStudents(path string) (Groups, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return Groups{}, fmt.Errorf("не удалось открыть students.xlsx: %w", err)
	}
	defer f.Close()
	groups := Groups{Names: f.GetSheetList(), Students: make(map[string][]Student)}
	for _, name := range groups.Names {
		rows, err := f.GetRows(name)
		if err != nil {
			return Groups{}, fmt.Errorf("ошибка чтения группы %q: %w", name, err)
		}
		for _, row := range rows[minimum(1, len(rows)):] {
			if len(row) < 2 {
				continue
			}
			student := Student{LastName: strings.TrimSpace(row[0]), FirstName: strings.TrimSpace(row[1])}
			if student.LastName == "" || student.FirstName == "" {
				continue
			}
			groups.Students[name] = append(groups.Students[name], student)
		}
	}
	if len(groups.Names) == 0 {
		return Groups{}, fmt.Errorf("в students.xlsx нет листов с группами")
	}
	return groups, nil
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
