//go:build gui

// Package gui implements the deen desktop interface: a Burp Decoder-style
// chain of plugin transforms backed by the pure internal/pipeline model.
package gui

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"
	"unicode/utf8"

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
	sourceName  string
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
		container.NewTabItem("Home", compactMinWidth(dg.homeTab())),
		container.NewTabItem("Examples", compactMinWidth(dg.examplesTab())),
		container.NewTabItem("Plugins", compactMinWidth(dg.pluginsTab())),
		container.NewTabItem("About", compactMinWidth(dg.aboutTab())),
	)
	dg.tabs.SetTabLocation(container.TabLocationTop)
	content := container.NewBorder(compactMinWidth(dg.toolbar()), nil, nil, nil, dg.tabs)
	dg.window.SetContent(content)
	dg.window.SetMainMenu(dg.mainMenu())
	dg.window.Resize(fyne.NewSize(760, 560))

	dg.rebuild()
	return dg, nil
}

// Run shows the window and blocks until it closes.
func (dg *DeenGUI) Run() { dg.window.ShowAndRun() }

func compactMinWidth(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.New(cappedMinWidthLayout{width: compactControlMinWidth}, obj)
}

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
		widget.NewToolbarAction(theme.HistoryIcon(), dg.showPresets),
		widget.NewToolbarAction(theme.ContentClearIcon(), dg.clear),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ListIcon(), dg.toggleHistory),
		widget.NewToolbarAction(theme.HelpIcon(), dg.showHelp),
	)
}

// toggleHistory collapses or expands the transformer-chain side panel.
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
		fyne.NewMenuItem("Examples", func() { dg.tabs.SelectIndex(1) }),
		fyne.NewMenuItem("Plugin catalog", func() { dg.tabs.SelectIndex(2) }),
		fyne.NewMenuItem("About", func() { dg.tabs.SelectIndex(3) }),
	)
	return fyne.NewMainMenu(chainMenu, themeMenu, help)
}

func (dg *DeenGUI) homeTab() fyne.CanvasObject {
	chain := container.NewVScroll(dg.stepsBox)
	historyPanel := widget.NewCard("Transformer Chain", "", container.NewBorder(
		container.NewHBox(widget.NewButtonWithIcon("Compare", theme.ViewFullScreenIcon(), dg.showCompare)),
		nil,
		nil,
		nil,
		container.NewVScroll(dg.history),
	))
	dg.split = container.NewHSplit(chain, container.New(cappedMinWidthLayout{width: 180}, historyPanel))
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
- Use the **Add transformer step** card to append a step: each plugin category
  (codecs, compressions, hashs, formatters, misc) has its own dropdown.
- Tick **decode** to run a step in reverse (e.g. base64 decode). One-way
  plugins like hashes cannot be decoded.
- Plugin options (e.g. base64 ` + "`-url`" + `, gzip ` + "`-level`" + `) appear as fields
  under each step.
- **hex** shows a step's output as a hex dump (read-only).
- Use **Open chain** and **Save chain** to reuse complete transform chains.
- Use **Search transformers** in the **Add transformer step** card to search the catalog and append a transform.
- Use **Presets** to load starter chains while keeping the current input.
- Use **Detect next** in the **Add transformer step** card to detect likely next transforms from the current result.
- Use **Compare** in the Transformer Chain panel to inspect any two pipeline points side by side.
- Copy the equivalent shell pipeline from the toolbar.
- Editing any step's output recomputes everything below it.
- Use the disclosure arrow to **collapse/expand** a step, the trash icon
  to remove it.

### Toolbar
- Open a file, save the final result, copy the result, or clear the chain.
- Toggle the **Transformer Chain** side panel, which lists the whole chain.

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
	dg.tabs.SelectIndex(3)
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

func (dg *DeenGUI) examplesTab() fyne.CanvasObject {
	query := widget.NewEntry()
	query.SetPlaceHolder("Search examples")
	list := container.NewVBox()
	render := func(q string) {
		list.RemoveAll()
		matches := 0
		for _, example := range pipeline.BuiltinExamples() {
			example := example
			if !pipeline.ExampleMatches(example, q) {
				continue
			}
			matches++
			list.Add(dg.exampleCard(example))
		}
		if matches == 0 {
			list.Add(widget.NewLabel("No examples found."))
		}
		list.Refresh()
	}
	query.OnChanged = render
	render("")
	return container.NewPadded(container.NewBorder(query, nil, nil, nil, container.NewVScroll(list)))
}

func (dg *DeenGUI) exampleCard(example pipeline.Example) fyne.CanvasObject {
	desc := widget.NewLabel(example.Description)
	desc.Wrapping = fyne.TextWrapWord

	chain := widget.NewLabel(exampleChainSummary(example.Steps))
	chain.Importance = widget.LowImportance
	chain.Wrapping = fyne.TextWrapBreak

	source := widget.NewLabel(exampleSourceSummary(example.Source))
	source.Importance = widget.LowImportance
	source.Wrapping = fyne.TextWrapBreak

	result, err := pipeline.ExampleResult(example)
	outputSummary := widget.NewLabel("Output: " + pipeline.DataMetadata(result, len(example.Source)).Summary())
	outputSummary.Importance = widget.LowImportance
	outputSummary.Wrapping = fyne.TextWrapBreak
	if err != nil {
		outputSummary.SetText("Output error: " + err.Error())
		outputSummary.Importance = widget.DangerImportance
	}

	inputEntry := multilineEntry(5)
	inputEntry.SetText(exampleDataText(example.Source))
	inputEntry.Disable()
	outputEntry := multilineEntry(5)
	outputEntry.SetText(exampleDataText(result))
	outputEntry.Disable()
	dataGrid := container.NewGridWithColumns(2,
		widget.NewCard("Input data", "", inputEntry),
		widget.NewCard("Output result", "", outputEntry),
	)

	load := widget.NewButtonWithIcon("Load example", theme.MediaPlayIcon(), func() {
		dg.pipe.ApplyExample(example)
		dg.rebuild()
		dg.tabs.SelectIndex(0)
	})

	body := container.NewVBox(desc, source, outputSummary, chain)
	if example.WantContains != "" {
		want := widget.NewLabel("Expected result contains: " + example.WantContains)
		want.Importance = widget.LowImportance
		want.Wrapping = fyne.TextWrapBreak
		body.Add(want)
	}
	body.Add(load)
	details := container.New(cappedMinWidthLayout{width: compactControlMinWidth}, container.NewVBox(body, dataGrid))
	accordion := widget.NewAccordion(widget.NewAccordionItem(example.Name, details))
	accordion.CloseAll()
	return container.NewPadded(accordion)
}

func exampleSourceSummary(source []byte) string {
	return "Input: " + pipeline.DataMetadata(source, 0).Summary()
}

func exampleChainSummary(steps []pipeline.PresetStep) string {
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		name := plugins.PluginLabel(step.Plugin)
		if step.Unprocess {
			name = "." + name
		}
		parts = append(parts, name)
	}
	return "Chain: " + strings.Join(parts, " -> ")
}

