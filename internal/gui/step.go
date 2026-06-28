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

// stepPalette gives each step a distinct accent colour (cycled).
var stepPalette = []color.NRGBA{
	{0x42, 0x85, 0xf4, 0xff}, // blue
	{0x0f, 0x9d, 0x58, 0xff}, // green
	{0xf4, 0xb4, 0x00, 0xff}, // amber
	{0xdb, 0x44, 0x37, 0xff}, // red
	{0xab, 0x47, 0xbc, 0xff}, // purple
	{0x00, 0xac, 0xc1, 0xff}, // cyan
}

func accent(i int) color.NRGBA { return stepPalette[i%len(stepPalette)] }

// tint returns the accent at low opacity, for card backgrounds.
func tint(c color.NRGBA) color.NRGBA { return color.NRGBA{R: c.R, G: c.G, B: c.B, A: 0x22} }

func disabledAccent() color.NRGBA { return color.NRGBA{R: 0x8c, G: 0x96, B: 0x9b, A: 0xff} }

const outputViewerHeight float32 = 260
const compactControlMinWidth float32 = 360
const sourceInputRows = 14
const sourceInputMinHeight float32 = 160
const sourceInputMaxHeight float32 = 306

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

// multilineEntry returns a word-wrapping multi-line entry with a readable
// minimum height.
func multilineEntry(rows int) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapBreak
	e.SetMinRowsVisible(rows)
	return e
}

func sourceInputHeight(data []byte) float32 {
	if len(data) == 0 {
		return sourceInputMinHeight
	}
	if pipeline.IsLargeData(data) || pipeline.IsBinaryData(data) || len(data) > 4096 {
		return sourceInputMaxHeight
	}
	lines := bytes.Count(data, []byte{'\n'}) + 1
	switch {
	case lines <= 3:
		return sourceInputMinHeight
	case lines <= 8:
		return 240
	default:
		return sourceInputMaxHeight
	}
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

	return container.New(cappedMinWidthLayout{width: compactControlMinWidth}, container.NewGridWithColumns(2, categorySelect, transformerSelect))
}

