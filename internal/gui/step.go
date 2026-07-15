//go:build gui

package gui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	neturl "net/url"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

// stepPalette gives adjacent steps decorative visual separation. These colours
// do not encode transform type or state; labels, icons, and controls do that.
var stepPalette = []color.NRGBA{
	{0x42, 0x85, 0xf4, 0xff}, // blue
	{0x0f, 0x9d, 0x58, 0xff}, // green
	{0xf4, 0xb4, 0x00, 0xff}, // amber
	{0xdb, 0x44, 0x37, 0xff}, // red
	{0xab, 0x47, 0xbc, 0xff}, // purple
	{0x00, 0xac, 0xc1, 0xff}, // cyan
}

func accent(i int) color.NRGBA { return stepPalette[i%len(stepPalette)] }

func disabledAccent() color.NRGBA { return color.NRGBA{R: 0x8c, G: 0x96, B: 0x9b, A: 0xff} }

const compactControlMinWidth float32 = 360
const focusedStepEditorHeight float32 = 520
const defaultStepEditorSplit = 0.40
const minStepEditorSplit = 0.28
const maxStepEditorSplit = 0.62
const sourceInputRows = 14

type fixedHeightLayout struct {
	height float32
}

func (l fixedHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		obj.Resize(size)
	}
}

func (l fixedHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(compactControlMinWidth, l.height)
}

type cappedMinWidthLayout struct {
	width float32
}

func (l cappedMinWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		obj.Resize(size)
	}
}

func (l cappedMinWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width, height float32
	for _, obj := range objects {
		min := obj.MinSize()
		if min.Width > width {
			width = min.Width
		}
		if min.Height > height {
			height = min.Height
		}
	}
	if l.width > 0 && width > l.width {
		width = l.width
	}
	return fyne.NewSize(width, height)
}

func newStepSurface(content fyne.CanvasObject, accentColor color.NRGBA, disabled bool) fyne.CanvasObject {
	backgroundColor := theme.Color(theme.ColorNameButton)
	if disabled {
		backgroundColor = theme.Color(theme.ColorNameDisabledButton)
		accentColor = disabledAccent()
	}
	background := canvas.NewRectangle(backgroundColor)
	background.CornerRadius = theme.Size(theme.SizeNameCardRadius)
	background.Shadow = canvas.Shadow{
		Color:      theme.Color(theme.ColorNameShadow),
		BlurRadius: 12,
		Spread:     -2,
		Offset:     fyne.NewPos(0, 3),
		Variant:    canvas.BoxShadow,
	}

	rail := canvas.NewRectangle(accentColor)
	rail.CornerRadius = 2
	rail.SetMinSize(fyne.NewSize(4, 1))
	body := container.NewBorder(nil, nil, rail, nil, container.NewPadded(content))
	return container.NewPadded(container.NewStack(background, body))
}

// multilineEntry returns a word-wrapping multi-line entry with a readable
// minimum height.
func multilineEntry(rows int) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapBreak
	e.SetMinRowsVisible(rows)
	return e
}

func guiTextDisplay(data []byte) (string, bool) {
	return guiTextDisplayMode(data, false)
}

func guiTextDisplayMode(data []byte, full bool) (string, bool) {
	if full {
		return pipeline.TextDisplayFull(data), true
	}
	if pipeline.IsLargeData(data) {
		return pipeline.LargeDataPlaceholder(data) + "\n\nPreview disabled in the desktop GUI for large data.", true
	}
	return pipeline.TextDisplay(data)
}

func guiHexDisplay(data []byte) (string, bool) {
	return guiHexDisplayMode(data, false)
}

func guiHexDisplayMode(data []byte, full bool) (string, bool) {
	if full {
		return pipeline.HexDisplayFull(data), true
	}
	if pipeline.IsLargeData(data) {
		return pipeline.LargeDataPlaceholder(data) + "\n\nHex preview disabled in the desktop GUI for large data.", true
	}
	return pipeline.HexDisplay(data)
}

func guiStringsDisplay(data []byte) (string, bool) {
	return guiStringsDisplayMode(data, false)
}

func guiStringsDisplayMode(data []byte, full bool) (string, bool) {
	if full {
		return pipeline.StringsDisplayFull(data), true
	}
	if pipeline.IsLargeData(data) {
		return pipeline.LargeDataPlaceholder(data) + "\n\nStrings preview disabled in the desktop GUI for large data.", true
	}
	return pipeline.StringsDisplay(data)
}

func pluginSelectLabels(category string) (labels []string, labelToName, nameToLabel map[string]string) {
	labelToName = map[string]string{}
	nameToLabel = map[string]string{}
	for _, name := range plugins.InCategory(category) {
		label := plugins.PluginLabel(name)
		labelToName[label] = name
		nameToLabel[name] = label
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		return strings.ToLower(labels[i]) < strings.ToLower(labels[j])
	})
	return labels, labelToName, nameToLabel
}

