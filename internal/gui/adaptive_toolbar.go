//go:build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// adaptiveCommandLayout keeps the primary command set visible and removes
// secondary commands from the row only when the complete row cannot fit. The
// enclosing horizontal scroll remains as a final fallback below the supported
// window minimum or with unusually large theme metrics.
type adaptiveCommandLayout struct {
	objects       []fyne.CanvasObject
	compactHidden map[fyne.CanvasObject]bool
	onCompact     func(bool)
	compact       bool
	initialized   bool
}

func (layout *adaptiveCommandLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	compact := size.Width < layout.width(layout.objects)
	if !layout.initialized || compact != layout.compact {
		layout.compact = compact
		layout.initialized = true
		if layout.onCompact != nil {
			layout.onCompact(compact)
		}
	}

	visible := layout.visibleObjects(compact, true)
	x := float32(0)
	padding := theme.Padding()
	for i, object := range visible {
		object.Move(fyne.NewPos(x, 0))
		object.Resize(fyne.NewSize(object.MinSize().Width, size.Height))
		x += object.MinSize().Width
		if i < len(visible)-1 {
			x += padding
		}
	}
}

func (layout *adaptiveCommandLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	compact := layout.visibleObjects(true, false)
	return fyne.NewSize(layout.width(compact), layout.height(compact))
}

func (layout *adaptiveCommandLayout) visibleObjects(compact, applyVisibility bool) []fyne.CanvasObject {
	visible := make([]fyne.CanvasObject, 0, len(layout.objects))
	for _, object := range layout.objects {
		show := !compact || !layout.compactHidden[object]
		if applyVisibility {
			setAdaptiveObjectVisible(object, show)
		}
		if show {
			visible = append(visible, object)
		}
	}
	return visible
}

func (layout *adaptiveCommandLayout) width(objects []fyne.CanvasObject) float32 {
	width := float32(0)
	for i, object := range objects {
		width += object.MinSize().Width
		if i < len(objects)-1 {
			width += theme.Padding()
		}
	}
	return width
}

func (layout *adaptiveCommandLayout) height(objects []fyne.CanvasObject) float32 {
	height := float32(0)
	for _, object := range objects {
		height = max(height, object.MinSize().Height)
	}
	return height
}
