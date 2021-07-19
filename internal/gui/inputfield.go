// +build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// DeenInputField is a subclass of widget.Entry with additional fields.
type DeenInputField struct {
	widget.Entry
	Parent *DeenEncoder
}

// KeyDown is a wrapper for the input field to support
// shortcuts such as opening the plugin search.
func (f *DeenInputField) KeyDown(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyF2:
		f.Parent.Parent.showPluginSearch()
		f.Entry.KeyDown(key)
	default:
		f.Entry.KeyDown(key)
	}
}

// NewDeenInputField initializes a new DeenInputField
func NewDeenInputField(parent *DeenEncoder) *DeenInputField {
	e := &DeenInputField{
		widget.Entry{MultiLine: true},
		parent,
	}
	e.Wrapping = fyne.TextWrapBreak
	e.ExtendBaseWidget(e)
	return e
}