// categorySelectors builds compact category and transformer dropdowns.
func (dg *DeenGUI) categorySelectors(current string, onPick func(name string)) *fyne.Container {
	categoryToID := map[string]string{}
	var categoryLabels []string
	for _, category := range plugins.PluginCategories {
		label := plugins.CategoryLabel(category)
		categoryToID[label] = category
		categoryLabels = append(categoryLabels, label)
	}

	var transformerByLabel map[string]string
	updatingSelectors := false
	transformerSelect := widget.NewSelect(nil, func(label string) {
		if updatingSelectors || label == "" || transformerByLabel == nil {
			return
		}
		name := transformerByLabel[label]
		if name != "" {
			onPick(name)
		}
	})
	transformerSelect.PlaceHolder = "Select transformer"
	transformerSelect.Disable()

	setCategory := func(category, selectedPlugin string) {
		labels, labelToName, nameToLabel := pluginSelectLabels(category)
		transformerByLabel = labelToName
		updatingSelectors = true
		transformerSelect.Options = labels
		transformerSelect.Selected = ""
		if selectedPlugin != "" {
			transformerSelect.Selected = nameToLabel[selectedPlugin]
		}
		transformerSelect.PlaceHolder = plugins.CategorySelectLabel(category)
		transformerSelect.Enable()
		transformerSelect.Refresh()
		updatingSelectors = false
	}

	categorySelect := widget.NewSelect(categoryLabels, func(label string) {
		category := categoryToID[label]
		if category != "" {
			setCategory(category, "")
		}
	})
	categorySelect.PlaceHolder = "Select category"

	if current != "" {
		if category := plugins.CategoryOf(current); category != "" {
			categorySelect.Selected = plugins.CategoryLabel(category)
			categorySelect.Refresh()
			setCategory(category, current)
		}
	}
	dg.registerWorkControl(categorySelect)
	dg.registerWorkControl(transformerSelect)

	return container.NewVBox(categorySelect, transformerSelect)
}

func normalizeStepEditorSplit(value float64) float64 {
	if value != value { // NaN cannot be meaningfully restored from preferences.
		return defaultStepEditorSplit
	}
	if value < minStepEditorSplit {
		return minStepEditorSplit
	}
	if value > maxStepEditorSplit {
		return maxStepEditorSplit
	}
	return value
}

// rememberStepEditorSplit captures Fyne's live divider position before a
// focused card is discarded. Split does not expose a drag callback, so the
// offset is persisted on rebuild and once more when Run returns.
func (dg *DeenGUI) rememberStepEditorSplit() {
	for _, card := range dg.cards {
		if card == nil || card.editorSplit == nil {
			continue
		}
		offset := normalizeStepEditorSplit(card.editorSplit.Offset)
		if card.editorSplitController != nil {
			offset = card.editorSplitController.remember()
		}
		if offset == dg.stepEditorSplit {
			return
		}
		dg.stepEditorSplit = offset
		if dg.app != nil {
			dg.app.Preferences().SetFloat(stepEditorSplitPreferenceKey, offset)
		}
		return
	}
}

