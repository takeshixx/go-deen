//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type adversecTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func newAdversecTheme(variant fyne.ThemeVariant) fyne.Theme {
	return &adversecTheme{base: theme.DefaultTheme(), variant: variant}
}

func (t *adversecTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if t.variant == theme.VariantLight {
		switch n {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0xf7, G: 0xf9, B: 0xf8, A: 0xff}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0x0b, G: 0x12, B: 0x12, A: 0xff}
		case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 0x4b, G: 0x5d, B: 0x5b, A: 0xff}
		case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameHeaderBackground:
			return color.NRGBA{R: 0xef, G: 0xf4, B: 0xf3, A: 0xff}
		case theme.ColorNameInputBorder, theme.ColorNameSeparator:
			return color.NRGBA{R: 0x9b, G: 0xad, B: 0xaa, A: 0xff}
		case theme.ColorNamePrimary, theme.ColorNameHyperlink:
			return color.NRGBA{R: 0x0b, G: 0x7f, B: 0x75, A: 0xff}
		case theme.ColorNameFocus, theme.ColorNameWarning:
			return color.NRGBA{R: 0x9a, G: 0x67, B: 0x00, A: 0xff}
		case theme.ColorNameSuccess:
			return color.NRGBA{R: 0x13, G: 0x7f, B: 0x4f, A: 0xff}
		case theme.ColorNameError:
			return color.NRGBA{R: 0xbf, G: 0x2f, B: 0x35, A: 0xff}
		}
		return t.base.Color(n, t.variant)
	}

	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x15, G: 0x20, B: 0x2b, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xf2, G: 0xf5, B: 0xf4, A: 0xff}
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb8, G: 0xc3, B: 0xc1, A: 0xff}
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x10, G: 0x13, B: 0x14, A: 0xff}
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0x15, G: 0x20, B: 0x2b, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x17, G: 0x1c, B: 0x1d, A: 0xff}
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return color.NRGBA{R: 0x34, G: 0x41, B: 0x42, A: 0xff}
	case theme.ColorNameHover, theme.ColorNamePressed, theme.ColorNameSelection:
		return color.NRGBA{R: 0x12, G: 0x4c, B: 0x48, A: 0xff}
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return color.NRGBA{R: 0x38, G: 0xd9, B: 0xc8, A: 0xff}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0x03, G: 0x10, B: 0x0f, A: 0xff}
	case theme.ColorNameFocus, theme.ColorNameWarning:
		return color.NRGBA{R: 0xf2, G: 0xc6, B: 0x6d, A: 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x61, G: 0xd3, B: 0x94, A: 0xff}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xff, G: 0x6b, B: 0x6b, A: 0xff}
	}
	return t.base.Color(n, t.variant)
}

func (t *adversecTheme) Font(s fyne.TextStyle) fyne.Resource     { return t.base.Font(s) }
func (t *adversecTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return t.base.Icon(n) }
func (t *adversecTheme) Size(n fyne.ThemeSizeName) float32       { return t.base.Size(n) }
