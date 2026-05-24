package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type synthwaveTheme struct{}

func (synthwaveTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{0x1a, 0x1a, 0x2e, 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{0x0f, 0x34, 0x60, 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{0x7a, 0x7a, 0x9a, 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{0x0f, 0x34, 0x60, 0x60}
	case theme.ColorNameError:
		return color.NRGBA{0xe9, 0x45, 0x60, 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{0xe9, 0x45, 0x60, 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{0xea, 0xea, 0xea, 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{0xe9, 0x45, 0x60, 0x30}
	case theme.ColorNameInputBackground:
		return color.NRGBA{0x16, 0x21, 0x3e, 0xff}
	case theme.ColorNameInputBorder:
		return color.NRGBA{0x0f, 0x34, 0x60, 0xff}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{0x16, 0x21, 0x3e, 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{0x16, 0x21, 0x3e, 0xe0}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{0x7a, 0x7a, 0x9a, 0xff}
	case theme.ColorNamePressed:
		return color.NRGBA{0xe9, 0x45, 0x60, 0x50}
	case theme.ColorNamePrimary:
		return color.NRGBA{0xe9, 0x45, 0x60, 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{0xe9, 0x45, 0x60, 0x80}
	case theme.ColorNameSelection:
		return color.NRGBA{0x0f, 0x34, 0x60, 0xff}
	case theme.ColorNameSeparator:
		return color.NRGBA{0x0f, 0x34, 0x60, 0xff}
	case theme.ColorNameShadow:
		return color.NRGBA{0x00, 0x00, 0x00, 0x40}
	case theme.ColorNameSuccess:
		return color.NRGBA{0x4e, 0xcc, 0xa3, 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{0xff, 0x95, 0x00, 0xff}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (synthwaveTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (synthwaveTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (synthwaveTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	default:
		return theme.DefaultTheme().Size(name)
	}
}
