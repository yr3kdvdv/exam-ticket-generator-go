package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/yr3kdvdv/exam-ticket-generator-go/internal/lab2"
)

func gap(w, h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(w, h))
	return r
}

func inset(child fyne.CanvasObject, top, bottom, left, right float32) fyne.CanvasObject {
	return container.NewBorder(gap(0, top), gap(0, bottom), gap(left, 0), gap(right, 0), child)
}

func rounded(fill color.Color, radius float32) *canvas.Rectangle {
	r := canvas.NewRectangle(fill)
	r.CornerRadius = radius
	return r
}

func headline(value string, size float32, fill color.Color, bold bool) *canvas.Text {
	t := canvas.NewText(value, fill)
	t.TextSize = size
	t.TextStyle.Bold = bold
	return t
}

func pill(label string, width float32, background, foreground color.Color) fyne.CanvasObject {
	bg := rounded(background, 12)
	bg.SetMinSize(fyne.NewSize(width, 28))
	t := headline(label, 12, foreground, true)
	return container.NewStack(bg, container.NewCenter(t))
}

func frame(content fyne.CanvasObject) fyne.CanvasObject {
	card := rounded(palette.white, 18)
	card.Shadow = canvas.Shadow{
		Color:      color.NRGBA{R: 0x20, G: 0x38, B: 0x5c, A: 0x25},
		BlurRadius: 16,
		Offset:     fyne.NewPos(0, 7),
		Variant:    canvas.DropShadow,
	}
	return container.NewStack(
		canvas.NewRectangle(palette.paper),
		inset(container.NewStack(card, content), 20, 20, 22, 22),
	)
}

func appHeader() fyne.CanvasObject {
	background := rounded(palette.navy, 16)
	iconBackground := rounded(palette.softBlue, 10)
	iconBackground.SetMinSize(fyne.NewSize(43, 43))
	icon := widget.NewIcon(theme.DocumentIcon())
	iconBox := container.NewStack(iconBackground, inset(icon, 9, 9, 9, 9))
	titles := container.NewVBox(
		headline("Генератор билетов", 23, palette.white, true),
		headline("Proiectarea Sistemelor Informaționale", 13, color.NRGBA{R: 0xc9, G: 0xd8, B: 0xea, A: 0xff}, false),
	)
	row := container.NewHBox(iconBox, gap(11, 0), titles, layout.NewSpacer())
	return container.NewStack(background, inset(row, 20, 20, 22, 22))
}

func step(number, title, subtitle string, active bool) fyne.CanvasObject {
	bg := color.Color(palette.softBlue)
	fg := color.Color(palette.muted)
	if active {
		bg = color.NRGBA{R: 0xdd, G: 0xf1, B: 0xea, A: 0xff}
		fg = palette.green
	}
	numberBox := pill(number, 35, bg, fg)
	labels := container.NewVBox(
		headline(title, 17, palette.ink, true),
		headline(subtitle, 13, palette.muted, false),
	)
	return container.NewHBox(numberBox, gap(9, 0), labels)
}

func field(label string, selectWidget *widget.Select) fyne.CanvasObject {
	return container.NewVBox(
		headline(label, 13, palette.muted, true),
		gap(0, 5),
		selectWidget,
	)
}

func selectionView(group, student *widget.Select, generate *widget.Button, status *widget.Label) fyne.CanvasObject {
	statusBackground := rounded(color.NRGBA{R: 0xf1, G: 0xf6, B: 0xfb, A: 0xff}, 10)
	statusDot := canvas.NewCircle(color.NRGBA{R: 0x20, G: 0xa7, B: 0x7a, A: 0xff})
	statusIcon := container.NewCenter(container.NewGridWrap(fyne.NewSize(8, 8), statusDot))
	statusRow := container.NewHBox(statusIcon, gap(7, 0), status)
	statusStrip := container.NewStack(statusBackground, inset(statusRow, 6, 6, 10, 10))

	buttonHeight := gap(0, 55)
	button := container.NewStack(buttonHeight, generate)
	form := container.NewVBox(
		container.NewHBox(pill("ЛАБОРАТОРНАЯ 2", 178, palette.softBlue, color.NRGBA{R: 0x23, G: 0x59, B: 0xac, A: 0xff}), layout.NewSpacer()),
		gap(0, 12),
		headline("Выдача билета", 29, palette.ink, true),
		headline("Выберите студента, чтобы открыть билет.", 16, palette.muted, false),
		gap(0, 13),
		widget.NewSeparator(),
		gap(0, 11),
		container.NewGridWithColumns(2,
			step("1", "Студент", "Группа и имя", true),
			step("2", "Билет", "Три вопроса", false),
		),
		gap(0, 17),
		container.NewGridWithColumns(2,
			field("ГРУППА", group),
			field("СТУДЕНТ", student),
		),
		gap(0, 20),
		button,
		gap(0, 16),
		statusStrip,
	)
	return frame(container.NewVBox(appHeader(), inset(form, 22, 22, 28, 28)))
}

func questionCard(number int, question string) fyne.CanvasObject {
	bg := rounded(palette.field, 12)
	bg.StrokeColor = palette.line
	bg.StrokeWidth = 1
	bg.SetMinSize(fyne.NewSize(0, 72))
	numberChip := pill(fmt.Sprint(number), 34, color.NRGBA{R: 0xdf, G: 0xec, B: 0xfc, A: 0xff}, color.NRGBA{R: 0x23, G: 0x59, B: 0xac, A: 0xff})
	q := widget.NewLabel(question)
	q.Wrapping = fyne.TextWrapWord
	content := container.NewBorder(nil, nil, numberChip, nil, q)
	return container.NewStack(bg, inset(content, 12, 12, 13, 13))
}

func resultView(result lab2.Result, ticket lab2.Ticket, back func()) fyne.CanvasObject {
	badge := "БИЛЕТ СОХРАНЁН"
	if result.Repeated {
		badge = "ПОВТОРНЫЙ БИЛЕТ"
	}
	backButton := widget.NewButtonWithIcon("К выбору студента", theme.NavigateBackIcon(), back)
	backButton.Importance = widget.MediumImportance
	footer := container.NewBorder(nil, nil,
		headline("ESC — вернуться к выбору", 13, palette.muted, false),
		backButton,
	)
	questions := container.NewVBox(
		container.NewHBox(pill(badge, 185, color.NRGBA{R: 0xe1, G: 0xf4, B: 0xed, A: 0xff}, palette.green), layout.NewSpacer()),
		gap(0, 8),
		headline(fmt.Sprintf("Ваш билет № %02d", result.TicketNumber), 30, palette.ink, true),
		headline(result.Student.Label()+"   ·   "+result.Group, 16, palette.muted, false),
		gap(0, 12),
		widget.NewSeparator(),
		gap(0, 8),
		questionCard(1, ticket.Questions[0]),
		gap(0, 7),
		questionCard(2, ticket.Questions[1]),
		gap(0, 7),
		questionCard(3, ticket.Questions[2]),
	)
	cardContent := container.NewBorder(
		appHeader(),
		inset(footer, 9, 18, 26, 26),
		nil, nil,
		container.NewScroll(inset(questions, 20, 12, 26, 26)),
	)
	return frame(cardContent)
}
