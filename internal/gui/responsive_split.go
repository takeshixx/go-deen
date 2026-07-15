//go:build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

const (
	defaultCompactSplit             = 0.50
	defaultCompactMasterDetailSplit = 0.56
	minCompactSplit                 = 0.25
	maxCompactSplit                 = 0.75
	compactWorkspaceSplitWidth      = 720
	compactWorkspaceSplitMinHeight  = 520
)

type responsiveSplitController struct {
	gui                 *DeenGUI
	split               *container.Split
	host                *fyne.Container
	compactWhen         func(fyne.Size) bool
	horizontalKey       string
	compactKey          string
	normalizeHorizontal func(float64) float64
	horizontalOffset    float64
	compactOffset       float64
	compact             bool
	initialized         bool
}

type responsiveSplitLayout struct {
	controller *responsiveSplitController
}

func newResponsiveSplit(
	dg *DeenGUI,
	split *container.Split,
	horizontalKey string,
	compactKey string,
	horizontalFallback float64,
	compactFallback float64,
	normalizeHorizontal func(float64) float64,
	compactWhen func(fyne.Size) bool,
) *responsiveSplitController {
	horizontalOffset := normalizeHorizontal(horizontalFallback)
	compactOffset := normalizeCompactSplit(compactFallback)
	if dg != nil && dg.app != nil {
		preferences := dg.app.Preferences()
		horizontalOffset = normalizeHorizontal(preferences.FloatWithFallback(horizontalKey, horizontalOffset))
		compactOffset = normalizeCompactSplit(preferences.FloatWithFallback(compactKey, compactOffset))
	}
	split.Horizontal = true
	split.SetOffset(horizontalOffset)
	controller := &responsiveSplitController{
		gui:                 dg,
		split:               split,
		compactWhen:         compactWhen,
		horizontalKey:       horizontalKey,
		compactKey:          compactKey,
		normalizeHorizontal: normalizeHorizontal,
		horizontalOffset:    horizontalOffset,
		compactOffset:       compactOffset,
	}
	controller.host = container.New(&responsiveSplitLayout{controller: controller}, split)
	return controller
}

func (layout *responsiveSplitLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	controller := layout.controller
	compact := controller.compactWhen != nil && controller.compactWhen(size)
	controller.setCompact(compact)
	controller.split.Move(fyne.NewPos(0, 0))
	controller.split.Resize(size)
}

func (layout *responsiveSplitLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return layout.controller.split.MinSize()
}

func (controller *responsiveSplitController) setCompact(compact bool) {
	if controller.initialized && controller.compact == compact {
		return
	}
	if controller.initialized {
		controller.captureCurrentOffset()
	}
	controller.compact = compact
	controller.initialized = true
	controller.split.Horizontal = !compact
	if compact {
		controller.split.SetOffset(controller.compactOffset)
	} else {
		controller.split.SetOffset(controller.horizontalOffset)
	}
	controller.split.Refresh()
}

func (controller *responsiveSplitController) captureCurrentOffset() {
	if controller.compact {
		controller.compactOffset = normalizeCompactSplit(controller.split.Offset)
	} else {
		controller.horizontalOffset = controller.normalizeHorizontal(controller.split.Offset)
	}
}

func (controller *responsiveSplitController) remember() (horizontal float64) {
	if controller == nil {
		return 0
	}
	controller.captureCurrentOffset()
	if controller.gui != nil && controller.gui.app != nil {
		preferences := controller.gui.app.Preferences()
		preferences.SetFloat(controller.horizontalKey, controller.horizontalOffset)
		preferences.SetFloat(controller.compactKey, controller.compactOffset)
	}
	return controller.horizontalOffset
}

func normalizeCompactSplit(value float64) float64 {
	if value != value {
		return defaultCompactSplit
	}
	if value < minCompactSplit {
		return minCompactSplit
	}
	if value > maxCompactSplit {
		return maxCompactSplit
	}
	return value
}

func compactWorkspaceSplit(size fyne.Size) bool {
	return size.Width < compactWorkspaceSplitWidth && size.Height >= compactWorkspaceSplitMinHeight
}
