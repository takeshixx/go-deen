//go:build gui

package gui

import "fyne.io/fyne/v2"

const (
	defaultSourcePane   = "Editor"
	defaultAddSplit     = 0.42
	minAddSplit         = 0.30
	maxAddSplit         = 0.65
	defaultCompareSplit = 0.50
	minCompareSplit     = 0.30
	maxCompareSplit     = 0.70
	defaultSourceView   = "Hex"
	defaultWindowWidth  = 1180
	defaultWindowHeight = 800
	minimumWindowWidth  = 720
	minimumWindowHeight = 520
)

func normalizeSourcePane(value string) string {
	if value == "Inspector" {
		return "Inspector"
	}
	return defaultSourcePane
}

func normalizeAddSplit(value float64) float64 {
	if value != value {
		return defaultAddSplit
	}
	if value < minAddSplit {
		return minAddSplit
	}
	if value > maxAddSplit {
		return maxAddSplit
	}
	return value
}

func normalizeCompareSplit(value float64) float64 {
	if value != value {
		return defaultCompareSplit
	}
	if value < minCompareSplit {
		return minCompareSplit
	}
	if value > maxCompareSplit {
		return maxCompareSplit
	}
	return value
}

func normalizeWorkspace(index int) int {
	if index < 0 || index > 4 {
		return 0
	}
	return index
}

func normalizeWindowDimension(value, fallback, minimum float64) float32 {
	if value != value || value < minimum || value > 10_000 {
		return float32(fallback)
	}
	return float32(value)
}

func (dg *DeenGUI) restoredWindowSize() fyne.Size {
	if dg.app == nil {
		return fyne.NewSize(defaultWindowWidth, defaultWindowHeight)
	}
	prefs := dg.app.Preferences()
	width := normalizeWindowDimension(prefs.FloatWithFallback(windowWidthPreferenceKey, defaultWindowWidth), defaultWindowWidth, minimumWindowWidth)
	height := normalizeWindowDimension(prefs.FloatWithFallback(windowHeightPreferenceKey, defaultWindowHeight), defaultWindowHeight, minimumWindowHeight)
	return fyne.NewSize(width, height)
}

func (dg *DeenGUI) rememberAddSplit() {
	dg.rememberTransformerCatalogSplit(dg.addCatalog)
}

func (dg *DeenGUI) rememberTransformerCatalogSplit(catalog *transformerCatalog) {
	if catalog == nil || catalog.split == nil || dg.app == nil {
		return
	}
	if catalog.splitController != nil {
		catalog.splitController.remember()
		return
	}
	offset := normalizeAddSplit(catalog.split.Offset)
	if offset != catalog.split.Offset {
		catalog.split.SetOffset(offset)
	}
	dg.app.Preferences().SetFloat(catalog.splitPreferenceKey, offset)
}

func (dg *DeenGUI) rememberWorkspaceState() {
	dg.rememberStepEditorSplit()
	dg.rememberAddSplit()
	dg.rememberTransformerCatalogSplit(dg.browserCatalog)
	if dg.workflowLibrary != nil && dg.workflowLibrary.split != nil && dg.app != nil {
		if dg.workflowLibrary.splitController != nil {
			dg.workflowLibrary.splitController.remember()
		} else {
			offset := normalizeAddSplit(dg.workflowLibrary.split.Offset)
			if offset != dg.workflowLibrary.split.Offset {
				dg.workflowLibrary.split.SetOffset(offset)
			}
			dg.app.Preferences().SetFloat(workflowSplitPreferenceKey, offset)
		}
	}
	if dg.compareWorkspace != nil && dg.compareWorkspace.split != nil && dg.app != nil {
		if dg.compareWorkspace.splitController != nil {
			dg.compareWorkspace.splitController.remember()
		} else {
			offset := normalizeCompareSplit(dg.compareWorkspace.split.Offset)
			if offset != dg.compareWorkspace.split.Offset {
				dg.compareWorkspace.split.SetOffset(offset)
			}
			dg.app.Preferences().SetFloat(compareSplitPreferenceKey, offset)
		}
	}
	if dg.app == nil {
		return
	}
	dg.app.Preferences().SetInt(workspacePreferenceKey, normalizeWorkspace(dg.activeTab))
	dg.app.Preferences().SetInt(selectedStagePreferenceKey, dg.selectedStage)
	if dg.window == nil {
		return
	}
	size := dg.window.Canvas().Size()
	if size.Width >= minimumWindowWidth {
		dg.app.Preferences().SetFloat(windowWidthPreferenceKey, float64(size.Width))
	}
	if size.Height >= minimumWindowHeight {
		dg.app.Preferences().SetFloat(windowHeightPreferenceKey, float64(size.Height))
	}
}
