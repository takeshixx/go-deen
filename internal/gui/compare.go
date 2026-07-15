//go:build gui

package gui

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

const (
	compareModeText   = "Text"
	compareModeHex    = "Hex"
	compareModeBase64 = "Base64"
)

var compareModes = []string{compareModeText, compareModeHex, compareModeBase64}

type comparePoint struct {
	Label string
	Data  []byte
}

type compareDifference struct {
	Equal           bool
	LeftBytes       int
	RightBytes      int
	FirstDifference int
}

type compareWorkspace struct {
	gui *DeenGUI

	points      []comparePoint
	leftIndex   int
	rightIndex  int
	leftFinal   bool
	rightFinal  bool
	initialized bool
	updating    bool

	leftSelect      *widget.Select
	rightSelect     *widget.Select
	modeSelect      *widget.Select
	leftMeta        *widget.Label
	rightMeta       *widget.Label
	leftBody        *widget.Entry
	rightBody       *widget.Entry
	difference      *widget.Label
	split           *container.Split
	splitController *responsiveSplitController
	controls        []fyne.Disableable
}

func (dg *DeenGUI) comparePoints() []comparePoint {
	points := []comparePoint{{Label: "Input", Data: dg.pipe.Source()}}
	for i, step := range dg.pipe.Steps() {
		direction := "encode"
		if step.Unprocess {
			direction = "decode"
		}
		points = append(points, comparePoint{
			Label: fmt.Sprintf("Step %d · %s · %s", i+1, plugins.PluginLabel(step.Plugin), direction),
			Data:  dg.pipe.Output(i),
		})
	}
	return points
}

func compareLabels(points []comparePoint) []string {
	labels := make([]string, len(points))
	for i, point := range points {
		labels[i] = point.Label
	}
	return labels
}

func compareIndex(labels []string, label string) int {
	for i, candidate := range labels {
		if candidate == label {
			return i
		}
	}
	return -1
}

func compareData(points []comparePoint, index int) []byte {
	if index < 0 || index >= len(points) {
		return nil
	}
	return points[index].Data
}

func normalizeCompareMode(mode string) string {
	switch mode {
	case compareModeHex:
		return compareModeHex
	case compareModeBase64:
		return compareModeBase64
	default:
		return compareModeText
	}
}

func formatCompareData(data []byte, mode string) string {
	switch normalizeCompareMode(mode) {
	case compareModeHex:
		text, _ := guiHexDisplay(data)
		return text
	case compareModeBase64:
		if pipeline.IsLargeData(data) {
			return pipeline.LargeDataPlaceholder(data) + "\n\nBase64 preview disabled for large data."
		}
		return base64.StdEncoding.EncodeToString(data)
	default:
		text, _ := guiTextDisplay(data)
		return text
	}
}

func compareBytes(left, right []byte) compareDifference {
	difference := compareDifference{
		Equal:           bytes.Equal(left, right),
		LeftBytes:       len(left),
		RightBytes:      len(right),
		FirstDifference: -1,
	}
	if difference.Equal {
		return difference
	}
	limit := min(len(left), len(right))
	for i := 0; i < limit; i++ {
		if left[i] != right[i] {
			difference.FirstDifference = i
			return difference
		}
	}
	difference.FirstDifference = limit
	return difference
}

func (difference compareDifference) Summary() string {
	if difference.Equal {
		return fmt.Sprintf("Identical · %d B", difference.LeftBytes)
	}
	return fmt.Sprintf(
		"Different · %d B vs %d B · first difference at byte offset %d",
		difference.LeftBytes,
		difference.RightBytes,
		difference.FirstDifference,
	)
}

