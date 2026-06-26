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

// categorySelectors builds one dropdown per plugin category (codecs, hashs, …),
// each listing the plugins in that category. Selecting a plugin in any dropdown
// calls onPick and clears the others. When current is non-empty its dropdown is
// pre-selected.
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
	gui        *DeenGUI
	index      int
	pluginName string
	hexView    bool

	decode    *widget.Check
	options   *fyne.Container
	body      *widget.Entry
	status    *widget.Label
	container fyne.CanvasObject
}

func (dg *DeenGUI) newStepCard(i int) *stepCard {
	step := dg.pipe.Steps()[i]
	c := &stepCard{gui: dg, index: i, pluginName: step.Plugin}

	c.decode = widget.NewCheck("decode", nil)
	c.decode.SetChecked(step.Unprocess)

	apply := func() {
		if c.pluginName == "" {
			return
		}
		dg.pipe.SetPlugin(c.index, c.pluginName, c.decode.Checked)
		c.rebuildOptions()
		dg.refreshFrom(c.index)
	}
	selectors := dg.categorySelectors(step.Plugin, func(name string) {
		c.pluginName = name
		apply()
	})
	c.decode.OnChanged = func(bool) { apply() }

	hexToggle := widget.NewCheck("hex", func(b bool) {
		c.hexView = b
		c.refresh()
	})
	remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dg.pipe.RemoveStep(c.index)
		dg.rebuild()
	})
	controls := container.NewBorder(nil, nil, container.NewHBox(c.decode, hexToggle), remove)

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

	cardBody := container.NewVBox(selectors, controls, c.options, c.body, c.status)
	c.container = widget.NewCard(fmt.Sprintf("Step %d", i+1), "", cardBody)

	c.rebuildOptions()
	c.refresh()
	return c
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

// newAddSlot builds the trailing per-category dropdowns that append a step.
func (dg *DeenGUI) newAddSlot() fyne.CanvasObject {
	selectors := dg.categorySelectors("", func(name string) {
		dg.pipe.AddStep(name, false)
		dg.rebuild()
	})
	return widget.NewCard("Add transform", "", selectors)
}
