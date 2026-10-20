package main

import (
	"image/png"
	"os"
	"path/filepath"
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
