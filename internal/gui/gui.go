// +build gui

package gui

import (
	"errors"
	"log"
)

// CurrentEncoder returns the currently focussed encoder widget.
func (dg *DeenGUI) CurrentEncoder() (ce *DeenEncoder) {
	if dg.CurrentFocus > len(dg.Encoders)-1 {
		// Invalid state, use last one
		ce = dg.Encoders[len(dg.Encoders)-1]
	} else {
		ce = dg.Encoders[dg.CurrentFocus]
	}
	return
}

// AddEncoder creates and adds a new DeenEncoder instance to the Encoders list.
func (dg *DeenGUI) AddEncoder() (enc *DeenEncoder, err error) {
	enc, err = NewDeenEncoderWidget(dg)
	if err != nil {
		return
	}
	dg.Encoders = append(dg.Encoders, enc)
	return
}

// RemoveEncoder removes a given DeenEncoder from the Encoders list.
func (dg *DeenGUI) RemoveEncoder(enc *DeenEncoder) {
	if enc == dg.Encoders[0] {
		// We cannot remove the root widget, just clearing content and plugin
		enc.ClearContent()
		dg.Encoders[0].Plugin = nil
		// And remove all following widgets.
		dg.Encoders = []*DeenEncoder{dg.Encoders[0]}
	} else {
		// If enc is not the root widget, we have a previous widget
		previous, err := dg.PreviousEncoder(enc)
		if err != nil {
			log.Printf("[WARN] PreviousEncoder() failed: %v\n", err)
		}
		if enc == dg.Encoders[len(dg.Encoders)-1] {
			// Remove the last encoder
			dg.Encoders = dg.Encoders[:len(dg.Encoders)-1]
			// And clear the plugin of the previous encoder
			previous.Plugin = nil
		} else {
			for i, e := range dg.Encoders {
				if e == enc {
					dg.Encoders = append(dg.Encoders[:i], dg.Encoders[i+1:]...)
					// Transfer plugin to previous widget
					dg.Encoders[i-1].Plugin = e.Plugin
					break
				}
			}
		}
	}
	dg.updateGUI()
}

// SetEncoderFocus sets focus of the encoder widget on index.
func (dg *DeenGUI) SetEncoderFocus(index int) {
	// Make sure we do not reference an invalid index
	if index < 0 || len(dg.Encoders)-1 < index {
		return
	}
	// Set focus on referenced encoder widget
	dg.MainWindow.Canvas().Focus(dg.Encoders[index].InputField)
	// Set the cursor to the end of the input field
	dg.Encoders[index].InputField.CursorColumn = len(dg.Encoders[index].InputField.Text)
	// Refresh the widget to make the changes take effect
	dg.Encoders[index].InputField.Refresh()
	// Update the global CurrentFocus
	dg.CurrentFocus = index
}

// NextEncoder returns the next encoder instances from Encoders.
func (dg *DeenGUI) NextEncoder(pe *DeenEncoder) (ne *DeenEncoder, err error) {
	for i, e := range dg.Encoders {
		if e == pe {
			if len(dg.Encoders)-1 < i+1 {
				// There is no no next widget, create a new one.
				//ne, err = dg.AddEncoder()
				err = errors.New("No next encoder found")
				return
			}
			ne = dg.Encoders[i+1]
			return
		}
	}
	return
}

// PreviousEncoder returns the previous encoder instances from Encoders.
func (dg *DeenGUI) PreviousEncoder(ne *DeenEncoder) (pe *DeenEncoder, err error) {
	if ne == dg.Encoders[0] {
		err = errors.New("Root widget has no previous encoders")
		return
	}
	for i, e := range dg.Encoders {
		if e == ne {
			pe = dg.Encoders[i-1]
			return
		}
	}
	return
}