func exampleDataText(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}
	if looksReadable(data) {
		return string(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func looksReadable(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	for _, b := range data {
		if b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		if b < 0x20 || b == 0x7f {
			return false
		}
	}
	return true
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
	meta := []string{plugins.CategoryLabel(info.Category), direction}
	if info.Label != info.Name {
		meta = append(meta, "Command: "+info.Name)
	}
	if len(info.Aliases) > 0 {
		meta = append(meta, "Aliases: "+strings.Join(info.Aliases, ", "))
	}

	metaLabel := widget.NewLabel(strings.Join(meta, " · "))
	metaLabel.Wrapping = fyne.TextWrapBreak
	desc := widget.NewLabel(info.Description)
	desc.Wrapping = fyne.TextWrapWord
	useFor := widget.NewLabel("Use for: " + info.UseFor)
	useFor.Wrapping = fyne.TextWrapWord
	body := container.NewVBox(metaLabel, desc, useFor)
	for _, ex := range info.Examples {
		exampleLabel := widget.NewLabel("Example: " + ex.Label)
		exampleLabel.Wrapping = fyne.TextWrapBreak
		input := widget.NewLabel("Input: " + ex.Input)
		input.Wrapping = fyne.TextWrapBreak
		output := widget.NewLabel("Output: " + ex.Output)
		output.Wrapping = fyne.TextWrapBreak
		body.Add(exampleLabel)
		body.Add(input)
		body.Add(output)
	}
	for _, ref := range info.References {
		u, err := url.Parse(ref.URL)
		if err == nil {
			body.Add(container.NewHBox(widget.NewLabel("Reference:"), widget.NewHyperlink(ref.Label, u)))
		}
	}
	return widget.NewCard(info.Label, "", body)
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
		name := plugins.PluginLabel(s.Plugin)
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
		dg.sourceMeta.SetText(dg.sourceMetadataSummary())
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
		dg.sourceName = rc.URI().Name()
		dg.pipe.SetSource(data)
		dg.setText(dg.sourceEntry, string(data))
		dg.refreshFrom(0)
	}, dg.window)
}

