//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type navTab struct {
	widget.BaseWidget
	label     *canvas.Text
	underline *canvas.Rectangle
	onTapped  func()
	active    bool
}

func newNavTab(text string, tapped func()) *navTab {
	tab := &navTab{
		label:     canvas.NewText(text, theme.Color(theme.ColorNameForeground)),
		underline: canvas.NewRectangle(color.Transparent),
		onTapped:  tapped,
	}
	tab.label.TextSize = theme.TextSize()
	tab.underline.SetMinSize(fyne.NewSize(1, 2))
	tab.ExtendBaseWidget(tab)
	return tab
}

func (t *navTab) setActive(active bool) {
	t.active = active
	if active {
		t.label.Color = theme.Color(theme.ColorNamePrimary)
		t.label.TextStyle = fyne.TextStyle{Bold: true}
		t.underline.FillColor = theme.Color(theme.ColorNamePrimary)
	} else {
		t.label.Color = theme.Color(theme.ColorNameForeground)
		t.label.TextStyle = fyne.TextStyle{}
		t.underline.FillColor = color.Transparent
	}
	t.label.Refresh()
	t.underline.Refresh()
	t.Refresh()
}

func (t *navTab) Tapped(*fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *navTab) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewBorder(nil, t.underline, nil, nil, container.NewPadded(t.label))
	return widget.NewSimpleRenderer(content)
}
