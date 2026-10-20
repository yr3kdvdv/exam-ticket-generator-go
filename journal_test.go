package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJournalCreateAndAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.xlsx")
	if err := appendJournal(path, "Иванов", "Иван", 7); err != nil {
		t.Fatal(err)
	}
	first := readSheet(t, path)
	if !strings.Contains(first, "Иванов") || !strings.Contains(first, `r="C2"><v>7</v>`) {
		t.Fatalf("first record missing: %s", first)
	}
	if err := appendJournal(path, "Петров", "Пётр", 13); err != nil {
		t.Fatal(err)
	}
	second := readSheet(t, path)
	if !strings.Contains(second, "Иванов") || !strings.Contains(second, "Петров") {
		t.Fatalf("records missing: %s", second)
	}
	if !strings.Contains(second, `r="C3"><v>13</v>`) {
		t.Fatalf("third row missing: %s", second)
	}
	if !strings.HasPrefix(second, strings.TrimSuffix(first, "</sheetData></worksheet>")) {
		t.Fatal("existing rows changed during append")
	}
	if got, err := nextRow([]byte(second)); err != nil || got != 4 {
		t.Fatalf("nextRow = %d, %v; want 4", got, err)
	}
}

func TestJournalEscapesNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.xlsx")
	if err := appendJournal(path, "А&Б", "<Имя>", 1); err != nil {
		t.Fatal(err)
	}
	sheet := readSheet(t, path)
	if !strings.Contains(sheet, "А&amp;Б") || !strings.Contains(sheet, "&lt;Имя&gt;") {
		t.Fatalf("XML escaping failed: %s", sheet)
	}
}

func TestJournalExcelLockKeepsPreviousData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.xlsx")
	if err := appendJournal(path, "First", "Student", 1); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "~$journal.xlsx"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := appendJournal(path, "Second", "Student", 2); err == nil {
		t.Fatal("append succeeded while Excel lock was present")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("journal changed after rejected append")
	}
}

func TestEmptyNameRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.xlsx")
	if err := appendJournal(path, "  ", "Student", 1); err == nil {
		t.Fatal("empty last name accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("journal exists after rejected student: %v", err)
	}
}

func TestReadNonemptyAndESC(t *testing.T) {
	input, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if _, err := io.WriteString(writer, "   \n  Иванов  \n\x1b"); err != nil {
		t.Fatal(err)
	}
	writer.Close()
	var output strings.Builder
	c, err := newConsole(input, &output)
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	name, err := c.readNonempty("Last name: ")
	if err != nil || name != "Иванов" {
		t.Fatalf("name = %q, %v", name, err)
	}
	if _, err := c.readNonempty("Last name: "); err != errExit {
		t.Fatalf("ESC returned %v", err)
	}
	if !strings.Contains(output.String(), "Поле не может быть пустым.") {
		t.Fatal("empty-field hint missing")
	}
}

func TestQuestionBank(t *testing.T) {
	for i, question := range questions {
		if strings.TrimSpace(question) == "" {
			t.Fatalf("ticket %d has no question", i+1)
		}
	}
}

func readSheet(t *testing.T, path string) string {
	t.Helper()
	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	for _, f := range z.File {
		if f.Name == sheetPath {
			r, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(r)
			r.Close()
			if err != nil {
				t.Fatal(err)
			}
			return string(body)
		}
	}
	t.Fatal("sheet1.xml missing")
	return ""
}
