//go:build gui

// Package gui implements the deen desktop interface: a Burp Decoder-style
// chain of plugin transforms backed by the pure internal/pipeline model.
package gui

import (
	"io"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/core"
	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

// DeenGUI is the top-level GUI state.
type DeenGUI struct {
	app    fyne.App
	window fyne.Window

	pipe        *pipeline.Pipeline
	pluginNames []string

	sourceEntry *widget.Entry
	stepsBox    *fyne.Container // holds the source card, step cards and add-slot
	cards       []*stepCard     // parallel to pipe.Steps()

	// updating guards programmatic SetText so it does not re-enter OnChanged.
	updating bool
}

// NewDeenGUI builds the GUI.
func NewDeenGUI() (*DeenGUI, error) {
	dg := &DeenGUI{
		app:         app.NewWithID("io.deen.app"),
		pipe:        pipeline.New(),
		pluginNames: plugins.Names(),
	}
	dg.window = dg.app.NewWindow("deen")
	dg.window.SetMaster()

	dg.stepsBox = container.NewVBox()
	content := container.NewBorder(dg.toolbar(), nil, nil, nil, container.NewVScroll(dg.stepsBox))
	dg.window.SetContent(content)
	dg.window.SetMainMenu(dg.mainMenu())
	dg.window.Resize(fyne.NewSize(900, 640))

	dg.rebuild()
	return dg, nil
}

// Run shows the window and blocks until it closes.
func (dg *DeenGUI) Run() { dg.window.ShowAndRun() }

// toolbar builds the top action bar.
func (dg *DeenGUI) toolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), dg.openFile),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), dg.saveResult),
		widget.NewToolbarAction(theme.ContentCopyIcon(), dg.copyResult),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentClearIcon(), dg.clear),
	)
}

// mainMenu builds the window menu (theme switching).
func (dg *DeenGUI) mainMenu() *fyne.MainMenu {
	setTheme := func(t fyne.Theme) func() { return func() { dg.app.Settings().SetTheme(t) } }
	themeMenu := fyne.NewMenu("Theme",
		fyne.NewMenuItem("System", setTheme(theme.DefaultTheme())),
		fyne.NewMenuItem("Light", setTheme(&forcedVariantTheme{theme.DefaultTheme(), theme.VariantLight})),
		fyne.NewMenuItem("Dark", setTheme(&forcedVariantTheme{theme.DefaultTheme(), theme.VariantDark})),
	)
	help := fyne.NewMenu("Help", fyne.NewMenuItem("About", dg.showAbout))
	return fyne.NewMainMenu(themeMenu, help)
}

// showAbout displays version and project information.
func (dg *DeenGUI) showAbout() {
	version := core.Version()
	if b := core.Branch(); b != "" {
		version += " (" + b + ")"
	}
	docsURL, _ := url.Parse("https://deen.adversec.com")
	repoURL, _ := url.Parse("https://github.com/takeshixx/go-deen")

	title := widget.NewLabelWithStyle("deen", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	content := container.NewVBox(
		title,
		widget.NewLabel("Version: "+version),
		widget.NewLabel("Chain data encodings, decodings, hashes and formatters."),
		container.NewHBox(widget.NewLabel("Documentation:"), widget.NewHyperlink("deen.adversec.com", docsURL)),
		container.NewHBox(widget.NewLabel("Source:"), widget.NewHyperlink("github.com/takeshixx/go-deen", repoURL)),
	)
	dialog.ShowCustom("About deen", "Close", content, dg.window)
}

// rebuild recreates the whole card stack. Called on structural changes
// (adding/removing steps). Content-only changes use refreshFrom instead.
func (dg *DeenGUI) rebuild() {
	dg.stepsBox.RemoveAll()
	dg.cards = dg.cards[:0]

	dg.stepsBox.Add(dg.newSourceCard())
	for i := range dg.pipe.Steps() {
		c := dg.newStepCard(i)
		dg.cards = append(dg.cards, c)
		dg.stepsBox.Add(c.container)
	}
	dg.stepsBox.Add(dg.newAddSlot())
	dg.stepsBox.Refresh()
}

// refreshFrom updates the displayed output of every card from index `from`
// downward without recreating widgets.
func (dg *DeenGUI) refreshFrom(from int) {
	for i := from; i < len(dg.cards); i++ {
		dg.cards[i].refresh()
	}
}

// setText updates an entry programmatically without triggering its OnChanged.
func (dg *DeenGUI) setText(e *widget.Entry, s string) {
	dg.updating = true
	e.SetText(s)
	dg.updating = false
}

// --- toolbar actions ---

func (dg *DeenGUI) openFile() {
	dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			dialog.ShowError(err, dg.window)
			return
		}
		dg.pipe.SetSource(data)
		dg.setText(dg.sourceEntry, string(data))
		dg.refreshFrom(0)
	}, dg.window)
}

func (dg *DeenGUI) saveResult() {
	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		defer wc.Close()
		if _, err := wc.Write(dg.pipe.Result()); err != nil {
			dialog.ShowError(err, dg.window)
		}
	}, dg.window)
}

func (dg *DeenGUI) copyResult() {
	dg.window.Clipboard().SetContent(string(dg.pipe.Result()))
}

func (dg *DeenGUI) clear() {
	dg.pipe = pipeline.New()
	dg.rebuild()
}
