package lab2

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const ResultsSheet = "Результаты"

var resultsHeader = []interface{}{"Группа", "Фамилия", "Имя", "Номер билета", "Дата и время", "Повтор"}

type Result struct {
	Group        string
	Student      Student
	TicketNumber int
	Repeated     bool
	At           time.Time
}

// PrepareResult fixes the ticket number before saving; retrying a failed save
// must never draw a different number for the same click.
func PrepareResult(path, group string, student Student, tickets map[int]Ticket, now time.Time) (Result, error) {
	if group == "" || strings.TrimSpace(student.LastName) == "" || strings.TrimSpace(student.FirstName) == "" {
		return Result{}, fmt.Errorf("выберите группу и студента")
	}
	ids := SortedTicketNumbers(tickets)
	if len(ids) == 0 {
		return Result{}, fmt.Errorf("нет доступных билетов")
	}
	result := Result{Group: group, Student: student, At: now}
	if _, err := os.Stat(path); err == nil {
		f, err := excelize.OpenFile(path)
		if err != nil {
			return Result{}, fmt.Errorf("не удалось прочитать results.xlsx: %w", err)
		}
		defer f.Close()
		rows, err := f.GetRows(ResultsSheet)
		if err != nil {
			return Result{}, fmt.Errorf("не удалось прочитать лист результатов: %w", err)
		}
		for _, row := range rows[minimum(1, len(rows)):] {
			if len(row) < 4 {
				continue
			}
			if row[0] == group && strings.TrimSpace(row[1]) == student.LastName && strings.TrimSpace(row[2]) == student.FirstName {
				n, err := strconv.Atoi(row[3])
				if err != nil {
					return Result{}, fmt.Errorf("в первой записи студента неверный номер билета: %w", err)
				}
				if _, ok := tickets[n]; !ok {
					return Result{}, fmt.Errorf("билет № %d из истории отсутствует в tickets.docx", n)
				}
				result.TicketNumber, result.Repeated = n, true
				return result, nil
			}
		}
	} else if !os.IsNotExist(err) {
		return Result{}, err
	}
	choice, err := rand.Int(rand.Reader, big.NewInt(int64(len(ids))))
	if err != nil {
		return Result{}, err
	}
	result.TicketNumber = ids[choice.Int64()]
	return result, nil
}

func SaveResult(path string, result Result) error {
	lockPath := filepath.Join(filepath.Dir(path), "~$"+filepath.Base(path))
	if _, err := os.Stat(lockPath); err == nil {
		return fmt.Errorf("закройте results.xlsx в Excel и повторите сохранение")
	}
	var f *excelize.File
	if _, err := os.Stat(path); err == nil {
		f, err = excelize.OpenFile(path)
		if err != nil {
			return fmt.Errorf("не удалось открыть results.xlsx: %w", err)
		}
	} else if os.IsNotExist(err) {
		f = excelize.NewFile()
		if err = f.SetSheetName("Sheet1", ResultsSheet); err != nil {
			return err
		}
		if err = f.SetSheetRow(ResultsSheet, "A1", &resultsHeader); err != nil {
			return err
		}
		_ = f.SetColWidth(ResultsSheet, "A", "C", 20)
		_ = f.SetColWidth(ResultsSheet, "D", "D", 16)
		_ = f.SetColWidth(ResultsSheet, "E", "E", 22)
	} else {
		return err
	}
	defer f.Close()
	rows, err := f.GetRows(ResultsSheet)
	if err != nil {
		return fmt.Errorf("не удалось прочитать results.xlsx: %w", err)
	}
	row := []interface{}{result.Group, result.Student.LastName, result.Student.FirstName, result.TicketNumber, result.At.Format("2006-01-02 15:04:05"), "нет"}
	if result.Repeated {
		row[5] = "да"
	}
	cell := fmt.Sprintf("A%d", len(rows)+1)
	if err := f.SetSheetRow(ResultsSheet, cell, &row); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".results-*.xlsx")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)
	if err := f.SaveAs(tmpPath); err != nil {
		return fmt.Errorf("не удалось сохранить results.xlsx: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("не удалось заменить results.xlsx (возможно, файл открыт в Excel): %w", err)
	}
	return nil
}