func (dg *DeenGUI) sourceMetadataSummary() string {
	summary := pipeline.DataMetadata(dg.pipe.Source(), 0).Summary()
	if strings.TrimSpace(dg.sourceName) == "" {
		return summary
	}
	return "source: " + dg.sourceName + " · " + summary
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

type comparePoint struct {
	Label string
	Data  []byte
}

func (dg *DeenGUI) comparePoints() []comparePoint {
	points := []comparePoint{{Label: "Input", Data: dg.pipe.Source()}}
	for i := range dg.pipe.Steps() {
		points = append(points, comparePoint{
			Label: fmt.Sprintf("Step %d output", i+1),
			Data:  dg.pipe.Output(i),
		})
	}
	return points
}

func compareLabels(points []comparePoint) []string {
	labels := make([]string, len(points))
	for i, point := range points {
		labels[i] = point.Label
	}
	return labels
}

func compareData(points []comparePoint, label string) []byte {
	for _, point := range points {
		if point.Label == label {
			return point.Data
		}
	}
	return nil
}

func formatCompareData(data []byte, mode string) string {
	switch mode {
	case "hex":
		return hex.Dump(data)
	case "base64":
		return base64.StdEncoding.EncodeToString(data)
	default:
		return string(data)
	}
}

func (dg *DeenGUI) showCompare() {
	points := dg.comparePoints()
	labels := compareLabels(points)
	leftSelect := widget.NewSelect(labels, nil)
	rightSelect := widget.NewSelect(labels, nil)
	modeSelect := widget.NewSelect([]string{"text", "hex", "base64"}, nil)
	leftSelect.SetSelected(labels[0])
	rightSelect.SetSelected(labels[len(labels)-1])
	modeSelect.SetSelected("text")

	leftMeta := widget.NewLabel("")
	leftMeta.Importance = widget.LowImportance
	rightMeta := widget.NewLabel("")
	rightMeta.Importance = widget.LowImportance
	leftBody := multilineEntry(12)
	leftBody.Disable()
	rightBody := multilineEntry(12)
	rightBody.Disable()

	refresh := func() {
		mode := modeSelect.Selected
		left := compareData(points, leftSelect.Selected)
		right := compareData(points, rightSelect.Selected)
		leftMeta.SetText(pipeline.DataMetadata(left, 0).Summary())
		rightMeta.SetText(pipeline.DataMetadata(right, 0).Summary())
		leftBody.SetText(formatCompareData(left, mode))
		rightBody.SetText(formatCompareData(right, mode))
	}
	leftSelect.OnChanged = func(string) { refresh() }
	rightSelect.OnChanged = func(string) { refresh() }
	modeSelect.OnChanged = func(string) { refresh() }
	refresh()

	leftPanel := container.NewBorder(container.NewVBox(leftSelect, leftMeta), nil, nil, nil, leftBody)
	rightPanel := container.NewBorder(container.NewVBox(rightSelect, rightMeta), nil, nil, nil, rightBody)
	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.5)
	content := container.NewBorder(container.NewHBox(widget.NewLabel("View"), modeSelect), nil, nil, nil, split)
	content.Resize(fyne.NewSize(820, 520))
	d := dialog.NewCustom("Compare pipeline data", "Close", content, dg.window)
	d.Resize(fyne.NewSize(900, 600))
	d.Show()
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
			dg.pipe.AddStepWithOptions(s.Plugin, s.Unprocess, s.Options)
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
	query.SetPlaceHolder("Search transformers")
	results := container.NewVBox()
	scroll := container.NewVScroll(results)
	scroll.SetMinSize(fyne.NewSize(640, 420))

	var d dialog.Dialog
	refresh := func(q string) {
		results.RemoveAll()
		matches := plugins.SearchUICatalog(q)
		if len(matches) == 0 {
			results.Add(widget.NewLabel("No transformers found."))
		}
		for _, info := range matches {
			info := info
			direction := "encode"
			if !info.CanDecode {
				direction = "run"
			}
			title := fmt.Sprintf("%s / %s", plugins.CategoryLabel(info.Category), info.Label)
			var titleMeta []string
			if info.Label != info.Name {
				titleMeta = append(titleMeta, "command: "+info.Name)
			}
			if len(info.Aliases) > 0 {
				titleMeta = append(titleMeta, "aliases: "+strings.Join(info.Aliases, ", "))
			}
			if len(titleMeta) > 0 {
				title += " (" + strings.Join(titleMeta, "; ") + ")"
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

	d = dialog.NewCustom("Search transformers", "Close", container.NewBorder(query, nil, nil, nil, scroll), dg.window)
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
