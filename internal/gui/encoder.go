package gui

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/takeshixx/deen/pkg/types"
)

// DeenEncoder represents an encoder that can be added to the GUI's Encoders list.
type DeenEncoder struct {
	Parent      *DeenGUI
	Content     []byte // The actual content of the widget. Should never be changed, only by following encoder widgets.
	ContentLen  *widget.Label
	View        string // The current view (plain/hex)
	InputField  *DeenInputField
	InputLen    *widget.Label
	ViewButton  *widget.Select // Change the view of the encoder (plain/hex)
	CopyButton  *widget.Button // Copy the content of the encoder to the clipboard
	ClearButton *widget.Button // Clear the contents of the encoder/Remove the encoder widget
	Plugin      *types.DeenPlugin
}

// DeenInputField is a subclass of widget.Entry with additional fields.
type DeenInputField struct {
	widget.Entry
	Parent *DeenEncoder
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

func (de *DeenEncoder) createLayout() (layout *widget.Box) {
	layout = widget.NewVBox()
	encoderWrapper := widget.NewScrollContainer(de.InputField)
	encoderWrapper.SetMinSize(fyne.NewSize(0, 200))
	layout.Append(encoderWrapper)
	buttonsLayout := widget.NewHBox()
	buttonsLayout.Append(de.ViewButton)
	buttonsLayout.Append(de.CopyButton)
	buttonsLayout.Append(de.ClearButton)
	buttonsLayout.Append(de.ContentLen)
	buttonsLayout.Append(de.InputLen)
	layout.Append(buttonsLayout)
	return
}

// SetContent overwrites the Content and InputField data.
func (de *DeenEncoder) SetContent(data []byte) {
	de.Content = data
	de.ContentLen.SetText(fmt.Sprintf("CL: %d", len(de.Content)))
	de.InputField.SetText(string(data))
	de.InputLen.SetText(fmt.Sprintf("IL: %d", len(de.InputField.Text)))
}

// ClearContent clears the content of the DeenEncoder.
func (de *DeenEncoder) ClearContent() {
	de.Content = []byte("")
	de.InputField.SetText("")
}

// Process processes the given data string and returns the processed bytes.
func (de *DeenEncoder) Process() (processed []byte, err error) {
	var reader io.Reader
	if len(de.Content) > 1 {
		reader = bytes.NewReader(de.Content)
	} else {
		if len(de.InputField.Text) > 0 {
			if s := de.InputField.SelectedText(); s != "" {
				reader = strings.NewReader(de.InputField.SelectedText())
			} else {
				reader = strings.NewReader(de.InputField.Text)
			}
		} else {
			return
		}
	}
	if de.Plugin != nil {
		if de.Plugin.ProcessDeenTaskFunc != nil {
			var outWriter bytes.Buffer
			task := types.NewDeenTask(&outWriter)
			task.Reader = reader
			if de.Plugin.Unprocess {
				de.Plugin.UnprocessDeenTaskFunc(task)
			} else {
				de.Plugin.ProcessDeenTaskFunc(task)
			}
			select {
			case err = <-task.ErrChan:
			case <-task.DoneChan:
			}
			processed = outWriter.Bytes()
		} else {
			if de.Plugin.Unprocess {
				processed, err = de.Plugin.UnprocessStreamFunc(reader)
			} else {
				processed, err = de.Plugin.ProcessStreamFunc(reader)
			}
		}
	} else {
		err = errors.New("Plugin not set")
	}
	return
}

// OnChangedWrapper triggers chain processing when text changes.
func (de *DeenEncoder) OnChangedWrapper(input string) {
	de.InputLen.SetText(fmt.Sprintf("IL: %d", len(de.InputField.Text)))
	de.Parent.processChainFrom(de)
}

// NewDeenEncoderWidget initializes a new DeenEconder widget.
func NewDeenEncoderWidget(parent *DeenGUI) (de *DeenEncoder, err error) {
	de = &DeenEncoder{}
	de.Parent = parent
	de.InputField = NewDeenInputField(de)
	de.ViewButton = widget.NewSelect([]string{"Plain", "Hexdump"}, func(mode string) {
		if len(de.Content) < 1 && len(de.InputField.Text) < 1 {
			return
		}
		if mode == "Hexdump" {
			if len(de.Content) < 1 {
				de.Content = []byte(de.InputField.Text)
			}
			processed := hex.Dump(de.Content)
			de.InputField.SetText(processed)
		} else {
			de.InputField.SetText(string(de.Content))
		}
	})
	de.ViewButton.SetSelected("Plain") // Default to plain view
	de.CopyButton = widget.NewButton("Copy", func() {
		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(string(de.Content))
	})
	de.ClearButton = widget.NewButton("Clear", func() {
		de.Parent.RemoveEncoder(de)
	})
	de.InputField.OnChanged = de.OnChangedWrapper
	de.ContentLen = widget.NewLabel(fmt.Sprintf("CL: %d", len(de.Content)))
	de.InputLen = widget.NewLabel(fmt.Sprintf("IL: %d", len(de.InputField.Text)))
	return
}
