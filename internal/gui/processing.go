// +build gui

package gui

import (
	"log"

	"fyne.io/fyne/v2/dialog"
	"github.com/takeshixx/deen/internal/plugins"
)

// RunPlugin executes a given plugin
func (dg *DeenGUI) RunPlugin(pluginCmd string) {
	plugin := plugins.GetForCmd(pluginCmd)
	if plugin == nil {
		return
	}
	log.Printf("[DEBUG] Found plugin: %s\n", plugin.Name)
	ce := dg.CurrentEncoder()
	ce.Plugin = plugin
	dg.updateGUI()
}

func (dg *DeenGUI) showPluginSearch() {
	search := NewDeenSearchField(dg)
	search.Dialog = dialog.NewCustomConfirm("Search Plugin", "Run", "Cancel", search.Layout, search.ConfirmCallBack, dg.MainWindow)
	search.Show()
	dg.MainWindow.Canvas().Focus(search)
}

// Process the Encoders chain starting from the root widget.
func (dg *DeenGUI) processChain() (err error) {
	return dg.processChainFrom(dg.Encoders[0])
}

// Process through the whole Encoders chain, starting from a given DeenEncoder.
func (dg *DeenGUI) processChainFrom(ew *DeenEncoder) (err error) {
	encodersIndex := 0
	if ew != dg.Encoders[0] {
		// We are not starting at the root widget
		for i, e := range dg.Encoders {
			if e == ew {
				encodersIndex = i
				break
			}
		}
	}

	var processed []byte
	var nextEncoder *DeenEncoder
	for i, e := range dg.Encoders {
		if i < encodersIndex {
			// Skip encoders before the current one
			continue
		}
		processed, err = e.Process()
		if err != nil {
			return
		}
		if len(processed) < 1 {
			log.Printf("[DEBUG] processed length is smaller than 1")
			// TODO: should we remove the following widgets?
			if i < len(dg.Encoders)-1 {
				log.Printf("[WARN] processed empty, but not last widget")
				for _, de := range dg.Encoders[i:] {
					log.Printf("[DEBUG] Removing decoder %v\n", de)
					dg.RemoveEncoder(de)
				}
				return
			}
		} else {
			nextEncoder, err = dg.NextEncoder(e)
			if err != nil {
				log.Printf("[DEBUG] No next encoder found, creating a new one")
				nextEncoder, err = dg.AddEncoder()
				nextEncoder.SetContent(processed)
				return
			}
			nextEncoder.SetContent(processed)
		}
	}
	return
}
