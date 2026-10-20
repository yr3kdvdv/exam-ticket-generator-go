package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var palette = struct {
	navy, ink, muted, blue, paper, white, softBlue, field, line, green color.NRGBA
}{
	navy:     color.NRGBA{R: 0x14, G: 0x2d, B: 0x4b, A: 0xff},
	ink:      color.NRGBA{R: 0x17, G: 0x2b, B: 0x43, A: 0xff},
	muted:    color.NRGBA{R: 0x69, G: 0x7b, B: 0x91, A: 0xff},
	blue:     color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff},
	paper:    color.NRGBA{R: 0xea, G: 0xf0, B: 0xf7, A: 0xff},
	white:    color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	softBlue: color.NRGBA{R: 0xe7, G: 0xf1, B: 0xff, A: 0xff},
	field:    color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff},
	line:     color.NRGBA{R: 0xdc, G: 0xe4, B: 0xed, A: 0xff},
	green:    color.NRGBA{R: 0x16, G: 0x82, B: 0x65, A: 0xff},
}

// academicTheme keeps Fyne controls native while giving both windows the
// chosen light palette, including on machines configured for dark mode.
type academicTheme struct{}

var _ fyne.Theme = (*academicTheme)(nil)

func (*academicTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return palette.paper
	case theme.ColorNameForeground:
		return palette.ink
	case theme.ColorNamePrimary:
		return palette.blue
	case theme.ColorNameForegroundOnPrimary:
		return palette.white
	case theme.ColorNameInputBackground:
		return palette.field
	case theme.ColorNameButton:
		return palette.softBlue
	case theme.ColorNameSeparator:
		return palette.line
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return palette.muted
	case theme.ColorNameSuccess:
		return palette.green
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (*academicTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (*academicTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (*academicTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 15
	}
	return theme.DefaultTheme().Size(name)
}
