package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
	"github.com/schollz/closestmatch"
	"github.com/takeshixx/deen/internal/plugins"
)

// RunPlugin executes a given plugin
func (dg *DeenGUI) RunPlugin(pluginCmd string) {
	plugin := plugins.GetForCmd(pluginCmd)
	log.Printf("[DEBUG] Found plugin: %s\n", plugin.Name)
	ce := dg.CurrentEncoder()
	ce.Plugin = plugin
	dg.updateGUI()
}

func (dg *DeenGUI) showPluginSearch() {
	content := widget.NewEntry()
	content.SetPlaceHolder("Type plugin name")

	var closest []string
	bagSizes := []int{2}
	cm := closestmatch.New(dg.Plugins, bagSizes)

	layout := widget.NewVBox()
	layout.Append(content)

	content.OnChanged = func(text string) {
		fmt.Println("Entered:", text)
		closest = cm.ClosestN(text, 5)
		fmt.Println("Closest:", closest)

		layout.Children = []fyne.CanvasObject{}
		layout.Append(content)

		for _, s := range closest {
			layout.Append(widget.NewButton(s, func() {}))
		}
	}

	dialog.ShowCustom("Search Plugin", "Cancel", layout, dg.MainWindow)
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
