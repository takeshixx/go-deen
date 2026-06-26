//go:build gui

package gui

import (
	"encoding/hex"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

// multilineEntry returns a word-wrapping multi-line entry with a readable
// minimum height.
func multilineEntry(rows int) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapBreak
	e.SetMinRowsVisible(rows)
	return e
}

// pluginPicker builds a category dropdown plus a plugin dropdown for that
// category. Selecting a plugin calls onPick. When current is non-empty its
// category and name are pre-selected without firing onPick.
func (dg *DeenGUI) pluginPicker(current string, onPick func(name string)) (*widget.Select, *widget.Select) {
	catSel := widget.NewSelect(plugins.PluginCategories, nil)
	catSel.PlaceHolder = "category"
	plugSel := widget.NewSelect(nil, nil)
	plugSel.PlaceHolder = "plugin"

	if current != "" {
		cat := plugins.CategoryOf(current)
		catSel.Selected = cat
		plugSel.Options = plugins.InCategory(cat)
		plugSel.Selected = current
	}

	catSel.OnChanged = func(cat string) {
		plugSel.Options = plugins.InCategory(cat)
		plugSel.ClearSelected()
		plugSel.Refresh()
	}
	plugSel.OnChanged = func(name string) {
		if name != "" {
			onPick(name)
		}
	}
	return catSel, plugSel
}

// newSourceCard builds the editable source-input card at the top of the chain.
func (dg *DeenGUI) newSourceCard() fyne.CanvasObject {
	dg.sourceEntry = multilineEntry(6)
	dg.sourceEntry.SetText(string(dg.pipe.Source()))
	dg.sourceEntry.OnChanged = func(s string) {
		if dg.updating {
			return
		}
		dg.pipe.SetSource([]byte(s))
		dg.refreshFrom(0)
	}
	return widget.NewCard("Input", "", dg.sourceEntry)
}

// stepCard is the view for a single pipeline step.
type stepCard struct {
	gui     *DeenGUI
	index   int
	hexView bool

	plugSel   *widget.Select
	decode    *widget.Check
	options   *fyne.Container
	body      *widget.Entry
	status    *widget.Label
	container fyne.CanvasObject
}

func (dg *DeenGUI) newStepCard(i int) *stepCard {
	step := dg.pipe.Steps()[i]
	c := &stepCard{gui: dg, index: i}

	c.decode = widget.NewCheck("decode", nil)
	c.decode.SetChecked(step.Unprocess)

	apply := func() {
		name := c.plugSel.Selected
		if name == "" {
			return
		}
		dg.pipe.SetPlugin(c.index, name, c.decode.Checked)
		c.rebuildOptions()
		dg.refreshFrom(c.index)
	}
	catSel, plugSel := dg.pluginPicker(step.Plugin, func(string) { apply() })
	c.plugSel = plugSel
	c.decode.OnChanged = func(bool) { apply() }

	hexToggle := widget.NewCheck("hex", func(b bool) {
		c.hexView = b
		c.refresh()
	})
	remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dg.pipe.RemoveStep(c.index)
		dg.rebuild()
	})

	header := container.NewBorder(nil, nil,
		container.NewHBox(catSel, c.plugSel, c.decode),
		container.NewHBox(hexToggle, remove),
	)

	c.options = container.NewVBox()
	c.body = multilineEntry(6)
	c.body.OnChanged = func(s string) {
		if dg.updating || c.hexView {
			return
		}
		dg.pipe.EditOutput(c.index, []byte(s))
		dg.refreshFrom(c.index + 1)
	}
	c.status = widget.NewLabel("")
	c.status.Importance = widget.DangerImportance
	c.status.Hide()

	cardBody := container.NewVBox(header, c.options, c.body, c.status)
	c.container = widget.NewCard(fmt.Sprintf("Step %d", i+1), "", cardBody)

	c.rebuildOptions()
	c.refresh()
	return c
}

// rebuildOptions repopulates the per-plugin option widgets (as a form) for this
// step.
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
	if err := c.gui.pipe.Err(c.index); err != nil {
		c.status.SetText("error: " + err.Error())
		c.status.Show()
	} else {
		c.status.Hide()
	}
	if c.hexView {
		c.gui.setText(c.body, hex.Dump(out))
		c.body.Disable() // hex view is read-only
	} else {
		c.body.Enable()
		c.gui.setText(c.body, string(out))
	}
}

// newAddSlot builds the trailing category/plugin pickers that append a step.
func (dg *DeenGUI) newAddSlot() fyne.CanvasObject {
	catSel, plugSel := dg.pluginPicker("", func(name string) {
		dg.pipe.AddStep(name, false)
		dg.rebuild()
	})
	return widget.NewCard("Add transform", "",
		container.NewHBox(catSel, plugSel))
}
