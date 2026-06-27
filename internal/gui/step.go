//go:build gui

package gui

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image/color"

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

// multilineEntry returns a word-wrapping multi-line entry with a readable
// minimum height.
func multilineEntry(rows int) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapBreak
	e.SetMinRowsVisible(rows)
	return e
}

// categorySelectors builds one dropdown per plugin category (codecs, hashs, …),
// each listing the plugins in that category. Selecting a plugin in any dropdown
// calls onPick and clears the others. When current is non-empty its dropdown is
// pre-selected (highlighting which plugin/category is in use).
func (dg *DeenGUI) categorySelectors(current string, onPick func(name string)) *fyne.Container {
	cats := plugins.PluginCategories
	selects := make([]*widget.Select, len(cats))
	for i, cat := range cats {
		s := widget.NewSelect(plugins.InCategory(cat), nil)
		s.PlaceHolder = cat
		selects[i] = s
	}

	clearOthers := func(except *widget.Select) {
		for _, s := range selects {
			if s != except && s.Selected != "" {
				s.Selected = ""
				s.Refresh()
			}
		}
	}
	for _, s := range selects {
		s := s
		s.OnChanged = func(name string) {
			if name == "" {
				return
			}
			clearOthers(s)
			onPick(name)
		}
	}

	if current != "" {
		cat := plugins.CategoryOf(current)
		for i, c := range cats {
			if c == cat {
				selects[i].Selected = current // set directly so OnChanged does not fire
				selects[i].Refresh()
			}
		}
	}

	objs := make([]fyne.CanvasObject, len(selects))
	for i, s := range selects {
		objs[i] = s
	}
	return container.NewGridWithColumns(len(selects), objs...)
}

// newSourceCard builds the editable source-input card at the top of the chain.
func (dg *DeenGUI) newSourceCard() fyne.CanvasObject {
	dg.sourceEntry = multilineEntry(6)
	dg.sourceEntry.SetText(string(dg.pipe.Source()))
	dg.sourceMeta = widget.NewLabel(pipeline.DataMetadata(dg.pipe.Source(), 0).Summary())
	dg.sourceMeta.Importance = widget.LowImportance
	dg.sourceEntry.OnChanged = func(s string) {
		if dg.updating {
			return
		}
		dg.pipe.SetSource([]byte(s))
		dg.refreshFrom(0)
	}
	return widget.NewCard("Input", "", container.NewVBox(dg.sourceEntry, dg.sourceMeta))
}

// stepCard is the view for a single pipeline step.
type stepCard struct {
	gui        *DeenGUI
	index      int
	pluginName string
	viewMode   string
	collapsed  bool

	decode    *widget.Check
	enabled   *widget.Check
	summary   *canvas.Text
	collapse  *widget.Button
	detail    *fyne.Container
	options   *fyne.Container
	body      *widget.Entry
	meta      *widget.Label
	status    *widget.Label
	container fyne.CanvasObject
}

func (dg *DeenGUI) newStepCard(i int) *stepCard {
	step := dg.pipe.Steps()[i]
	c := &stepCard{gui: dg, index: i, pluginName: step.Plugin, viewMode: "text"}
	col := accent(i)

	c.decode = widget.NewCheck("decode", nil)
	c.decode.SetChecked(step.Unprocess)
	c.enabled = widget.NewCheck("enabled", nil)
	c.enabled.SetChecked(!step.Disabled)

	apply := func() {
		if c.pluginName == "" {
			return
		}
		dg.pipe.SetPlugin(c.index, c.pluginName, c.decode.Checked)
		c.rebuildOptions()
		c.updateSummary()
		dg.refreshFrom(c.index)
		dg.updateHistory()
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

	viewMode := widget.NewSelect([]string{"text", "hex", "base64"}, func(mode string) {
		if mode == "" {
			return
		}
		c.viewMode = mode
		c.refresh()
	})
	viewMode.Selected = c.viewMode

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
		if dg.updating || c.viewMode != "text" {
			return
		}
		dg.pipe.EditOutput(c.index, []byte(s))
		dg.refreshFrom(c.index + 1)
	}
	c.meta = widget.NewLabel("")
	c.meta.Importance = widget.LowImportance
	c.status = widget.NewLabel("")
	c.status.Importance = widget.DangerImportance
	c.status.Hide()
	c.detail = container.NewVBox(selectors, container.NewHBox(c.enabled, c.decode, widget.NewLabel("view"), viewMode), c.options, c.body, c.meta, c.status)

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
	c.summary.Text = fmt.Sprintf("  %s / %s · %s", cat, name, dir)
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
			form.Append(opt.Name, chk)
		} else {
			entry := widget.NewEntry()
			entry.SetPlaceHolder(fmt.Sprintf("default: %s", opt.Default))
			if v, ok := step.Options[opt.Name]; ok {
				entry.SetText(v)
			}
			entry.OnChanged = func(s string) {
				c.gui.pipe.SetOption(c.index, opt.Name, s)
				c.gui.refreshFrom(c.index)
			}
			form.Append(opt.Name, entry)
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
	c.meta.SetText(pipeline.DataMetadata(out, inputBytes).Summary())
	if err := c.gui.pipe.Err(c.index); err != nil {
		c.status.SetText("error: " + err.Error())
		c.status.Show()
	} else {
		c.status.Hide()
	}
	switch c.viewMode {
	case "hex":
		c.gui.setText(c.body, hex.Dump(out))
		c.body.Disable()
	case "base64":
		c.gui.setText(c.body, base64.StdEncoding.EncodeToString(out))
		c.body.Disable()
	default:
		c.viewMode = "text"
		c.body.Enable()
		c.gui.setText(c.body, string(out))
	}
	if c.viewMode == "text" {
		c.body.Enable()
	} else {
		c.body.Disable()
	}
}

// newAddSlot builds the trailing per-category dropdowns that append a step.
func (dg *DeenGUI) newAddSlot() fyne.CanvasObject {
	selectors := dg.categorySelectors("", func(name string) {
		dg.pipe.AddStep(name, false)
		dg.rebuild()
	})
	return widget.NewCard("Add transform", "", selectors)
}
