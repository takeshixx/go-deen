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
)

// multilineEntry returns a word-wrapping multi-line entry sized for content.
func multilineEntry() *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapBreak
	e.SetMinRowsVisible(4)
	return e
}

// newSourceCard builds the editable source-input card at the top of the chain.
func (dg *DeenGUI) newSourceCard() fyne.CanvasObject {
	dg.sourceEntry = multilineEntry()
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

	options   *fyne.Container
	body      *widget.Entry
	status    *widget.Label
	container fyne.CanvasObject
}

func (dg *DeenGUI) newStepCard(i int) *stepCard {
	step := dg.pipe.Steps()[i]
	c := &stepCard{gui: dg, index: i}

	pluginSel := widget.NewSelectEntry(dg.pluginNames)
	pluginSel.SetText(step.Plugin)
	decode := widget.NewCheck("decode", nil)
	decode.SetChecked(step.Unprocess)

	apply := func() {
		name := pluginSel.Text
		if !dg.isPlugin(name) {
			return
		}
		dg.pipe.SetPlugin(i, name, decode.Checked)
		c.rebuildOptions()
		dg.refreshFrom(i)
	}
	pluginSel.OnChanged = func(string) { apply() }
	decode.OnChanged = func(bool) { apply() }

	hexToggle := widget.NewCheck("hex", func(b bool) {
		c.hexView = b
		c.refresh()
	})
	remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dg.pipe.RemoveStep(i)
		dg.rebuild()
	})

	header := container.NewBorder(nil, nil,
		container.NewHBox(pluginSel, decode),
		container.NewHBox(hexToggle, remove),
	)

	c.options = container.NewHBox()
	c.body = multilineEntry()
	c.body.OnChanged = func(s string) {
		if dg.updating || c.hexView {
			return
		}
		dg.pipe.EditOutput(c.index, []byte(s))
		dg.refreshFrom(c.index + 1)
	}
	c.status = widget.NewLabel("")
	c.status.Hide()

	body := container.NewVBox(header, c.options, c.body, c.status)
	c.container = widget.NewCard(fmt.Sprintf("Step %d", i+1), "", body)

	c.rebuildOptions()
	c.refresh()
	return c
}

// rebuildOptions repopulates the per-plugin option widgets for this step.
func (c *stepCard) rebuildOptions() {
	c.options.RemoveAll()
	step := c.gui.pipe.Steps()[c.index]
	for _, opt := range pipeline.PluginOptions(step.Plugin) {
		opt := opt
		if opt.IsBool {
			chk := widget.NewCheck(opt.Name, nil)
			chk.SetChecked(step.Options[opt.Name] == "true")
			chk.OnChanged = func(b bool) {
				val := "false"
				if b {
					val = "true"
				}
				c.gui.pipe.SetOption(c.index, opt.Name, val)
				c.gui.refreshFrom(c.index)
			}
			c.options.Add(chk)
		} else {
			entry := widget.NewEntry()
			entry.SetPlaceHolder(fmt.Sprintf("%s (%s)", opt.Name, opt.Default))
			if v, ok := step.Options[opt.Name]; ok {
				entry.SetText(v)
			}
			entry.OnChanged = func(s string) {
				c.gui.pipe.SetOption(c.index, opt.Name, s)
				c.gui.refreshFrom(c.index)
			}
			c.options.Add(entry)
		}
	}
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

// newAddSlot builds the trailing "add transform" selector.
func (dg *DeenGUI) newAddSlot() fyne.CanvasObject {
	sel := widget.NewSelectEntry(dg.pluginNames)
	sel.SetPlaceHolder("+ add transform")
	sel.OnChanged = func(name string) {
		if !dg.isPlugin(name) {
			return
		}
		dg.pipe.AddStep(name, false)
		dg.rebuild()
	}
	return container.NewBorder(nil, nil, widget.NewLabel("Add:"), nil, sel)
}

// isPlugin reports whether name is a known plugin.
func (dg *DeenGUI) isPlugin(name string) bool {
	for _, n := range dg.pluginNames {
		if n == name {
			return true
		}
	}
	return false
}
