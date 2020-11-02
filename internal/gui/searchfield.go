// +build gui

package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
	"github.com/schollz/closestmatch"
	"github.com/takeshixx/deen/internal/plugins"
)

// DeenSearchField is a subclass of widget.Entry with additional fields.
type DeenSearchField struct {
	widget.Entry
	Parent          *DeenGUI
	Layout          *widget.Box
	ConfirmCallBack func(bool)
	CM              *closestmatch.ClosestMatch
	Dialog          dialog.Dialog
	Closest         string
}

// KeyDown is a wrapper for the input field to support
// shortcuts such as opening the plugin search.
func (f *DeenSearchField) KeyDown(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyReturn:
		if strings.HasSuffix(f.Text, ".") {
			f.Text = "." + strings.TrimSuffix(f.Text, ".")
		}
		p := plugins.GetForCmd(f.Text)
		if p != nil {
			f.Parent.RunPlugin(f.Text)
			f.Dialog.Hide()
			f.Parent.MainWindow.Canvas().Focus(f.Parent.CurrentEncoder().InputField)
		}
		f.Entry.KeyDown(key)
	case fyne.KeyEscape:
		f.Dialog.Hide()
		f.Parent.MainWindow.Canvas().Focus(f.Parent.CurrentEncoder().InputField)
	case fyne.KeyTab:
		if f.Closest != "" {
			f.Text = f.Closest
		}
	default:
		f.Entry.KeyDown(key)
	}
}

// Show shows the search field.
func (f *DeenSearchField) Show() {
	f.Dialog.Show()
}

// NewDeenSearchField initializes a new DeenInputField
func NewDeenSearchField(parent *DeenGUI) (e *DeenSearchField) {
	e = &DeenSearchField{
		widget.Entry{},
		parent,
		widget.NewVBox(),
		nil,
		nil,
		nil,
		"",
	}
	e.SetPlaceHolder("Type plugin name")
	e.Layout.Append(e)
	// Create a list of plugins without dot prefix
	var p []string
	for _, x := range e.Parent.Plugins {
		if strings.HasPrefix(x, ".") {
			continue
		}
		p = append(p, x)
	}
	e.CM = closestmatch.New(p, []int{2})
	e.OnChanged = func(text string) {
		fmt.Printf("onchanged text: %s\n", text)
		e.Closest = e.CM.Closest(text)
		fmt.Printf("closest: %s\n", e.Closest)
		if strings.HasPrefix(text, e.Closest) {
			// Do nothing
		} else if e.Closest != "" {
			e.SetText(e.Closest)
		}
	}
	e.ExtendBaseWidget(e)
	return
}