func (dg *DeenGUI) compareTab() fyne.CanvasObject {
	workspace := &compareWorkspace{gui: dg, leftIndex: 0, rightIndex: -1}
	if dg.app != nil {
		preferences := dg.app.Preferences()
		workspace.leftIndex = preferences.IntWithFallback(compareLeftPreferenceKey, 0)
		workspace.rightIndex = preferences.IntWithFallback(compareRightPreferenceKey, -1)
	}

	workspace.leftSelect = widget.NewSelect(nil, workspace.selectLeft)
	workspace.rightSelect = widget.NewSelect(nil, workspace.selectRight)
	workspace.modeSelect = widget.NewSelect(compareModes, workspace.selectMode)
	mode := compareModeText
	if dg.app != nil {
		mode = normalizeCompareMode(dg.app.Preferences().StringWithFallback(compareModePreferenceKey, compareModeText))
	}
	workspace.updating = true
	workspace.modeSelect.SetSelected(mode)
	workspace.updating = false

	workspace.leftMeta = compareMetadataLabel()
	workspace.rightMeta = compareMetadataLabel()
	workspace.leftBody = multilineEntry(14)
	workspace.leftBody.Disable()
	workspace.rightBody = multilineEntry(14)
	workspace.rightBody.Disable()
	workspace.difference = widget.NewLabel("")
	workspace.difference.Wrapping = fyne.TextWrapWord

	swap := widget.NewButtonWithIcon("Swap sides", theme.ViewRefreshIcon(), workspace.swap)
	copyLeft := widget.NewButtonWithIcon("Copy left view", theme.ContentCopyIcon(), func() { workspace.copyView(true) })
	copyRight := widget.NewButtonWithIcon("Copy right view", theme.ContentCopyIcon(), func() { workspace.copyView(false) })
	workspace.controls = []fyne.Disableable{workspace.leftSelect, workspace.rightSelect, workspace.modeSelect, swap, copyLeft, copyRight}

	leftPanel := comparePanel("LEFT", workspace.leftSelect, workspace.leftMeta, workspace.leftBody, copyLeft)
	rightPanel := comparePanel("RIGHT", workspace.rightSelect, workspace.rightMeta, workspace.rightBody, copyRight)
	workspace.split = container.NewHSplit(leftPanel, rightPanel)
	offset := defaultCompareSplit
	if dg.app != nil {
		offset = normalizeCompareSplit(dg.app.Preferences().FloatWithFallback(compareSplitPreferenceKey, defaultCompareSplit))
	}
	workspace.splitController = newResponsiveSplit(
		dg,
		workspace.split,
		compareSplitPreferenceKey,
		compareCompactPreferenceKey,
		offset,
		defaultCompactSplit,
		normalizeCompareSplit,
		compactWorkspaceSplit,
	)

	viewLabel := widget.NewLabel("View")
	viewLabel.Importance = widget.LowImportance
	toolbar := container.NewHBox(viewLabel, workspace.modeSelect, swap)
	status := widget.NewCard("Comparison", "", container.NewPadded(workspace.difference))
	content := container.NewBorder(container.NewVBox(toolbar, status), nil, nil, nil, workspace.splitController.host)
	dg.compareWorkspace = workspace
	workspace.refreshPoints()
	return container.NewPadded(content)
}

func compareMetadataLabel() *widget.Label {
	label := widget.NewLabel("")
	label.Importance = widget.LowImportance
	label.Wrapping = fyne.TextWrapWord
	return label
}

func comparePanel(side string, selection *widget.Select, metadata *widget.Label, body *widget.Entry, copyButton *widget.Button) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(side, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.LowImportance
	header := container.NewVBox(title, selection, metadata)
	footer := container.NewHBox(copyButton)
	return widget.NewCard("", "", container.NewPadded(container.NewBorder(header, footer, nil, nil, body)))
}

func (workspace *compareWorkspace) selectLeft(label string) {
	if workspace.updating || workspace.gui.working {
		return
	}
	workspace.leftIndex = compareIndex(compareLabels(workspace.points), label)
	workspace.leftFinal = workspace.leftIndex > 0 && workspace.leftIndex == len(workspace.points)-1
	workspace.persistSelections()
	workspace.refreshComparison()
}

func (workspace *compareWorkspace) selectRight(label string) {
	if workspace.updating || workspace.gui.working {
		return
	}
	workspace.rightIndex = compareIndex(compareLabels(workspace.points), label)
	workspace.rightFinal = workspace.rightIndex == len(workspace.points)-1
	workspace.persistSelections()
	workspace.refreshComparison()
}

