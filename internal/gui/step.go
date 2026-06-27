//go:build gui

package gui

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

const outputViewerHeight float32 = 260
const compactControlMinWidth float32 = 360

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
			dg.pipe.AddStep(name, false)
			dg.rebuild()
		})
		selectBox.PlaceHolder = plugins.CategorySelectLabel(category)

		label := widget.NewLabelWithStyle(plugins.CategoryLabel(category), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		pickers = append(pickers, container.NewVBox(label, selectBox))
	}
	return container.New(cappedMinWidthLayout{width: compactControlMinWidth}, container.NewGridWithColumns(2, pickers...))
}

// newSourceCard builds the editable source-input card at the top of the chain.
func (dg *DeenGUI) newSourceCard() fyne.CanvasObject {
	dg.sourceEntry = multilineEntry(6)
	dg.sourceEntry.SetText(string(dg.pipe.Source()))
	dg.sourceMeta = widget.NewLabel(dg.sourceMetadataSummary())
	dg.sourceMeta.Importance = widget.LowImportance
	dg.sourceMeta.Wrapping = fyne.TextWrapBreak
	dg.sourceEntry.OnChanged = func(s string) {
		if dg.updating {
			return
		}
		dg.sourceName = ""
		dg.pipe.SetSource([]byte(s))
		dg.refreshFrom(0)
	}
	content := container.New(cappedMinWidthLayout{width: compactControlMinWidth}, container.NewVBox(dg.sourceEntry, dg.sourceMeta))
	return widget.NewCard("Input", "", container.NewPadded(content))
}

// stepCard is the view for a single pipeline step.
type stepCard struct {
	gui        *DeenGUI
	index      int
	pluginName string
	collapsed  bool

	decode    *widget.Check
	enabled   *widget.Check
	summary   *canvas.Text
	collapse  *widget.Button
	detail    *fyne.Container
	options   *fyne.Container
	body      *widget.Entry
	hexBody   *widget.Entry
	b64Body   *widget.Entry
	statsBody *widget.Entry
	preview   *widget.TextGrid
	image     *canvas.Image
	imageMsg  *widget.Label
	meta      *widget.Label
	status    *widget.Label
	container fyne.CanvasObject
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
		dg.pipe.SetPlugin(c.index, c.pluginName, decode)
		c.rebuildOptions()
		c.updateSummary()
		dg.rebuild()
	}
	selectors := dg.categorySelectors(step.Plugin, func(name string) {
		c.pluginName = name
		apply()
	})
	c.decode.OnChanged = func(bool) { apply() }
	c.enabled.OnChanged = func(enabled bool) {
		dg.pipe.SetStepDisabled(c.index, !enabled)
		c.updateSummary()
		dg.refreshFrom(c.index)
		dg.updateHistory()
	}

	// Title row: collapse toggle, coloured title, active-plugin summary, remove.
	c.summary = canvas.NewText("", col)
	c.summary.TextStyle = fyne.TextStyle{Bold: true}
	title := canvas.NewText(fmt.Sprintf("Step %d", i+1), col)
	title.TextStyle = fyne.TextStyle{Bold: true}
	c.collapse = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), c.toggleCollapse)
	c.collapse.Importance = widget.LowImportance
	moveUp := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		dg.pipe.MoveStep(c.index, c.index-1)
		dg.rebuild()
	})
	moveDown := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		dg.pipe.MoveStep(c.index, c.index+1)
		dg.rebuild()
	})
	duplicate := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		dg.pipe.DuplicateStep(c.index)
		dg.rebuild()
	})
	remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dg.pipe.RemoveStep(c.index)
		dg.rebuild()
	})
	titleRow := container.NewBorder(nil, nil,
		container.NewHBox(c.collapse, title, c.summary),
		container.NewHBox(moveUp, moveDown, duplicate, remove))

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
	c.b64Body = multilineEntry(6)
	c.b64Body.Disable()
	c.statsBody = multilineEntry(6)
	c.statsBody.Disable()
	c.preview = widget.NewTextGrid()
	c.preview.ShowLineNumbers = false
	c.preview.Scroll = fyne.ScrollBoth
	c.image = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	c.image.FillMode = canvas.ImageFillContain
	c.imageMsg = widget.NewLabel("No image preview available.")
	c.imageMsg.Alignment = fyne.TextAlignCenter
	c.image.Hide()
	imageViewer := container.NewBorder(nil, c.imageMsg, nil, nil, container.NewPadded(c.image))
	viewer := container.NewAppTabs(
		container.NewTabItem("Text", c.body),
		container.NewTabItem("Hex", c.hexBody),
		container.NewTabItem("Base64", c.b64Body),
		container.NewTabItem("Stats", c.statsBody),
		container.NewTabItem("Preview", c.preview),
		container.NewTabItem("Image", imageViewer),
	)
	viewer.SetTabLocation(container.TabLocationTop)
	viewerBox := container.New(fixedHeightLayout{height: outputViewerHeight}, viewer)
	c.meta = widget.NewLabel("")
	c.meta.Importance = widget.LowImportance
	c.meta.Wrapping = fyne.TextWrapBreak
	c.status = widget.NewLabel("")
	c.status.Importance = widget.DangerImportance
	c.status.Wrapping = fyne.TextWrapBreak
	c.status.Hide()
	toggles := container.NewHBox(c.enabled)
	if canDecode {
		toggles.Add(c.decode)
	}
	c.detail = container.NewVBox(selectors, toggles, c.options, viewerBox, c.meta, c.status)

	bg := canvas.NewRectangle(tint(col))
	bg.StrokeColor = col
	bg.StrokeWidth = 2
	bg.CornerRadius = 6
	inner := container.NewVBox(titleRow, c.detail)
	c.container = container.NewStack(bg, container.NewPadded(inner))

	c.rebuildOptions()
	c.updateSummary()
	c.refresh()
	return c
}