func (dg *DeenGUI) addCategorySelectors() fyne.CanvasObject {
	pickers := make([]fyne.CanvasObject, 0, len(plugins.PluginCategories))
	for _, category := range plugins.PluginCategories {
		category := category
		labels, labelToName, _ := pluginSelectLabels(category)
		selectBox := widget.NewSelect(labels, func(label string) {
			name := labelToName[label]
			if name == "" {
				return
			}
			dg.runPipelineWork("Processing", func() error {
				dg.pipe.AddStep(name, false)
				return nil
			}, dg.rebuild)
		})
		selectBox.PlaceHolder = plugins.CategorySelectLabel(category)

		label := widget.NewLabelWithStyle(plugins.CategoryLabel(category), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		pickers = append(pickers, container.NewVBox(label, selectBox))
	}
	return container.New(cappedMinWidthLayout{width: compactControlMinWidth}, container.NewGridWithColumns(2, pickers...))
}

// newSourceCard builds the editable source-input card at the top of the chain.
func (dg *DeenGUI) newSourceCard() fyne.CanvasObject {
	dg.sourceEntry = multilineEntry(sourceInputRows)
	dg.sourceHex = nil
	dg.sourceStrings = nil
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
	dg.sourceMeta = widget.NewLabel(dg.sourceMetadataSummary())
	dg.sourceMeta.Importance = widget.LowImportance
	dg.sourceMeta.Wrapping = fyne.TextWrapBreak
	dg.sourceMeta.TextStyle.Monospace = true
	dg.sourceEntry.OnChanged = func(s string) {
		if dg.updating {
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
	var sourceView fyne.CanvasObject = dg.sourceEntry
	if pipeline.IsBinaryData(dg.pipe.Source()) {
		dg.sourceHex = multilineEntry(sourceInputRows)
		hexText, _ := guiHexDisplayMode(dg.pipe.Source(), dg.sourceFullHex)
		dg.sourceHex.SetText(hexText)
		dg.sourceHex.Disable()
		dg.sourceStrings = multilineEntry(sourceInputRows)
		stringsText, _ := guiStringsDisplayMode(dg.pipe.Source(), dg.sourceFullStrings)
		dg.sourceStrings.SetText(stringsText)
		dg.sourceStrings.Disable()
		viewer := container.NewAppTabs(
			container.NewTabItem("Raw", dg.sourceEntry),
			container.NewTabItem("Hex", dg.sourceHex),
			container.NewTabItem("Strings", dg.sourceStrings),
		)
		viewer.SetTabLocation(container.TabLocationTop)
		viewer.SelectIndex(1)
		sourceView = viewer
	}
	sourceBox := container.New(fixedHeightLayout{height: sourceInputHeight(dg.pipe.Source())}, sourceView)
	dg.sourceFullControls = container.NewHBox()
	dg.refreshSourceFullControls(rawNeedsFull, hexNeedsFull, stringsNeedsFull)
	content := container.New(cappedMinWidthLayout{width: compactControlMinWidth}, container.NewVBox(sourceBox, dg.sourceFullControls, dg.sourceMeta))
	return widget.NewCard("Input", "", container.NewPadded(content))
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
			dialog.ShowConfirm(
				label+"?",
				"Rendering the full input view can use a lot of memory and may make the interface slow for large files.",
				func(ok bool) {
					if !ok {
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

	decode       *widget.Check
	enabled      *widget.Check
	summary      *canvas.Text
	collapse     *widget.Button
	detail       *fyne.Container
	options      *fyne.Container
	fullControls *fyne.Container
	body         *widget.Entry
	hexBody      *widget.Entry
	stringsBody  *widget.Entry
	viewer       *container.AppTabs
	rawTab       *container.TabItem
	hexTab       *container.TabItem
	stringsTab   *container.TabItem
	previewTab   *container.TabItem
	preview      *widget.TextGrid
	image        *canvas.Image
	imageMsg     *widget.Label
	meta         *widget.Label
	status       *widget.Label
	container    fyne.CanvasObject
	fullRaw      bool
	fullHex      bool
	fullStrings  bool
}

func (dg *DeenGUI) newStepCard(i int) *stepCard {
	step := dg.pipe.Steps()[i]
	c := &stepCard{gui: dg, index: i, pluginName: step.Plugin}
	col := accent(i)
	canDecode := plugins.CanDecode(step.Plugin)

	c.decode = widget.NewCheck("decode", nil)
	c.decode.SetChecked(step.Unprocess && canDecode)
	c.enabled = widget.NewCheck("enabled", nil)
	c.enabled.SetChecked(!step.Disabled)

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
		dg.runPipelineWork("Processing", func() error {
			if c.index < 0 || c.index >= dg.pipe.Len() {
				return nil
			}
			dg.pipe.SetStepDisabled(c.index, !dg.pipe.Steps()[c.index].Disabled)
			return nil
		}, dg.rebuild)
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
	c.collapse = stepIconButton(theme.MenuDropDownIcon(), c.toggleCollapse)
	moveUp := stepIconButton(theme.MoveUpIcon(), func() {
		dg.runPipelineWork("Processing", func() error {
			dg.pipe.MoveStep(c.index, c.index-1)
			return nil
		}, dg.rebuild)
	})
	moveDown := stepIconButton(theme.MoveDownIcon(), func() {
		dg.runPipelineWork("Processing", func() error {
			dg.pipe.MoveStep(c.index, c.index+1)
			return nil
		}, dg.rebuild)
	})
	duplicate := stepIconButton(theme.ContentCopyIcon(), func() {
		dg.runPipelineWork("Processing", func() error {
			dg.pipe.DuplicateStep(c.index)
			return nil
		}, dg.rebuild)
	})
	remove := stepIconButton(theme.DeleteIcon(), func() {
		dg.runPipelineWork("Processing", func() error {
			dg.pipe.RemoveStep(c.index)
			return nil
		}, dg.rebuild)
	})
	enabledIcon := theme.VisibilityIcon()
	if step.Disabled {
		enabledIcon = theme.VisibilityOffIcon()
	}
	enabledControl := stepIconButton(enabledIcon, toggleEnabled)
	titleRow := container.NewBorder(nil, nil,
		container.NewHBox(c.collapse, title, c.summary),
		container.NewHBox(enabledControl, moveUp, moveDown, duplicate, remove))

	// Detail: selectors, toggles, options, output, errors.
	c.options = container.NewVBox()
	c.body = multilineEntry(6)
	c.body.OnChanged = func(s string) {
		if dg.updating {
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
	if c.previewTab != nil {
		viewer.Select(c.previewTab)
	} else if pipeline.IsBinaryData(dg.pipe.Output(i)) {
		viewer.Select(c.hexTab)
	} else if stepGeneratesImage(step) {
		viewer.Select(viewerTabs[len(viewerTabs)-1])
	}
	viewerBox := container.New(fixedHeightLayout{height: outputViewerHeight}, viewer)
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
	c.detail = container.NewVBox(selectors, toggles, c.options, viewerBox, c.fullControls, c.meta, c.status)

	bg := canvas.NewRectangle(tint(displayCol))
	bg.StrokeColor = displayCol
	bg.StrokeWidth = 2
	bg.CornerRadius = 6
	inner := container.NewVBox(titleRow, c.detail)
	c.container = container.NewStack(bg, container.NewPadded(inner))

	c.rebuildOptions()
	c.updateSummary()
	if !dg.stepsExpanded && i < len(dg.pipe.Steps())-1 {
		c.collapsed = true
		c.detail.Hide()
		c.collapse.SetIcon(theme.NavigateNextIcon())
	}
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
		c.collapse.SetIcon(theme.NavigateNextIcon())
	} else {
		c.refresh()
		c.detail.Show()
		c.collapse.SetIcon(theme.MenuDropDownIcon())
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
	if !c.enabled.Checked {
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
			selectInput := widget.NewSelect(opt.Choices, func(s string) {
				c.gui.runPipelineWork("Processing", func() error {
					c.gui.pipe.SetOption(c.index, opt.Name, s)
					return nil
				}, func() { c.gui.refreshFrom(c.index) })
			})
			if v, ok := step.Options[opt.Name]; ok {
				selectInput.SetSelected(v)
			} else {
				selectInput.SetSelected(opt.Default)
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
			entry.SetPlaceHolder(optionPlaceholder(opt))
			if v, ok := step.Options[opt.Name]; ok {
				entry.SetText(v)
			}
			entry.OnChanged = func(s string) {
				c.gui.pipe.SetOption(c.index, opt.Name, s)
				c.gui.refreshFrom(c.index)
			}
			control = entry
			target = &fieldOptions
		}
		*target = append(*target, optionBlock(label, control, opt))
	}
	if len(checkOptions) > 0 {
		c.options.Add(optionSection("Checkboxes", checkOptions))
	}
	if len(fieldOptions) > 0 {
		c.options.Add(optionSection("Inputs", fieldOptions))
	}
	c.options.Refresh()
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

func stepIconButton(icon fyne.Resource, tapped func()) *widget.Button {
	button := widget.NewButtonWithIcon("", icon, tapped)
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
			dialog.ShowConfirm(
				label+"?",
				"Rendering the full view can use a lot of memory and may make the interface slow for large binary data.",
				func(ok bool) {
					if !ok {
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
	bg := canvas.NewRectangle(tint(accent))
	bg.StrokeColor = accent
	bg.StrokeWidth = 1
	bg.CornerRadius = 6
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

// newAddSlot builds the compact category/transformer picker that appends a step.
func (dg *DeenGUI) newAddSlot() fyne.CanvasObject {
	actions := container.NewHBox(
		widget.NewButtonWithIcon("Search transformers", theme.SearchIcon(), dg.showPluginSearch),
		widget.NewButtonWithIcon("Detect next", theme.ContentAddIcon(), dg.showSuggestions),
	)
	subtitle := widget.NewLabel("Choose a transformer by category or search the catalog.")
	subtitle.Importance = widget.LowImportance
	return widget.NewCard("Add transformer step", "", container.NewVBox(subtitle, actions, dg.addCategorySelectors()))
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