// newSourceCard builds the editable source-input card at the top of the chain.
func (dg *DeenGUI) newSourceCard() fyne.CanvasObject {
	dg.sourceEntry = multilineEntry(sourceInputRows)
	rawNeedsFull, hexNeedsFull, stringsNeedsFull := dg.sourceNeedsFull()
	if !rawNeedsFull {
		dg.sourceFullRaw = false
	}
	if !hexNeedsFull {
		dg.sourceFullHex = false
	}
	if !stringsNeedsFull {
		dg.sourceFullStrings = false
	}
	sourceText, sourceCapped := guiTextDisplayMode(dg.pipe.Source(), dg.sourceFullRaw)
	dg.sourceEntry.SetText(sourceText)
	if sourceCapped {
		dg.sourceEntry.Disable()
	}
	dg.registerWorkControl(dg.sourceEntry)
	dg.sourceRaw = multilineEntry(sourceInputRows)
	dg.sourceRaw.Disable()
	dg.sourceHex = multilineEntry(sourceInputRows)
	dg.sourceHex.Disable()
	dg.sourceStrings = multilineEntry(sourceInputRows)
	dg.sourceStrings.Disable()
	rawText, _ := guiTextDisplayMode(dg.pipe.Source(), dg.sourceFullRaw)
	hexText, _ := guiHexDisplayMode(dg.pipe.Source(), dg.sourceFullHex)
	stringsText, _ := guiStringsDisplayMode(dg.pipe.Source(), dg.sourceFullStrings)
	dg.sourceRaw.SetText(rawText)
	dg.sourceHex.SetText(hexText)
	dg.sourceStrings.SetText(stringsText)

	dg.sourceMeta = widget.NewLabel(dg.sourceMetadataSummary())
	dg.sourceMeta.Importance = widget.LowImportance
	dg.sourceMeta.Wrapping = fyne.TextWrapBreak
	dg.sourceMeta.TextStyle.Monospace = true
	dg.sourceEntry.OnChanged = func(s string) {
		if dg.updating || dg.working {
			return
		}
		if pipeline.IsLargeData(dg.pipe.Source()) {
			return
		}
		dg.sourceName = ""
		dg.clearSourceFullViews()
		dg.pipe.SetSourceOwned([]byte(s))
		dg.refreshFrom(0)
	}

	tabs := []*container.TabItem{
		container.NewTabItem("Raw", dg.sourceRaw),
		container.NewTabItem("Hex", dg.sourceHex),
		container.NewTabItem("Strings", dg.sourceStrings),
	}
	if pipeline.HasStructuredPreview(dg.pipe.Source()) {
		dg.sourcePreview = newPreviewGrid()
		preview, spans, _ := pipeline.HighlightedPreview(dg.pipe.Source())
		setPreviewText(dg.sourcePreview, preview, spans)
		dg.sourcePreviewTab = container.NewTabItem("Preview", dg.sourcePreview)
		tabs = append(tabs, dg.sourcePreviewTab)
	}
	dg.sourceViewer = container.NewAppTabs(tabs...)
	dg.sourceViewer.SetTabLocation(container.TabLocationTop)
	selectAppTab(dg.sourceViewer, dg.sourceView, defaultSourceView)
	dg.sourceViewer.OnSelected = func(tab *container.TabItem) {
		dg.sourceView = tab.Text
		if dg.app != nil {
			dg.app.Preferences().SetString(sourceViewPreferenceKey, tab.Text)
		}
	}

	dg.sourceFullControls = container.NewHBox()
	dg.refreshSourceFullControls(rawNeedsFull, hexNeedsFull, stringsNeedsFull)
	editorHint := lowImportanceLabel("Type or paste text here. You can also drop a file anywhere in the window.")
	editorPane := widget.NewCard("Editable source", "", container.NewBorder(editorHint, nil, nil, nil, dg.sourceEntry))
	inspectorBottom := container.NewVBox(dg.sourceFullControls, dg.sourceMeta)
	inspectorPane := widget.NewCard("Inspector", "", container.NewBorder(nil, inspectorBottom, nil, nil, dg.sourceViewer))
	dg.sourceWorkspace = container.NewAppTabs(
		container.NewTabItem("Editor", editorPane),
		container.NewTabItem("Inspector", inspectorPane),
	)
	dg.sourceWorkspace.SetTabLocation(container.TabLocationTop)
	dg.sourcePane = normalizeSourcePane(dg.sourcePane)
	selectAppTab(dg.sourceWorkspace, dg.sourcePane, defaultSourcePane)
	dg.sourceWorkspace.OnSelected = func(tab *container.TabItem) {
		dg.sourcePane = normalizeSourcePane(tab.Text)
		if dg.app != nil {
			dg.app.Preferences().SetString(sourcePanePreferenceKey, dg.sourcePane)
		}
	}
	content := container.New(fixedHeightLayout{height: focusedStepEditorHeight}, dg.sourceWorkspace)
	return widget.NewCard("Input", "Edit or inspect the pipeline input one focused view at a time.", content)
}

func (dg *DeenGUI) syncSourcePreviewTab() {
	if dg.sourceViewer == nil {
		return
	}
	hasPreview := pipeline.HasStructuredPreview(dg.pipe.Source())
	if hasPreview && dg.sourcePreviewTab == nil {
		dg.sourcePreview = newPreviewGrid()
		dg.sourcePreviewTab = container.NewTabItem("Preview", dg.sourcePreview)
		dg.sourceViewer.Append(dg.sourcePreviewTab)
		dg.sourceViewer.Select(dg.sourcePreviewTab)
		return
	}
	if !hasPreview && dg.sourcePreviewTab != nil {
		if dg.sourceViewer.Selected() == dg.sourcePreviewTab {
			selectAppTab(dg.sourceViewer, dg.sourceView, defaultSourceView)
			if dg.sourceViewer.Selected() == dg.sourcePreviewTab {
				dg.sourceViewer.SelectIndex(1)
			}
		}
		dg.sourceViewer.Remove(dg.sourcePreviewTab)
		dg.sourcePreviewTab = nil
		dg.sourcePreview = nil
	}
}

func selectAppTab(tabs *container.AppTabs, preferred, fallback string) {
	for _, name := range []string{preferred, fallback, "Raw"} {
		for _, item := range tabs.Items {
			if item.Text == name {
				tabs.Select(item)
				return
			}
		}
	}
}

func (dg *DeenGUI) sourceNeedsFull() (raw, hexView, stringsView bool) {
	_, raw = guiTextDisplayMode(dg.pipe.Source(), false)
	_, hexView = guiHexDisplayMode(dg.pipe.Source(), false)
	_, stringsView = guiStringsDisplayMode(dg.pipe.Source(), false)
	return raw, hexView, stringsView
}

