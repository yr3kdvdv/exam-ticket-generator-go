package main

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/yr3kdvdv/exam-ticket-generator-go/internal/lab2"
)

func TestViewsRender(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(&academicTheme{})
	group := widget.NewSelect([]string{"SI-231", "SI-232"}, nil)
	group.PlaceHolder = "Выберите группу"
	group.SetSelected("SI-231")
	student := widget.NewSelect([]string{"Popescu Ana"}, nil)
	student.PlaceHolder = "Выберите студента"
	student.SetSelected("Popescu Ana")
	generate := widget.NewButton("Сгенерировать билет", nil)
	generate.Importance = widget.HighImportance

	views := []struct {
		name string
		size fyne.Size
		view fyne.CanvasObject
	}{
		{
			name: "selection",
			size: fyne.NewSize(760, 675),
			view: selectionView(
				group,
				student,
				generate,
				widget.NewLabel("20 билетов загружено  ·  запись в results.xlsx"),
			),
		},
		{
			name: "result",
			size: fyne.NewSize(650, 690),
			view: resultView(
				lab2.Result{Group: "SI-231", Student: lab2.Student{LastName: "Popescu", FirstName: "Ana"}, TicketNumber: 7},
				lab2.Ticket{Questions: [3]string{
					"Explicați etapele analizei cerințelor pentru un sistem informațional.",
					"Descrieți rolul diagramelor UML în proiectarea sistemului.",
					"Cum se verifică integritatea datelor în baza de date proiectată?",
				}},
				func() {},
			),
		},
	}

	for _, tc := range views {
		t.Run(tc.name, func(t *testing.T) {
			window := test.NewWindow(tc.view)
			defer window.Close()
			window.Resize(tc.size)
			window.Show()
			captured := window.Canvas().Capture()
			if captured == nil || captured.Bounds().Dx() < int(tc.size.Width) || captured.Bounds().Dy() < int(tc.size.Height) {
				t.Fatalf("incomplete view capture: %v", captured)
			}
			if dir := os.Getenv("LAB2_PREVIEW_DIR"); dir != "" {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(filepath.Join(dir, tc.name+".png"))
				if err != nil {
					t.Fatal(err)
				}
				if err := png.Encode(file, captured); err != nil {
					file.Close()
					t.Fatal(err)
				}
				if err := file.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestIssueTicketAndReturn(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	dir := t.TempDir()
	for _, name := range []string{"students.xlsx", "tickets.docx"} {
		data, err := os.ReadFile(filepath.Join("..", "..", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	window, group, student, generate := buildSelectionWindow(a, dir)
	defer window.Close()
	window.Show()
	if !student.Disabled() || !generate.Disabled() {
		t.Fatal("student and issue controls must start disabled")
	}
	group.SetSelected(group.Options[0])
	student.SetSelected(student.Options[0])
	group.SetSelected(group.Options[1])
	if student.Selected != "" || !generate.Disabled() {
		t.Fatal("changing the group must clear the student and disable issuance")
	}
	student.SetSelected(student.Options[0])
	if generate.Disabled() {
		t.Fatal("valid selection did not enable issuance")
	}
	test.Tap(generate)
	if _, err := os.Stat(filepath.Join(dir, "results.xlsx")); err != nil {
		t.Fatal(err)
	}
	for _, ticket := range a.Driver().AllWindows() {
		if strings.HasPrefix(ticket.Title(), "Ваш билет №") {
			ticket.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
			for _, open := range a.Driver().AllWindows() {
				if open == ticket {
					t.Fatal("Escape did not close the ticket")
				}
			}
			return
		}
	}
	t.Fatal("issuing the ticket did not open its window")
}
