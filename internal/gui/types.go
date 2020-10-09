package gui

import (
	"encoding/hex"
	"io/ioutil"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/takeshixx/deen/internal/plugins"
	"github.com/takeshixx/deen/pkg/types"
)

// DeenGUI represents a GUI instance.
type DeenGUI struct {
	App                  fyne.App
	MainWindow           fyne.Window
	MainLayout           *fyne.Container
	PluginList           *widget.ScrollContainer
	Plugins              []string
	EncoderWidgetsScroll *widget.ScrollContainer
	EncoderWidgets       *widget.Box
	Encoders             []*DeenEncoder
	HistoryList          *widget.Group
	History              []string
	CurrentFocus         int // The index of the encoder widget in Encoders
}

// NewDeenGUI initializes a new DeenGUI instance.
func NewDeenGUI() (dg *DeenGUI, err error) {
	dg = &DeenGUI{
		App: app.NewWithID("io.deen.app"),
		PluginList: widget.NewScrollContainer(
			widget.NewAccordionContainer(),
		),
		EncoderWidgetsScroll: widget.NewScrollContainer(widget.NewVBox()),
		EncoderWidgets:       widget.NewVBox(),
		HistoryList:          widget.NewGroup("History"),
	}
	dg.MainWindow = dg.App.NewWindow("deen")
	dg.newMainLayout()
	dg.newMainMenu()
	dg.loadPluginList()

	// Create the root encoder widget (must always exist)
	if _, err = dg.AddEncoder(); err != nil {
		return
	}

	// Setup the theme
	if dg.App.Preferences().String("theme") == "light" {
		dg.App.Settings().SetTheme(theme.LightTheme())
	} else {
		dg.App.Settings().SetTheme(theme.DarkTheme())
	}

	dg.MainWindow.SetMaster()
	dg.MainWindow.SetContent(dg.MainLayout)
	dg.MainWindow.Resize(fyne.NewSize(640, 480))
	dg.addCustomShortcuts()
	dg.updateGUI()
	return
}

// Run is the main function that should
// be called to run the GUI. This will
// block until the GUI is closed.
func (dg *DeenGUI) Run() {
	dg.MainWindow.ShowAndRun()
}

func (dg *DeenGUI) newMainLayout() {
	dg.MainLayout = fyne.NewContainerWithLayout(
		layout.NewBorderLayout(nil, nil, dg.PluginList, dg.HistoryList),
		dg.PluginList,           // left
		dg.HistoryList,          // right
		dg.EncoderWidgetsScroll, // middle
	)
}

func (dg *DeenGUI) newMainMenu() {
	dg.MainWindow.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu("File",
				fyne.NewMenuItem("Open", func() {
					fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
						if err == nil && reader == nil {
							return
						}
						if err != nil {
							dialog.ShowError(err, dg.MainWindow)
							return
						}
						dg.fileOpened(reader)
					}, dg.MainWindow)
					fd.Show()
				}),
				// A quit item will be appended to our first menu
			),
			fyne.NewMenu("Theme",
				fyne.NewMenuItem("Light", func() {
					dg.App.Settings().SetTheme(theme.LightTheme())
					dg.App.Preferences().SetString("theme", "light")
				}),
				fyne.NewMenuItem("Dark", func() {
					dg.App.Settings().SetTheme(theme.DarkTheme())
					dg.App.Preferences().SetString("theme", "dark")
				}),
			),
			fyne.NewMenu("Help",
				fyne.NewMenuItem("About", func() {
					dialog.ShowInformation("About", "deen is a DEcoding/ENcoding application that processes arbitrary input data with a wide range of plugins.", dg.MainWindow)
				}),
			)))
}

// Populate the DeenGUI.PluginList field
func (dg *DeenGUI) loadPluginList() {
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

func (dg *DeenGUI) addCustomShortcuts() {
	ctrlR := desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: desktop.ControlModifier}
	dg.MainWindow.Canvas().AddShortcut(&ctrlR, func(shortcut fyne.Shortcut) {
		// Show fuzzy search
		dg.showPluginSearch()
	})
}

// Reprocess all encoder widgets and update the GUI elements.
func (dg *DeenGUI) updateGUI() {
	log.Println("[DEBUG] Updating GUI")
	// We should only start processing
	// when at least the root widget
	// has a plugin set.
	if dg.Encoders[0].Plugin != nil {
		// We have to process all
		// encoders before creating
		// the GUI layouts.
		dg.processChain()
	}

	dg.updateEncoderWidgets()
	dg.EncoderWidgetsScroll = widget.NewScrollContainer(dg.EncoderWidgets)
	dg.newMainLayout()

	dg.MainWindow.SetContent(dg.MainLayout)
	dg.EncoderWidgetsScroll.ScrollToBottom()
	// Always set focus to the newest encoder.
	dg.SetEncoderFocus(len(dg.Encoders) - 1)
	return
}

func (dg *DeenGUI) updateEncoderWidgets() {
	dg.EncoderWidgets = widget.NewVBox()
	dg.HistoryList = widget.NewGroup("History")
	var historyName string
	for _, e := range dg.Encoders {
		dg.EncoderWidgets.Append(e.createLayout())
		if e.Plugin != nil {
			if e.Plugin.Unprocess {
				historyName = "." + e.Plugin.Name
			} else {
				historyName = e.Plugin.Name
			}
			dg.HistoryList.Append(widget.NewLabel(historyName))
		}
	}
}

func (dg *DeenGUI) fileOpened(f fyne.URIReadCloser) {
	input, err := ioutil.ReadAll(f)
	if err != nil {
		dialog.ShowError(err, dg.MainWindow)
	}
	dg.Encoders[0].SetContent(input)
}

// DeenEncoder represents an encoder that can be added to the GUI's Encoders list.
type DeenEncoder struct {
	Parent      *DeenGUI
	Content     []byte // The actual content of the widget. Should never be changed, only by following encoder widgets.
	ContentLen  *widget.Label
	View        string // The current view (plain/hex)
	Layout      *widget.Box
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

// NewDeenEncoderWidget initializes a new DeenEconder widget.
func NewDeenEncoderWidget(parent *DeenGUI) (de *DeenEncoder, err error) {
	de = &DeenEncoder{
		Parent:     parent,
		InputField: NewDeenInputField(de),
		ContentLen: widget.NewLabel("CL: "),
		InputLen:   widget.NewLabel("IL: "),
	}
	de.InputField.OnChanged = de.OnChangedWrapper
	de.Layout = de.createLayout()
	de.newButtons()
	return
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

func (de *DeenEncoder) newButtons() {
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