func (dg *DeenGUI) clearSourceFullViews() {
	dg.sourceFullRaw = false
	dg.sourceFullHex = false
	dg.sourceFullStrings = false
}

func (dg *DeenGUI) refreshSourceFullControls(rawCapped, hexCapped, stringsCapped bool) {
	if dg.sourceFullControls == nil {
		return
	}
	dg.sourceFullControls.RemoveAll()
	add := func(label string, enabled bool, setFull func()) {
		if !enabled {
			return
		}
		button := widget.NewButton(label, func() {
			if dg.working {
				return
			}
			dialog.ShowConfirm(
				label+"?",
				"Rendering the full input view can use a lot of memory and may make the interface slow for large files.",
				func(ok bool) {
					if !ok || dg.working {
						return
					}
					setFull()
					dg.refreshFrom(0)
				},
				dg.window,
			)
		})
		button.Importance = widget.LowImportance
		dg.sourceFullControls.Add(button)
	}
	add("Show full Raw", rawCapped && !dg.sourceFullRaw, func() { dg.sourceFullRaw = true })
	if pipeline.IsBinaryData(dg.pipe.Source()) {
		add("Show full Hex", hexCapped && !dg.sourceFullHex, func() { dg.sourceFullHex = true })
		add("Show full Strings", stringsCapped && !dg.sourceFullStrings, func() { dg.sourceFullStrings = true })
	}
	if dg.sourceFullRaw || dg.sourceFullHex || dg.sourceFullStrings {
		notice := widget.NewLabel("Full input view enabled; input is read-only.")
		notice.Importance = widget.WarningImportance
		dg.sourceFullControls.Add(notice)
	}
	if len(dg.sourceFullControls.Objects) == 0 {
		dg.sourceFullControls.Hide()
	} else {
		dg.sourceFullControls.Show()
	}
	dg.sourceFullControls.Refresh()
}

// stepCard is the view for a single pipeline step.
type stepCard struct {
	gui        *DeenGUI
	index      int
	pluginName string
	collapsed  bool

	decode                *widget.Check
	summary               *canvas.Text
	collapse              *widget.Button
	headerActions         []*widget.Button
	detail                *fyne.Container
	options               *fyne.Container
	fullControls          *fyne.Container
	body                  *widget.Entry
	hexBody               *widget.Entry
	stringsBody           *widget.Entry
	viewer                *container.AppTabs
	editorSplit           *container.Split
	editorSplitController *responsiveSplitController
	rawTab                *container.TabItem
	hexTab                *container.TabItem
	stringsTab            *container.TabItem
	previewTab            *container.TabItem
	preview               *widget.TextGrid
	image                 *canvas.Image
	imageMsg              *widget.Label
	meta                  *widget.Label
	status                *widget.Label
	container             fyne.CanvasObject
	fullRaw               bool
	fullHex               bool
	fullStrings           bool
}

