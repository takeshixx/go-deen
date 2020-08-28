package gui

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/schollz/closestmatch"
	"github.com/takeshixx/deen/internal/plugins"
)

// DeenGUI represents a GUI instance.
type DeenGUI struct {
	App               fyne.App
	MainWindow        fyne.Window
	Layout            *fyne.Container
	PluginList        *widget.ScrollContainer
	Plugins           []string
	EncoderListScroll *widget.ScrollContainer
	EncoderList       *widget.Box
	Encoders          []*DeenEncoder
	HistoryList       *widget.Group
	History           []string
	CurrentFocus      int // The index of the encoder widget in Encoders
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

// RunPlugin executes a given plugin
func (dg *DeenGUI) RunPlugin(pluginCmd string) {
	plugin := plugins.GetForCmd(pluginCmd)
	log.Printf("[DEBUG] Found plugin: %s\n", plugin.Name)
	ce := dg.CurrentEncoder()
	ce.Plugin = plugin
	//dg.processChain()
	dg.updateGUI()

	// TODO: always set focus to last encoder widget?
	//dg.SetEncoderFocus(dg.CurrentFocus + 1)
	dg.SetEncoderFocus(len(dg.Encoders) - 1)
}

// Reprocess all encoder widgets and update the GUI elements.
func (dg *DeenGUI) updateGUI() (err error) {
	log.Println("[DEBUG] Updating GUI")
	dg.EncoderList = widget.NewVBox()
	dg.HistoryList = widget.NewGroup("History")
	var historyName string
	for _, e := range dg.Encoders {
		dg.EncoderList.Append(e.createLayout())
		if e.Plugin != nil {
			if e.Plugin.Unprocess {
				historyName = "." + e.Plugin.Name
			} else {
				historyName = e.Plugin.Name
			}
			dg.HistoryList.Append(widget.NewLabel(historyName))
		}
	}
	dg.Layout = fyne.NewContainerWithLayout(
		layout.NewBorderLayout(nil, nil, dg.PluginList, dg.HistoryList),
		dg.PluginList,  // left
		dg.HistoryList, // right
		dg.EncoderList, // middle
	)
	dg.processChain()
	dg.MainWindow.SetContent(dg.Layout)
	return
}

// AddEncoder creates and adds a new DeenEncoder instance to the Encoders list.
func (dg *DeenGUI) AddEncoder() (enc *DeenEncoder, err error) {
	enc, err = NewDeenEncoderWidget(dg)
	if err != nil {
		return
	}
	dg.Encoders = append(dg.Encoders, enc)
	// Create the layout and add it to the EncoderList
	dg.EncoderList.Append(enc.createLayout())
	dg.EncoderListScroll = widget.NewScrollContainer(dg.EncoderList)
	dg.EncoderListScroll.SetMinSize(fyne.NewSize(dg.EncoderList.MinSize().Width, 0))
	dg.EncoderListScroll.Refresh()
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

// Populate the DeenGUI.PluginList field
func (dg *DeenGUI) loadPluginList() (err error) {
	dg.Plugins = []string{}
	pluginList := widget.NewAccordionContainer()

	var pluginGroup *widget.AccordionItem
	for _, c := range plugins.PluginCategories {

		filteredPlugins := plugins.GetForCategory(c, false)

		var groupList *widget.Box
		groupList = widget.NewVBox()
		for _, p := range filteredPlugins {
			pluginName := p
			groupList.Append(widget.NewButton(p, func() {
				dg.RunPlugin(pluginName)
			}))
		}
		allPlugins := plugins.GetForCategory(c, true)
		for _, p := range allPlugins {
			pluginName := p
			dg.Plugins = append(dg.Plugins, pluginName)
		}

		pluginGroup = widget.NewAccordionItem(c, groupList)
		pluginList.Append(pluginGroup)
	}

	dg.PluginList = widget.NewScrollContainer(pluginList)
	// Ensure that the scroll container is wide enough
	dg.PluginList.SetMinSize(fyne.NewSize(pluginList.MinSize().Width, 0))
	dg.PluginList.Refresh()
	return
}

// Populate the DeenGUI.EncoderList field
func (dg *DeenGUI) loadEncoderList() (err error) {
	dg.EncoderList = widget.NewVBox()
	_, err = dg.AddEncoder() // Create the root encoder widget (must always exist)
	if err != nil {
		return
	}
	// Set initial focus to root widget
	dg.SetEncoderFocus(0)

	// Ensure that the scroll container is wide enough
	/* 	dg.EncoderListScroll.SetMinSize(fyne.NewSize(dg.EncoderList.MinSize().Width, 0))
	dg.EncoderListScroll.Refresh() */

	return
}

func (dg *DeenGUI) addCustomShortcuts() {
	ctrlR := desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: desktop.ControlModifier}
	dg.MainWindow.Canvas().AddShortcut(&ctrlR, func(shortcut fyne.Shortcut) {
		// Show fuzzy search
		dg.showPluginSearch()
	})
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

func (dg *DeenGUI) fileOpened(f fyne.URIReadCloser) {
	input, err := ioutil.ReadAll(f)
	if err != nil {
		dialog.ShowError(err, dg.MainWindow)
	}
	dg.Encoders[0].SetContent(input)
}

// NewDeenGUI initializes a new DeenGUI instance.
func NewDeenGUI(a fyne.App, w fyne.Window) (dg *DeenGUI, err error) {
	dg = &DeenGUI{}
	dg.App = a
	dg.MainWindow = w
	err = dg.loadPluginList()
	if err != nil {
		return
	}
	err = dg.loadEncoderList()
	if err != nil {
		return
	}
	dg.addCustomShortcuts()
	dg.HistoryList = widget.NewGroup("History")
	dg.Layout = fyne.NewContainerWithLayout(
		layout.NewBorderLayout(nil, nil, dg.PluginList, dg.HistoryList),
		dg.PluginList,        // left
		dg.HistoryList,       // right
		dg.EncoderListScroll, // middle
	)
	return
}
