//go:build gui

// Package gui implements the deen desktop interface: a Burp Decoder-style
// chain of plugin transforms backed by the pure internal/pipeline model.
package gui

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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
	sourceMeta  *widget.Label
	stepsBox    *fyne.Container // holds the source card, step cards and add-slot
	cards       []*stepCard     // parallel to pipe.Steps()
	history     *fyne.Container // side panel listing the chain
	split       *container.Split
	tabs        *container.AppTabs
	historyOpen bool

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
	dg.history = container.NewVBox()
	dg.historyOpen = true
	dg.tabs = container.NewAppTabs(
		container.NewTabItem("Home", dg.homeTab()),
		container.NewTabItem("Plugins", dg.pluginsTab()),
		container.NewTabItem("About", dg.aboutTab()),
	)
	dg.tabs.SetTabLocation(container.TabLocationTop)
	content := container.NewBorder(dg.toolbar(), nil, nil, nil, dg.tabs)
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
		widget.NewToolbarAction(theme.MailForwardIcon(), dg.copyCommand),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.FileTextIcon(), dg.openChain),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), dg.saveChain),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.NavigateBackIcon(), dg.undo),
		widget.NewToolbarAction(theme.NavigateNextIcon(), dg.redo),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentAddIcon(), dg.showPluginSearch),
		widget.NewToolbarAction(theme.HistoryIcon(), dg.showPresets),
		widget.NewToolbarAction(theme.SearchIcon(), dg.showSuggestions),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentClearIcon(), dg.clear),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ListIcon(), dg.toggleHistory),
		widget.NewToolbarAction(theme.HelpIcon(), dg.showHelp),
	)
}

// toggleHistory collapses or expands the history side panel.
func (dg *DeenGUI) toggleHistory() {
	dg.historyOpen = !dg.historyOpen
	if dg.historyOpen {
		dg.split.SetOffset(0.78)
	} else {
		dg.split.SetOffset(1.0)
	}
}

// mainMenu builds the window menu (theme switching).
func (dg *DeenGUI) mainMenu() *fyne.MainMenu {
	setTheme := func(t fyne.Theme) func() { return func() { dg.app.Settings().SetTheme(t) } }
	chainMenu := fyne.NewMenu("Chain",
		fyne.NewMenuItem("Open chain", dg.openChain),
		fyne.NewMenuItem("Save chain", dg.saveChain),
		fyne.NewMenuItem("Presets", dg.showPresets),
	)
	themeMenu := fyne.NewMenu("Theme",
		fyne.NewMenuItem("System", setTheme(theme.DefaultTheme())),
		fyne.NewMenuItem("Light", setTheme(&forcedVariantTheme{theme.DefaultTheme(), theme.VariantLight})),
		fyne.NewMenuItem("Dark", setTheme(&forcedVariantTheme{theme.DefaultTheme(), theme.VariantDark})),
	)
	help := fyne.NewMenu("Help",
		fyne.NewMenuItem("How to use", dg.showHelp),
		fyne.NewMenuItem("Plugin catalog", func() { dg.tabs.SelectIndex(1) }),
		fyne.NewMenuItem("About", func() { dg.tabs.SelectIndex(2) }),
	)
	return fyne.NewMainMenu(chainMenu, themeMenu, help)
}

func (dg *DeenGUI) homeTab() fyne.CanvasObject {
	chain := container.NewVScroll(dg.stepsBox)
	historyPanel := widget.NewCard("History", "", container.NewVScroll(dg.history))
	dg.split = container.NewHSplit(chain, historyPanel)
	dg.split.SetOffset(0.78)
	return dg.split
}

// showHelp displays a usage/info page describing the GUI.
func (dg *DeenGUI) showHelp() {
	md := `## Using deen

deen applies a **chain of transforms** to your input — like Burp Suite's
Decoder. The result of each step feeds into the next.

### Steps
- Type or open data into the **Input** card at the top.
- Use the **Add transform** card to append a step: each plugin category
  (codecs, compressions, hashs, formatters, misc) has its own dropdown.
- Tick **decode** to run a step in reverse (e.g. base64 decode). One-way
  plugins like hashes cannot be decoded.
- Plugin options (e.g. base64 ` + "`-url`" + `, gzip ` + "`-level`" + `) appear as fields
  under each step.
- **hex** shows a step's output as a hex dump (read-only).
- Use **Open chain** and **Save chain** to reuse complete transform chains.
- Use the plus icon to search the plugin catalog and append a transform.
- Use **Presets** to load starter chains while keeping the current input.
- Use the search icon to detect likely next transforms from the current result.
- Copy the equivalent shell pipeline from the toolbar.
- Editing any step's output recomputes everything below it.
- Use the disclosure arrow to **collapse/expand** a step, the trash icon
  to remove it.

### Toolbar
- Open a file, save the final result, copy the result, or clear the chain.
- Toggle the **History** side panel, which lists the whole chain.

### Menu
- **Theme**: follow the system theme or force light/dark.`

	rich := widget.NewRichTextFromMarkdown(md)
	rich.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(rich)
	scroll.SetMinSize(fyne.NewSize(520, 420))
	dialog.ShowCustom("How to use deen", "Close", scroll, dg.window)
}