func (dg *DeenGUI) newStepCard(i int) *stepCard {
	step := dg.pipe.Steps()[i]
	c := &stepCard{gui: dg, index: i, pluginName: step.Plugin}
	col := accent(i)
	canDecode := plugins.CanDecode(step.Plugin)

	c.decode = widget.NewCheck("decode", nil)
	c.decode.SetChecked(step.Unprocess && canDecode)
	dg.registerWorkControl(c.decode)

	apply := func() {
		if c.pluginName == "" {
			return
		}
		decode := c.decode.Checked && plugins.CanDecode(c.pluginName)
		dg.runPipelineWork("Processing", func() error {
			dg.pipe.SetPlugin(c.index, c.pluginName, decode)
			return nil
		}, dg.rebuild)
	}
	selectors := dg.categorySelectors(step.Plugin, func(name string) {
		c.pluginName = name
		apply()
	})
	c.decode.OnChanged = func(bool) { apply() }
	toggleEnabled := func() {
		dg.toggleStep(c.index)
	}

	// Title row: collapse toggle, coloured title, active-plugin summary, remove.
	displayCol := col
	if step.Disabled {
		displayCol = disabledAccent()
	}
	c.summary = canvas.NewText("", displayCol)
	c.summary.TextStyle = fyne.TextStyle{Bold: true}
	title := canvas.NewText(fmt.Sprintf("Step %d", i+1), displayCol)
	title.TextStyle = fyne.TextStyle{Bold: true}
	c.collapse = stepIconButton("Collapse step", theme.MenuDropDownIcon(), c.toggleCollapse)
	moveUp := stepIconButton("Move step up", theme.MoveUpIcon(), func() {
		dg.moveStep(c.index, -1)
	})
	if i == 0 {
		moveUp.Disable()
	}
	moveDown := stepIconButton("Move step down", theme.MoveDownIcon(), func() {
		dg.moveStep(c.index, 1)
	})
	if i == dg.pipe.Len()-1 {
		moveDown.Disable()
	}
	duplicate := stepIconButton("Duplicate step", theme.ContentCopyIcon(), func() {
		dg.duplicateStep(c.index)
	})
	remove := stepIconButton("Remove step", theme.DeleteIcon(), func() {
		dg.removeStep(c.index)
	})
	enabledIcon := theme.VisibilityIcon()
	enabledLabel := "Disable step"
	if step.Disabled {
		enabledIcon = theme.VisibilityOffIcon()
		enabledLabel = "Enable step"
	}
	enabledControl := stepIconButton(enabledLabel, enabledIcon, toggleEnabled)
	c.headerActions = []*widget.Button{c.collapse, enabledControl, moveUp, moveDown, duplicate, remove}
	for _, control := range c.headerActions {
		dg.registerWorkControl(control)
	}
	titleRow := container.NewBorder(nil, nil,
		container.NewHBox(c.collapse, title, c.summary),
		container.NewHBox(enabledControl, moveUp, moveDown, duplicate, remove))

	// Detail: selectors, toggles, options, output, errors.
	c.options = container.NewVBox()
	c.body = multilineEntry(6)
	dg.registerWorkControl(c.body)
	c.body.OnChanged = func(s string) {
		if dg.updating || dg.working {
			return
		}
		dg.pipe.EditOutput(c.index, []byte(s))
		dg.refreshFrom(c.index + 1)
	}
	c.hexBody = multilineEntry(6)
	c.hexBody.Disable()
	c.stringsBody = multilineEntry(6)
	c.stringsBody.Disable()
	c.rawTab = container.NewTabItem("Raw", c.body)
	c.hexTab = container.NewTabItem("Hex", c.hexBody)
	viewerTabs := []*container.TabItem{
		c.rawTab,
		c.hexTab,
	}
	if pipeline.IsBinaryData(dg.pipe.Output(i)) {
		c.stringsTab = container.NewTabItem("Strings", c.stringsBody)
		viewerTabs = append(viewerTabs, c.stringsTab)
	}
	if pipeline.HasStructuredPreview(dg.pipe.Output(i)) {
		c.preview = newPreviewGrid()
		c.previewTab = container.NewTabItem("Preview", c.preview)
		viewerTabs = append(viewerTabs, c.previewTab)
	}
	if stepGeneratesImage(step) {
		c.image = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
		c.image.FillMode = canvas.ImageFillContain
		c.imageMsg = widget.NewLabel("No image preview available.")
		c.imageMsg.Alignment = fyne.TextAlignCenter
		c.image.Hide()
		imageViewer := container.NewBorder(nil, c.imageMsg, nil, nil, container.NewPadded(c.image))
		viewerTabs = append(viewerTabs, container.NewTabItem("Image", imageViewer))
	}
	viewer := container.NewAppTabs(viewerTabs...)
	viewer.SetTabLocation(container.TabLocationTop)
	c.viewer = viewer
	fallbackView := "Raw"
	if c.previewTab != nil {
		fallbackView = "Preview"
	} else if pipeline.IsBinaryData(dg.pipe.Output(i)) {
		fallbackView = "Hex"
	} else if stepGeneratesImage(step) {
		fallbackView = "Image"
	}
	selectAppTab(viewer, dg.stepOutputView, fallbackView)
	viewer.OnSelected = func(tab *container.TabItem) {
		dg.stepOutputView = tab.Text
		if dg.app != nil {
			dg.app.Preferences().SetString(stepOutputPreferenceKey, tab.Text)
		}
	}
	c.fullControls = container.NewHBox()
	c.meta = widget.NewLabel("")
	c.meta.Importance = widget.LowImportance
	c.meta.Wrapping = fyne.TextWrapBreak
	c.meta.TextStyle.Monospace = true
	c.status = widget.NewLabel("")
	c.status.Importance = widget.DangerImportance
	c.status.Wrapping = fyne.TextWrapBreak
	c.status.Hide()
	toggles := container.NewHBox()
	if canDecode {
		toggles.Add(stepToggleControl(c.decode, "Mode", col))
	}
	configuration := container.NewVBox(selectors, toggles, c.options)
	configurationPane := widget.NewCard("Configuration", "", container.NewVScroll(configuration))
	outputDetails := container.NewVBox(c.fullControls, c.meta, c.status)
	outputPane := widget.NewCard("Output", "", container.NewBorder(nil, outputDetails, nil, nil, viewer))
	c.editorSplit = container.NewHSplit(configurationPane, outputPane)
	c.editorSplitController = newResponsiveSplit(
		dg,
		c.editorSplit,
		stepEditorSplitPreferenceKey,
		stepEditorCompactPreferenceKey,
		dg.stepEditorSplit,
		defaultCompactSplit,
		normalizeStepEditorSplit,
		func(fyne.Size) bool { return dg.compactStages },
	)
	c.detail = container.New(fixedHeightLayout{height: focusedStepEditorHeight}, c.editorSplitController.host)

	inner := container.NewVBox(titleRow, c.detail)
	c.container = newStepSurface(inner, displayCol, step.Disabled)

	c.rebuildOptions()
	c.updateSummary()
	if !c.collapsed {
		c.refresh()
	}
	return c
}

