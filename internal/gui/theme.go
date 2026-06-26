//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
)

// forcedVariantTheme wraps a base theme but pins the light/dark variant,
// letting the user override the automatic (OS-following) theme. The deprecated
// theme.LightTheme/DarkTheme helpers are avoided in favour of this.
type forcedVariantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (t *forcedVariantTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.base.Color(n, t.variant)
}
func (t *forcedVariantTheme) Font(s fyne.TextStyle) fyne.Resource     { return t.base.Font(s) }
func (t *forcedVariantTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return t.base.Icon(n) }
func (t *forcedVariantTheme) Size(n fyne.ThemeSizeName) float32       { return t.base.Size(n) }