// toggleCollapse hides or shows the step's detail section.
func (c *stepCard) toggleCollapse() {
	c.collapsed = !c.collapsed
	if c.collapsed {
		c.detail.Hide()
		c.collapse.SetIcon(theme.NavigateNextIcon())
	} else {
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
	form := widget.NewForm()
	for _, opt := range opts {
		opt := opt
		if opt.IsBool {
			chk := widget.NewCheck("", nil)
			chk.SetChecked(step.Options[opt.Name] == "true")
			chk.OnChanged = func(b bool) {
				val := "false"
				if b {
					val = "true"
				}
				c.gui.pipe.SetOption(c.index, opt.Name, val)
				c.gui.refreshFrom(c.index)
			}
			form.Append(opt.Label, chk)
		} else if opt.Kind == "select" {
			selectInput := widget.NewSelect(opt.Choices, func(s string) {
				c.gui.pipe.SetOption(c.index, opt.Name, s)
				c.gui.refreshFrom(c.index)
			})
			if v, ok := step.Options[opt.Name]; ok {
				selectInput.SetSelected(v)
			} else {
				selectInput.SetSelected(opt.Default)
			}
			form.Append(opt.Label, selectInput)
		} else {
			entry := widget.NewEntry()
			if opt.Kind == "secret" || opt.Secret {
				entry = widget.NewPasswordEntry()
			}
			entry.SetPlaceHolder(fmt.Sprintf("default: %s", opt.Default))
			if v, ok := step.Options[opt.Name]; ok {
				entry.SetText(v)
			}
			entry.OnChanged = func(s string) {
				c.gui.pipe.SetOption(c.index, opt.Name, s)
				c.gui.refreshFrom(c.index)
			}
			form.Append(opt.Label, entry)
		}
	}
	c.options.Add(form)
	c.options.Refresh()
}

// refresh updates the body and status from the pipeline output.
func (c *stepCard) refresh() {
	out := c.gui.pipe.Output(c.index)
	inputBytes := len(c.gui.pipe.Source())
	if c.index > 0 {
		inputBytes = len(c.gui.pipe.Output(c.index - 1))
	}
	summary := pipeline.DataMetadata(out, inputBytes).Summary()
	c.meta.SetText(summary)
	if err := c.gui.pipe.Err(c.index); err != nil {
		c.status.SetText("error: " + err.Error())
		c.status.Show()
	} else {
		c.status.Hide()
	}
	c.body.Enable()
	c.gui.setText(c.body, string(out))
	c.gui.setText(c.hexBody, hex.Dump(out))
	c.gui.setText(c.b64Body, base64.StdEncoding.EncodeToString(out))
	c.gui.setText(c.statsBody, summary)
	setImagePreview(c.image, c.imageMsg, out)
	preview, spans, ok := pipeline.HighlightedPreview(out)
	if !ok {
		preview = "No structured preview available."
		spans = nil
	}
	setPreviewText(c.preview, preview, spans)
	c.hexBody.Disable()
	c.b64Body.Disable()
	c.statsBody.Disable()
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
