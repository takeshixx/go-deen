package gui

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

// SetContent overwrites the Content and InputField data.
func (de *DeenEncoder) SetContent(data []byte) {
	de.Content = data
	de.ContentLen.SetText(fmt.Sprintf("CL: %d", len(de.Content)))
	de.InputField.SetText(string(data))
	de.InputLen.SetText(fmt.Sprintf("IL: %d", len(de.InputField.Text)))
}

// GetContent returns the content of the encoder widget.
func (de *DeenEncoder) GetContent() []byte {
	if len(de.Content) > 0 {
		return de.Content
	}
	curData := de.InputField.Text
	if len(curData) > 0 {
		return []byte(curData)
	}
	return []byte("")
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
