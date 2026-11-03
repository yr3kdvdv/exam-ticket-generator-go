package lab2

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestParseTicketsSkipsBrokenAndDuplicateBlocks(t *testing.T) {
	paragraphs := []string{
		"Билет 1", "1. Первый вопрос", "2. Второй вопрос", "3. Третий вопрос",
		"Билет 2", "1. Только один вопрос",
		"Билет 1", "1. Дубликат", "2. Второй", "3. Третий",
		"Билет 3", "1. Один", "2. Два", "3. Три", "4. Лишний",
	}
	tickets, warnings, err := ParseTickets(paragraphs)
	if err != nil {
		t.Fatal(err)
	}
	if len(tickets) != 2 || len(warnings) != 2 {
		t.Fatalf("tickets=%v warnings=%v", tickets, warnings)
	}
	if tickets[1].Questions[0] != "Первый вопрос" || tickets[3].Questions[2] != "Три" {
		t.Fatalf("неверные вопросы: %+v", tickets)
	}
}

func TestLoadTicketsReadsSplitRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tickets.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	w, err := z.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprint(w, `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`+
		`<w:p><w:r><w:t>Билет </w:t></w:r><w:r><w:t>7</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>1. Первый</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>2. Второй</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>3. Третий</w:t></w:r></w:p></w:body></w:document>`)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	tickets, _, err := LoadTickets(path)
	if err != nil {
		t.Fatal(err)
	}
	if tickets[7].Questions[1] != "Второй" {
		t.Fatalf("unexpected: %v", tickets)
	}
}

func TestStudentsAndRepeatHistory(t *testing.T) {
	dir := t.TempDir()
	studentsPath := filepath.Join(dir, "students.xlsx")
	f := excelize.NewFile()
	if err := f.SetSheetName("Sheet1", "TI-221"); err != nil {
		t.Fatal(err)
	}
	for i, row := range [][]interface{}{{"Фамилия", "Имя"}, {" Ионеску ", " Ана "}, {"", ""}} {
		if err := f.SetSheetRow("TI-221", fmt.Sprintf("A%d", i+1), &row); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.NewSheet("EMPTY"); err != nil {
		t.Fatal(err)
	}
	if err := f.SaveAs(studentsPath); err != nil {
		t.Fatal(err)
	}
	f.Close()
	groups, err := LoadStudents(studentsPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups.Names) != 2 || len(groups.Students["EMPTY"]) != 0 || groups.Students["TI-221"][0].Label() != "Ионеску Ана" {
		t.Fatalf("wrong groups: %+v", groups)
	}

	resultsPath := filepath.Join(dir, "results.xlsx")
	tickets := map[int]Ticket{3: {Number: 3}, 7: {Number: 7}}
	student := groups.Students["TI-221"][0]
	first, err := PrepareResult(resultsPath, "TI-221", student, tickets, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if first.Repeated {
		t.Fatal("first result marked repeat")
	}
	if err := SaveResult(resultsPath, first); err != nil {
		t.Fatal(err)
	}
	second, err := PrepareResult(resultsPath, "TI-221", student, tickets, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !second.Repeated || second.TicketNumber != first.TicketNumber {
		t.Fatalf("repeat changed ticket: %+v %+v", first, second)
	}
	if err := SaveResult(resultsPath, second); err != nil {
		t.Fatal(err)
	}
	f2, err := excelize.OpenFile(resultsPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	rows, err := f2.GetRows(ResultsSheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[1][5] != "нет" || rows[2][5] != "да" {
		t.Fatalf("wrong history: %v", rows)
	}
	if err := os.WriteFile(filepath.Join(dir, "~$results.xlsx"), []byte("locked"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := SaveResult(resultsPath, second); err == nil || !strings.Contains(err.Error(), "закройте") {
		t.Fatalf("expected lock error, got %v", err)
	}
}

func TestBundledDemoInputs(t *testing.T) {
	root := filepath.Join("..", "..")
	groups, err := LoadStudents(filepath.Join(root, "students.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	if len(groups.Names) != 3 || len(groups.Students["TI-221"]) != 4 || len(groups.Students["EMPTY"]) != 0 {
		t.Fatalf("unexpected demo groups: %+v", groups)
	}
	tickets, warnings, err := LoadTickets(filepath.Join(root, "tickets.docx"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tickets) != 20 || len(warnings) != 0 {
		t.Fatalf("demo tickets=%d warnings=%v", len(tickets), warnings)
	}
}
