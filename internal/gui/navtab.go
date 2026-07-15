//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type navTab struct {
	widget.BaseWidget
	icon       *widget.Icon
	label      *canvas.Text
	background *canvas.Rectangle
	underline  *canvas.Rectangle
	onTapped   func()
	active     bool
	hovered    bool
	focused    bool
	iconOnly   bool
	accent     color.Color
}

var _ fyne.Focusable = (*navTab)(nil)

func newPipelineStageItem(text string, icon fyne.Resource, accent color.Color, tapped func()) *navTab {
	item := newSidebarNavItem(text, icon, tapped)
	item.accent = accent
	item.applyStyle()
	return item
}

func newSidebarNavItem(text string, icon fyne.Resource, tapped func()) *navTab {
	tab := &navTab{
		icon:       widget.NewIcon(icon),
		label:      canvas.NewText(text, theme.Color(theme.ColorNameForeground)),
		background: canvas.NewRectangle(color.Transparent),
		underline:  canvas.NewRectangle(color.Transparent),
		onTapped:   tapped,
	}
	tab.label.TextSize = theme.TextSize()
	tab.label.TextStyle = fyne.TextStyle{Bold: true}
	tab.background.CornerRadius = theme.Size(theme.SizeNameSelectionRadius)
	tab.background.SetMinSize(fyne.NewSize(1, 38))
	tab.underline.CornerRadius = 2
	tab.underline.SetMinSize(fyne.NewSize(3, 1))
	tab.ExtendBaseWidget(tab)
	tab.applyStyle()
	return tab
}

func (t *navTab) setActive(active bool) {
	t.active = active
	t.Refresh()
}

func (t *navTab) setIconOnly(iconOnly bool) {
	if t.iconOnly == iconOnly {
		return
	}
	t.iconOnly = iconOnly
	if iconOnly {
		t.label.Hide()
	} else {
		t.label.Show()
	}
	t.Refresh()
}

func (t *navTab) applyStyle() {
	if t.focused {
		t.background.StrokeColor = theme.Color(theme.ColorNameFocus)
		t.background.StrokeWidth = 2
	} else {
		t.background.StrokeColor = color.Transparent
		t.background.StrokeWidth = 0
	}
	if t.active {
		t.label.Color = theme.Color(theme.ColorNamePrimary)
		t.label.TextStyle = fyne.TextStyle{Bold: true}
		if t.accent != nil {
			t.underline.FillColor = t.accent
		} else {
			t.underline.FillColor = theme.Color(theme.ColorNamePrimary)
		}
		t.background.FillColor = theme.Color(theme.ColorNameSelection)
	} else {
		t.label.Color = theme.Color(theme.ColorNameForeground)
		t.label.TextStyle = fyne.TextStyle{Bold: true}
		if t.accent != nil {
			t.underline.FillColor = t.accent
		} else {
			t.underline.FillColor = color.Transparent
		}
		if t.hovered || t.focused {
			t.background.FillColor = theme.Color(theme.ColorNameHover)
		} else {
			t.background.FillColor = color.Transparent
		}
	}
}

func (t *navTab) Refresh() {
	t.applyStyle()
	t.BaseWidget.Refresh()
}

// AccessibilityLabel exposes the destination represented by this custom tab.
func (t *navTab) AccessibilityLabel() string {
	return t.label.Text
}

// AccessibilityRole identifies a navigation tab as an actionable control.
func (t *navTab) AccessibilityRole() fyne.AccessibleRole {
	return fyne.AccessibleRoleButton
}

func (t *navTab) Tapped(*fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

// FocusGained makes the custom navigation item visible in keyboard focus
// order. The focus outline is intentionally independent from active state: a
// user may tab toward another destination without activating it yet.
func (t *navTab) FocusGained() {
	t.focused = true
	t.Refresh()
}

// FocusLost clears the keyboard focus treatment.
func (t *navTab) FocusLost() {
	t.focused = false
	t.Refresh()
}

// TypedRune is required by fyne.Focusable. Navigation items do not consume
// text input.
func (t *navTab) TypedRune(rune) {}

// TypedKey activates a focused destination using the conventional keyboard
// controls for buttons.
func (t *navTab) TypedKey(event *fyne.KeyEvent) {
	if event == nil {
		return
	}
	switch event.Name {
	case fyne.KeySpace, fyne.KeyReturn, fyne.KeyEnter:
		t.Tapped(nil)
	}
}

func (t *navTab) MouseIn(*desktop.MouseEvent) {
	t.hovered = true
	t.Refresh()
}

func (t *navTab) MouseMoved(*desktop.MouseEvent) {}

func (t *navTab) MouseOut() {
	t.hovered = false
	t.Refresh()
}

func (t *navTab) CreateRenderer() fyne.WidgetRenderer {
	row := container.NewHBox(t.icon, t.label)
	content := container.NewBorder(nil, nil, t.underline, nil, container.NewPadded(row))
	return widget.NewSimpleRenderer(container.NewStack(t.background, content))
}
