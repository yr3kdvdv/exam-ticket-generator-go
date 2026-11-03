package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var palette = struct {
	navy, ink, muted, blue, paper, white, softBlue, field, line, green color.NRGBA
}{
	navy:     rgb(0x202C46),
	ink:      rgb(0xF0F4FF),
	muted:    rgb(0xB0BED5),
	blue:     rgb(0x315BCB),
	paper:    rgb(0x0D1424),
	white:    rgb(0x151F33),
	softBlue: rgb(0x263958),
	field:    rgb(0x1D2A41),
	line:     rgb(0x394C6A),
	green:    rgb(0x7CE6BC),
}

func rgb(hex uint32) color.NRGBA {
	return color.NRGBA{R: uint8(hex >> 16), G: uint8(hex >> 8), B: uint8(hex), A: 255}
}

// academicTheme applies the dark palette regardless of system settings.
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
		return palette.ink
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
		return theme.DefaultTheme().Color(name, theme.VariantDark)
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