func (workspace *compareWorkspace) selectMode(mode string) {
	if workspace.updating || workspace.gui.working {
		return
	}
	mode = normalizeCompareMode(mode)
	if workspace.gui.app != nil {
		workspace.gui.app.Preferences().SetString(compareModePreferenceKey, mode)
	}
	workspace.refreshComparison()
}

func (workspace *compareWorkspace) refreshPoints() {
	workspace.points = workspace.gui.comparePoints()
	if len(workspace.points) == 0 {
		return
	}
	if workspace.initialized && workspace.leftFinal {
		workspace.leftIndex = len(workspace.points) - 1
	}
	if workspace.initialized && workspace.rightFinal || workspace.rightIndex < 0 {
		workspace.rightIndex = len(workspace.points) - 1
	}
	workspace.leftIndex = min(max(workspace.leftIndex, 0), len(workspace.points)-1)
	workspace.rightIndex = min(max(workspace.rightIndex, 0), len(workspace.points)-1)
	if !workspace.initialized {
		workspace.leftFinal = workspace.leftIndex > 0 && workspace.leftIndex == len(workspace.points)-1
		workspace.rightFinal = workspace.rightIndex == len(workspace.points)-1
		workspace.initialized = true
	}

	labels := compareLabels(workspace.points)
	workspace.updating = true
	workspace.leftSelect.SetOptions(labels)
	workspace.rightSelect.SetOptions(labels)
	workspace.leftSelect.SetSelectedIndex(workspace.leftIndex)
	workspace.rightSelect.SetSelectedIndex(workspace.rightIndex)
	workspace.updating = false
	workspace.persistSelections()
	workspace.refreshComparison()
}

func (workspace *compareWorkspace) refreshComparison() {
	left := compareData(workspace.points, workspace.leftIndex)
	right := compareData(workspace.points, workspace.rightIndex)
	workspace.leftMeta.SetText(pipeline.DataMetadata(left, 0).Summary())
	workspace.rightMeta.SetText(pipeline.DataMetadata(right, 0).Summary())
	mode := normalizeCompareMode(workspace.modeSelect.Selected)
	workspace.leftBody.SetText(formatCompareData(left, mode))
	workspace.rightBody.SetText(formatCompareData(right, mode))
	difference := compareBytes(left, right)
	workspace.difference.SetText(difference.Summary())
	if difference.Equal {
		workspace.difference.Importance = widget.SuccessImportance
	} else {
		workspace.difference.Importance = widget.WarningImportance
	}
	workspace.difference.Refresh()
}

func (workspace *compareWorkspace) swap() {
	if workspace.gui.working {
		return
	}
	workspace.leftIndex, workspace.rightIndex = workspace.rightIndex, workspace.leftIndex
	workspace.leftFinal, workspace.rightFinal = workspace.rightFinal, workspace.leftFinal
	workspace.updating = true
	workspace.leftSelect.SetSelectedIndex(workspace.leftIndex)
	workspace.rightSelect.SetSelectedIndex(workspace.rightIndex)
	workspace.updating = false
	workspace.persistSelections()
	workspace.refreshComparison()
}

func (workspace *compareWorkspace) copyView(left bool) {
	if workspace.gui.working || workspace.gui.window == nil {
		return
	}
	index := workspace.rightIndex
	if left {
		index = workspace.leftIndex
	}
	workspace.gui.window.Clipboard().SetContent(formatCompareData(compareData(workspace.points, index), workspace.modeSelect.Selected))
	if left {
		workspace.gui.showActionFeedback("Left comparison view copied")
	} else {
		workspace.gui.showActionFeedback("Right comparison view copied")
	}
}

func (workspace *compareWorkspace) persistSelections() {
	if workspace.gui.app == nil {
		return
	}
	preferences := workspace.gui.app.Preferences()
	preferences.SetInt(compareLeftPreferenceKey, workspace.leftIndex)
	preferences.SetInt(compareRightPreferenceKey, workspace.rightIndex)
}

func (workspace *compareWorkspace) setDisabled(disabled bool) {
	for _, control := range workspace.controls {
		if disabled {
			control.Disable()
		} else {
			control.Enable()
		}
	}
}