// toggleCollapse hides or shows the step's detail section.
func (c *stepCard) toggleCollapse() {
	c.collapsed = !c.collapsed
	if c.collapsed {
		c.detail.Hide()
		c.collapse.SetIcon(stepActionIcon("Expand step", theme.NavigateNextIcon()))
	} else {
		c.refresh()
		c.detail.Show()
		c.collapse.SetIcon(stepActionIcon("Collapse step", theme.MenuDropDownIcon()))
	}
}

// updateSummary refreshes the "category / plugin · direction" highlight.
func (c *stepCard) updateSummary() {
	name := c.pluginName
	if name == "" {
		c.summary.Text = "  (no transform)"
		c.summary.Refresh()
		return
	}
	dir := "encode"
	if c.decode.Checked {
		dir = "decode"
	}
	cat := plugins.CategoryOf(name)
	c.summary.Text = fmt.Sprintf("  %s / %s · %s", plugins.CategoryLabel(cat), plugins.PluginLabel(name), dir)
	if c.gui.pipe.Steps()[c.index].Disabled {
		c.summary.Text += " · disabled"
	}
	c.summary.Refresh()
}

// rebuildOptions repopulates the per-plugin option widgets (as a form).
func (c *stepCard) rebuildOptions() {
	c.options.RemoveAll()
	step := c.gui.pipe.Steps()[c.index]
	opts := pipeline.PluginOptions(step.Plugin)
	if len(opts) == 0 {
		c.options.Refresh()
		return
	}
	var checkOptions []fyne.CanvasObject
	var fieldOptions []fyne.CanvasObject
	var secretOptions []fyne.CanvasObject
	for _, opt := range opts {
		opt := opt
		label := widget.NewLabelWithStyle(opt.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		var control fyne.CanvasObject
		var target *[]fyne.CanvasObject
		if opt.IsBool {
			chk := widget.NewCheck("", nil)
			chk.SetChecked(step.Options[opt.Name] == "true")
			chk.OnChanged = func(b bool) {
				val := "false"
				if b {
					val = "true"
				}
				c.gui.runPipelineWork("Processing", func() error {
					c.gui.pipe.SetOption(c.index, opt.Name, val)
					return nil
				}, func() { c.gui.refreshFrom(c.index) })
			}
			control = chk
			target = &checkOptions
		} else if opt.Kind == "select" {
			selectInput := widget.NewSelect(opt.Choices, nil)
			if v, ok := step.Options[opt.Name]; ok {
				selectInput.SetSelected(v)
			} else {
				selectInput.SetSelected(opt.Default)
			}
			selectInput.OnChanged = func(s string) {
				c.gui.runPipelineWork("Processing", func() error {
					c.gui.pipe.SetOption(c.index, opt.Name, s)
					return nil
				}, func() { c.gui.refreshFrom(c.index) })
			}
			control = selectInput
			target = &fieldOptions
		} else {
			entry := widget.NewEntry()
			if opt.Multiline {
				entry = multilineEntry(3)
			}
			if opt.Kind == "secret" || opt.Secret {
				entry = widget.NewPasswordEntry()
			}
			entry.Validator = optionEntryValidator(step.Plugin, opt)
			entry.AlwaysShowValidationError = entry.Validator != nil
			entry.SetPlaceHolder(optionPlaceholder(opt))
			if v, ok := step.Options[opt.Name]; ok {
				entry.SetText(v)
			}
			entry.OnChanged = func(s string) {
				if c.gui.working {
					return
				}
				if entry.Validator != nil && entry.Validator(s) != nil {
					return
				}
				c.gui.pipe.SetOption(c.index, opt.Name, s)
				c.gui.refreshFrom(c.index)
			}
			control = entry
			if opt.Kind == "secret" || opt.Secret {
				target = &secretOptions
			} else {
				target = &fieldOptions
			}
		}
		*target = append(*target, optionBlock(label, control, opt))
		if disableable, ok := control.(fyne.Disableable); ok {
			c.gui.registerWorkControl(disableable)
		}
	}
	if len(checkOptions) > 0 {
		c.options.Add(optionSection("Behavior", checkOptions))
	}
	if len(fieldOptions) > 0 {
		c.options.Add(optionSection("Values", fieldOptions))
	}
	if len(secretOptions) > 0 {
		c.options.Add(optionSection("Sensitive values", secretOptions))
	}
	c.options.Refresh()
}

func optionEntryValidator(plugin string, opt pipeline.Option) fyne.StringValidator {
	if opt.Kind != "number" {
		return nil
	}
	// Arithmetic operands intentionally accept decimal, hex, or one character.
	if (plugin == "add" || plugin == "sub" || plugin == "xor") && opt.Name == "value" {
		return nil
	}
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("enter an integer")
		}
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("enter an integer")
		}
		return nil
	}
}