// showAbout displays version and project information.
func (dg *DeenGUI) showAbout() {
	dg.tabs.SelectIndex(2)
}

func (dg *DeenGUI) aboutTab() fyne.CanvasObject {
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
		widget.NewLabel("deen encodes, decodes, hashes, compresses and formats data\nthrough a configurable chain of plugins."),
		widget.NewLabel("Built with Go, Fyne (desktop GUI) and WebAssembly (web)."),
		container.NewHBox(widget.NewLabel("Documentation:"), widget.NewHyperlink("deen.adversec.com", docsURL)),
		container.NewHBox(widget.NewLabel("Source:"), widget.NewHyperlink("github.com/takeshixx/go-deen", repoURL)),
		widget.NewSeparator(),
		widget.NewLabel("All processing happens locally in the desktop GUI. The web UI runs the same pipeline model in WebAssembly."),
	)
	return container.NewPadded(container.NewVScroll(content))
}

func (dg *DeenGUI) pluginsTab() fyne.CanvasObject {
	list := container.NewVBox()
	currentCategory := ""
	for _, info := range plugins.UICatalog() {
		if info.Category != currentCategory {
			currentCategory = info.Category
			list.Add(widget.NewLabelWithStyle(plugins.CategoryLabel(currentCategory), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		}
		list.Add(pluginInfoCard(info))
	}
	return container.NewPadded(container.NewVScroll(list))
}

func pluginInfoCard(info plugins.UIPluginInfo) fyne.CanvasObject {
	direction := "Encode only"
	if info.CanDecode {
		direction = "Encode and decode"
	}
	meta := []string{info.Category, direction}
	if len(info.Aliases) > 0 {
		meta = append(meta, "Aliases: "+strings.Join(info.Aliases, ", "))
	}

	body := container.NewVBox(
		widget.NewLabel(strings.Join(meta, " · ")),
		widget.NewLabel(info.Description),
		widget.NewLabel("Use for: "+info.UseFor),
	)
	for _, ref := range info.References {
		u, err := url.Parse(ref.URL)
		if err == nil {
			body.Add(container.NewHBox(widget.NewLabel("Reference:"), widget.NewHyperlink(ref.Label, u)))
		}
	}
	return widget.NewCard(info.Name, "", body)
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
	dg.updateHistory()
}

// updateHistory redraws the side panel listing the chain of steps.
func (dg *DeenGUI) updateHistory() {
	dg.history.RemoveAll()
	dg.history.Add(widget.NewLabelWithStyle("Input", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	for i, s := range dg.pipe.Steps() {
		name := s.Plugin
		if name == "" {
			name = "(none)"
		}
		dir := "encode"
		if s.Unprocess {
			dir = "decode"
		}
		if s.Disabled {
			dir += ", disabled"
		}
		line := canvas.NewText(fmt.Sprintf("%d. %s (%s)", i+1, name, dir), accent(i))
		dg.history.Add(line)
	}
	dg.history.Refresh()
}

// refreshFrom updates the displayed output of every card from index `from`
// downward without recreating widgets.
func (dg *DeenGUI) refreshFrom(from int) {
	if dg.sourceMeta != nil {
		dg.sourceMeta.SetText(pipeline.DataMetadata(dg.pipe.Source(), 0).Summary())
	}
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

func (dg *DeenGUI) openChain() {
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
		if err := dg.pipe.ImportJSON(data); err != nil {
			dialog.ShowError(err, dg.window)
			return
		}
		dg.rebuild()
	}, dg.window)
}

func (dg *DeenGUI) saveChain() {
	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		defer wc.Close()
		data, err := dg.pipe.ExportJSON()
		if err != nil {
			dialog.ShowError(err, dg.window)
			return
		}
		if _, err := wc.Write(data); err != nil {
			dialog.ShowError(err, dg.window)
		}
	}, dg.window)
}

func (dg *DeenGUI) copyResult() {
	dg.window.Clipboard().SetContent(string(dg.pipe.Result()))
}

func (dg *DeenGUI) copyCommand() {
	command := dg.pipe.CommandLine()
	if command == "" {
		dialog.ShowInformation("Command line", "No enabled transforms to export.", dg.window)
		return
	}
	dg.window.Clipboard().SetContent(command)
	entry := widget.NewMultiLineEntry()
	entry.SetText(command)
	entry.Wrapping = fyne.TextWrapBreak
	entry.SetMinRowsVisible(4)
	entry.Disable()
	dialog.ShowCustom("Command copied", "Close", entry, dg.window)
}

func (dg *DeenGUI) showSuggestions() {
	suggestions := pipeline.Suggestions(dg.pipe.Result())
	list := container.NewVBox()
	if len(suggestions) == 0 {
		list.Add(widget.NewLabel("No likely transforms detected."))
		dialog.ShowCustom("Suggested transforms", "Close", list, dg.window)
		return
	}

	var d dialog.Dialog
	for _, s := range suggestions {
		s := s
		label := s.Label
		if s.Reason != "" {
			label += " - " + s.Reason
		}
		list.Add(widget.NewButton(label, func() {
			dg.pipe.AddStep(s.Plugin, s.Unprocess)
			dg.rebuild()
			if d != nil {
				d.Hide()
			}
		}))
	}
	d = dialog.NewCustom("Suggested transforms", "Close", container.NewVScroll(list), dg.window)
	d.Resize(fyne.NewSize(560, 360))
	d.Show()
}

func (dg *DeenGUI) showPluginSearch() {
	query := widget.NewEntry()
	query.SetPlaceHolder("Search plugins")
	results := container.NewVBox()
	scroll := container.NewVScroll(results)
	scroll.SetMinSize(fyne.NewSize(640, 420))

	var d dialog.Dialog
	refresh := func(q string) {
		results.RemoveAll()
		matches := plugins.SearchUICatalog(q)
		if len(matches) == 0 {
			results.Add(widget.NewLabel("No plugins found."))
		}
		for _, info := range matches {
			info := info
			direction := "encode"
			if !info.CanDecode {
				direction = "run"
			}
			title := fmt.Sprintf("%s / %s", plugins.CategoryLabel(info.Category), info.Name)
			if len(info.Aliases) > 0 {
				title += " (" + strings.Join(info.Aliases, ", ") + ")"
			}
			desc := widget.NewLabel(info.Description)
			desc.Wrapping = fyne.TextWrapWord
			addEncode := widget.NewButton("Add "+direction, func() {
				dg.pipe.AddStep(info.Name, false)
				dg.rebuild()
				if d != nil {
					d.Hide()
				}
			})
			actions := container.NewHBox(addEncode)
			if info.CanDecode {
				actions.Add(widget.NewButton("Add decode", func() {
					dg.pipe.AddStep(info.Name, true)
					dg.rebuild()
					if d != nil {
						d.Hide()
					}
				}))
			}
			results.Add(widget.NewCard(title, "", container.NewVBox(desc, actions)))
		}
		results.Refresh()
	}
	query.OnChanged = refresh
	refresh("")

	d = dialog.NewCustom("Add transform", "Close", container.NewBorder(query, nil, nil, nil, scroll), dg.window)
	d.Resize(fyne.NewSize(700, 520))
	d.Show()
	dg.window.Canvas().Focus(query)
}

func (dg *DeenGUI) showPresets() {
	list := container.NewVBox()
	var d dialog.Dialog
	for _, preset := range pipeline.BuiltinPresets() {
		preset := preset
		desc := widget.NewLabel(preset.Description)
		desc.Wrapping = fyne.TextWrapWord
		apply := widget.NewButton("Apply", func() {
			dg.pipe.ApplyPreset(preset)
			dg.rebuild()
			if d != nil {
				d.Hide()
			}
		})
		list.Add(widget.NewCard(preset.Name, "", container.NewVBox(desc, apply)))
	}
	scroll := container.NewVScroll(list)
	scroll.SetMinSize(fyne.NewSize(640, 420))
	d = dialog.NewCustom("Presets", "Close", scroll, dg.window)
	d.Resize(fyne.NewSize(700, 520))
	d.Show()
}

func (dg *DeenGUI) undo() {
	if dg.pipe.Undo() {
		dg.rebuild()
	}
}

func (dg *DeenGUI) redo() {
	if dg.pipe.Redo() {
		dg.rebuild()
	}
}

func (dg *DeenGUI) clear() {
	dg.pipe.Clear()
	dg.rebuild()
}
