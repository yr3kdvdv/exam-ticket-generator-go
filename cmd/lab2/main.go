package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/yr3kdvdv/exam-ticket-generator-go/internal/lab2"
)

func fileDirectory() string {
	if _, err := os.Stat("students.xlsx"); err == nil {
		return "."
	}
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func main() {
	a := app.New()
	mainWindow := a.NewWindow("Генератор билетов — Лаб 2")
	mainWindow.Resize(fyne.NewSize(620, 380))
	base := fileDirectory()
	groups, studentsErr := lab2.LoadStudents(filepath.Join(base, "students.xlsx"))
	tickets, warnings, ticketsErr := lab2.LoadTickets(filepath.Join(base, "tickets.docx"))
	for _, warning := range warnings {
		log.Printf("tickets.docx: %s", warning)
	}

	groupSelect := widget.NewSelect(groups.Names, nil)
	groupSelect.PlaceHolder = "Выберите группу"
	studentSelect := widget.NewSelect(nil, nil)
	studentSelect.PlaceHolder = "Выберите студента"
	studentSelect.Disable()
	generate := widget.NewButton("Сгенерировать билет", nil)
	generate.Importance = widget.HighImportance
	generate.Disable()
	updateButton := func() {
		if groupSelect.Selected != "" && studentSelect.Selected != "" && studentsErr == nil && ticketsErr == nil {
			generate.Enable()
		} else {
			generate.Disable()
		}
	}
	groupSelect.OnChanged = func(group string) {
		studentSelect.ClearSelected()
		studentSelect.Options = nil
		for _, student := range groups.Students[group] {
			studentSelect.Options = append(studentSelect.Options, student.Label())
		}
		studentSelect.Refresh()
		if len(studentSelect.Options) == 0 {
			studentSelect.Disable()
		} else {
			studentSelect.Enable()
		}
		updateButton()
	}
	studentSelect.OnChanged = func(string) { updateButton() }

	status := widget.NewLabel("Выберите группу и студента. Результат сохраняется в results.xlsx.")
	status.Wrapping = fyne.TextWrapWord
	if studentsErr != nil {
		status.SetText(fmt.Sprintf("Ошибка: %v", studentsErr))
	}
	if ticketsErr != nil {
		status.SetText(fmt.Sprintf("Ошибка: %v", ticketsErr))
	}
	if len(warnings) > 0 && ticketsErr == nil {
		status.SetText(fmt.Sprintf("Загружено билетов: %d. Пропущено блоков: %d. Подробности в журнале приложения.", len(tickets), len(warnings)))
	}

	showResult := func(result lab2.Result) {
		ticket := tickets[result.TicketNumber]
		resultWindow := a.NewWindow(fmt.Sprintf("Ваш билет № %d", result.TicketNumber))
		resultWindow.Resize(fyne.NewSize(690, 420))
		content := container.NewVBox(
			widget.NewLabelWithStyle(fmt.Sprintf("Ваш билет № %d", result.TicketNumber), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("1. "+ticket.Questions[0]),
			widget.NewLabel("2. "+ticket.Questions[1]),
			widget.NewLabel("3. "+ticket.Questions[2]),
			widget.NewLabel("Нажмите ESC, чтобы вернуться к выбору студента."),
		)
		for _, object := range content.Objects {
			if label, ok := object.(*widget.Label); ok {
				label.Wrapping = fyne.TextWrapWord
			}
		}
		resultWindow.SetContent(container.NewPadded(content))
		resultWindow.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
			if key.Name == fyne.KeyEscape {
				resultWindow.Close()
			}
		})
		resultWindow.SetOnClosed(func() { mainWindow.Show() })
		resultWindow.Show()
		mainWindow.Hide()
	}

	generate.OnTapped = func() {
		var chosen lab2.Student
		for _, student := range groups.Students[groupSelect.Selected] {
			if student.Label() == studentSelect.Selected {
				chosen = student
				break
			}
		}
		path := filepath.Join(base, "results.xlsx")
		result, err := lab2.PrepareResult(path, groupSelect.Selected, chosen, tickets, time.Now())
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		var save func()
		save = func() {
			if err := lab2.SaveResult(path, result); err != nil {
				dialog.NewConfirm("Не удалось сохранить результат", err.Error()+"\nЗакройте файл в Excel и повторите попытку.", func(retry bool) {
					if retry {
						save()
					}
				}, mainWindow).Show()
				return
			}
			showResult(result)
		}
		save()
	}

	mainWindow.SetContent(container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Генератор экзаменационных билетов", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Proiectarea Sistemelor Informaționale"),
		widget.NewSeparator(),
		widget.NewLabel("Группа"), groupSelect,
		widget.NewLabel("Студент"), studentSelect,
		generate, status,
	)))
	mainWindow.ShowAndRun()
}