func optionPlaceholder(opt pipeline.Option) string {
	if opt.Default == "" {
		return opt.Label
	}
	return fmt.Sprintf("default: %s", opt.Default)
}

func optionBlock(label, control fyne.CanvasObject, opt pipeline.Option) fyne.CanvasObject {
	items := []fyne.CanvasObject{label, control}
	if help := optionHelp(opt); help != nil {
		items = append(items, help)
	}
	return container.NewVBox(items...)
}

func optionSection(title string, items []fyne.CanvasObject) fyne.CanvasObject {
	heading := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	heading.Importance = widget.LowImportance
	children := []fyne.CanvasObject{heading}
	children = append(children, items...)
	return container.NewVBox(children...)
}

func optionHelp(opt pipeline.Option) fyne.CanvasObject {
	var items []fyne.CanvasObject
	if opt.Description != "" {
		desc := widget.NewLabel(opt.Description)
		desc.Wrapping = fyne.TextWrapWord
		items = append(items, desc)
	}
	if opt.Secret || opt.Kind == "secret" {
		warning := widget.NewLabel("Sensitive value: masked on screen, but included as plaintext when this chain is saved.")
		warning.Importance = widget.WarningImportance
		warning.Wrapping = fyne.TextWrapWord
		items = append(items, warning)
	}
	if opt.HelpURL != "" {
		u, err := neturl.Parse(opt.HelpURL)
		if err == nil {
			label := opt.HelpLabel
			if label == "" {
				label = "Reference"
			}
			items = append(items, widget.NewHyperlink(label, u))
		}
	}
	if len(items) == 0 {
		return nil
	}
	return container.NewVBox(items...)
}

type namedThemedResource struct {
	fyne.Resource
	label string
}

func (r namedThemedResource) Name() string { return r.label }

func (r namedThemedResource) ThemeColorName() fyne.ThemeColorName {
	if themed, ok := r.Resource.(fyne.ThemedResource); ok {
		return themed.ThemeColorName()
	}
	return theme.ColorNameForeground
}

func stepActionIcon(label string, icon fyne.Resource) fyne.Resource {
	return namedThemedResource{Resource: icon, label: label}
}

func stepIconButton(label string, icon fyne.Resource, tapped func()) *widget.Button {
	button := widget.NewButtonWithIcon("", stepActionIcon(label, icon), tapped)
	button.Importance = widget.LowImportance
	return button
}

func stepGeneratesImage(step *pipeline.Step) bool {
	return step.Plugin == "qr" && !step.Unprocess
}

func newPreviewGrid() *widget.TextGrid {
	preview := widget.NewTextGrid()
	preview.ShowLineNumbers = false
	preview.Scroll = fyne.ScrollBoth
	return preview
}

func (c *stepCard) syncPreviewTab(out []byte) {
	if c.viewer == nil {
		return
	}
	hasPreview := pipeline.HasStructuredPreview(out)
	if hasPreview && c.previewTab == nil {
		c.preview = newPreviewGrid()
		c.previewTab = container.NewTabItem("Preview", c.preview)
		c.viewer.Append(c.previewTab)
		c.viewer.Select(c.previewTab)
		return
	}
	if !hasPreview && c.previewTab != nil {
		if c.viewer.Selected() == c.previewTab {
			if pipeline.IsBinaryData(out) {
				c.viewer.Select(c.hexTab)
			} else {
				c.viewer.Select(c.rawTab)
			}
		}
		c.viewer.Remove(c.previewTab)
		c.previewTab = nil
		c.preview = nil
	}
}

func (c *stepCard) syncStringsTab(out []byte) {
	if c.viewer == nil {
		return
	}
	hasStrings := pipeline.IsBinaryData(out)
	if hasStrings && c.stringsTab == nil {
		c.stringsTab = container.NewTabItem("Strings", c.stringsBody)
		c.viewer.Append(c.stringsTab)
		return
	}
	if !hasStrings && c.stringsTab != nil {
		if c.viewer.Selected() == c.stringsTab {
			c.viewer.Select(c.rawTab)
		}
		c.viewer.Remove(c.stringsTab)
		c.stringsTab = nil
	}
}

func (c *stepCard) refreshFullControls(rawCapped, hexCapped, stringsCapped bool) {
	c.fullControls.RemoveAll()
	add := func(label string, enabled bool, setFull func()) {
		if !enabled {
			return
		}
		button := widget.NewButton(label, func() {
			if c.gui.working {
				return
			}
			dialog.ShowConfirm(
				label+"?",
				"Rendering the full view can use a lot of memory and may make the interface slow for large binary data.",
				func(ok bool) {
					if !ok || c.gui.working {
						return
					}
					setFull()
					c.refresh()
				},
				c.gui.window,
			)
		})
		button.Importance = widget.LowImportance
		c.fullControls.Add(button)
	}
	add("Show full Raw", rawCapped && !c.fullRaw, func() { c.fullRaw = true })
	add("Show full Hex", hexCapped && !c.fullHex, func() { c.fullHex = true })
	add("Show full Strings", stringsCapped && !c.fullStrings, func() { c.fullStrings = true })
	if c.fullRaw || c.fullHex || c.fullStrings {
		notice := widget.NewLabel("Full view enabled; output is read-only.")
		notice.Importance = widget.WarningImportance
		c.fullControls.Add(notice)
	}
	if len(c.fullControls.Objects) == 0 {
		c.fullControls.Hide()
	} else {
		c.fullControls.Show()
	}
	c.fullControls.Refresh()
}

func stepToggleControl(check *widget.Check, label string, accent color.NRGBA) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.LowImportance
	body := container.NewVBox(title, check)
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.StrokeColor = accent
	bg.StrokeWidth = 1
	bg.CornerRadius = theme.Size(theme.SizeNameButtonRadius)
	return container.NewStack(bg, container.NewPadded(body))
}

// refresh updates the body and status from the pipeline output.
func (c *stepCard) refresh() {
	out := c.gui.pipe.Output(c.index)
	inputBytes := len(c.gui.pipe.Input(c.index))
	c.meta.SetText(metadataSummary("", pipeline.DataMetadata(out, inputBytes)))
	if err := c.gui.pipe.Err(c.index); err != nil {
		c.status.SetText("error: " + err.Error())
		c.status.Show()
	} else {
		c.status.Hide()
	}
	_, rawNeedsFull := guiTextDisplayMode(out, false)
	_, hexNeedsFull := guiHexDisplayMode(out, false)
	_, stringsNeedsFull := guiStringsDisplayMode(out, false)
	if !rawNeedsFull {
		c.fullRaw = false
	}
	if !hexNeedsFull {
		c.fullHex = false
	}
	if !stringsNeedsFull {
		c.fullStrings = false
	}
	text, textCapped := guiTextDisplayMode(out, c.fullRaw)
	if textCapped {
		c.body.Disable()
	} else {
		c.body.Enable()
	}
	c.gui.setText(c.body, text)
	hexText, _ := guiHexDisplayMode(out, c.fullHex)
	c.gui.setText(c.hexBody, hexText)
	stringsText, _ := guiStringsDisplayMode(out, c.fullStrings)
	c.gui.setText(c.stringsBody, stringsText)
	if c.image != nil {
		setImagePreview(c.image, c.imageMsg, out)
	}
	c.syncStringsTab(out)
	c.syncPreviewTab(out)
	if c.preview != nil {
		preview, spans, _ := pipeline.HighlightedPreview(out)
		setPreviewText(c.preview, preview, spans)
	}
	c.hexBody.Disable()
	c.stringsBody.Disable()
	c.refreshFullControls(rawNeedsFull, hexNeedsFull, stringsNeedsFull)
}

func setImagePreview(img *canvas.Image, msg *widget.Label, data []byte) {
	decoded, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		img.Hide()
		msg.SetText("No image preview available.")
		msg.Show()
		return
	}
	img.Image = decoded
	img.Show()
	img.Refresh()
	msg.SetText("image/" + format)
	msg.Show()
}

var previewStyles = map[pipeline.SyntaxKind]widget.TextGridStyle{
	pipeline.SyntaxKey:         &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x24, G: 0x74, B: 0xd5, A: 0xff}},
	pipeline.SyntaxString:      &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x0f, G: 0x9d, B: 0x58, A: 0xff}},
	pipeline.SyntaxNumber:      &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0xdb, G: 0x44, B: 0x37, A: 0xff}},
	pipeline.SyntaxBool:        &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0xab, G: 0x47, B: 0xbc, A: 0xff}},
	pipeline.SyntaxNull:        &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x8a, G: 0x6d, B: 0x00, A: 0xff}},
	pipeline.SyntaxPunctuation: &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x7a, G: 0x7a, B: 0x7a, A: 0xff}},
}

func setPreviewText(grid *widget.TextGrid, text string, spans []pipeline.SyntaxSpan) {
	grid.SetText(text)
	for _, span := range spans {
		style := previewStyles[span.Kind]
		if style == nil || span.Start < 0 || span.End > len(text) || span.Start >= span.End {
			continue
		}
		startRow, startCol := byteOffsetToGridPosition(text, span.Start)
		endRow, endCol := byteOffsetToGridPosition(text, span.End)
		if endCol > 0 {
			endCol--
		}
		grid.SetStyleRange(startRow, startCol, endRow, endCol, style)
	}
	grid.Refresh()
}

func byteOffsetToGridPosition(text string, offset int) (row, col int) {
	for i, r := range text {
		if i >= offset {
			return row, col
		}
		if r == '\n' {
			row++
			col = 0
			continue
		}
		col++
	}
	return row, col
}
