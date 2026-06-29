//go:build gui

// Package gui implements the deen desktop interface: a Burp Decoder-style
// chain of plugin transforms backed by the pure internal/pipeline model.
package gui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"io"
	"net/url"
	"sort"
	"strconv"
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

	sourceEntry        *widget.Entry
	sourceHex          *widget.Entry
	sourceStrings      *widget.Entry
	sourceMeta         *widget.Label
	sourceFullControls *fyne.Container
	sourceName         string
	sourceFullRaw      bool
	sourceFullHex      bool
	sourceFullStrings  bool
	stepsBox           *fyne.Container // holds the source card, step cards and add-slot
	cards              []*stepCard     // parallel to pipe.Steps()
	history            *fyne.Container // horizontal transformer-chain overview
	chainView          fyne.CanvasObject
	tabButtons         []*navTab
	tabContent         *fyne.Container
	tabViews           [4]fyne.CanvasObject
	workStatus         *widget.Label
	activeTab          int
	actionsOpen        bool
	stepsExpanded      bool
	working            bool

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
	dg.app.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	dg.window = dg.app.NewWindow("deen")
	dg.window.SetMaster()
	dg.window.SetIcon(deenLogoResource)

	dg.stepsBox = container.NewVBox()
	dg.history = fyne.NewContainerWithLayout(chainRowLayout{})
	dg.activeTab = -1
	dg.actionsOpen = false
	dg.tabContent = container.NewMax()
	dg.workStatus = widget.NewLabel("")
	dg.workStatus.Importance = widget.LowImportance
	dg.workStatus.Hide()
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	top := container.NewVBox(dg.tabHeader(), dg.workStatus)
	content := container.NewBorder(top, nil, nil, nil, dg.tabContent)
	dg.window.SetContent(container.NewStack(bg, content))
	dg.window.SetMainMenu(dg.mainMenu())
	dg.window.Resize(fyne.NewSize(760, 560))

	dg.selectTab(0)
	dg.rebuild()
	return dg, nil
}

// Run shows the window and blocks until it closes.
func (dg *DeenGUI) Run() { dg.window.ShowAndRun() }

func compactMinWidth(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.New(cappedMinWidthLayout{width: compactControlMinWidth}, obj)
}

func (dg *DeenGUI) tabHeader() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(deenLogoResource)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(24, 24))
	title := widget.NewLabelWithStyle("deen", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	brand := container.NewHBox(logo, title, widget.NewSeparator())

	dg.tabButtons = []*navTab{
		newNavTab("Home", func() { dg.selectTab(0) }),
		newNavTab("Examples", func() { dg.selectTab(1) }),
		newNavTab("Plugins", func() { dg.selectTab(2) }),
		newNavTab("About", func() { dg.selectTab(3) }),
	}
	nav := make([]fyne.CanvasObject, 0, len(dg.tabButtons)+1)
	nav = append(nav, brand)
	for _, tab := range dg.tabButtons {
		nav = append(nav, tab)
	}
	return container.NewPadded(container.NewHBox(nav...))
}

func (dg *DeenGUI) homeMenuBar() fyne.CanvasObject {
	menuBar := container.NewHBox(
		dg.menuButton("File", theme.FolderIcon(), fyne.NewMenu("File",
			fyne.NewMenuItemWithIcon("Open file", theme.FolderOpenIcon(), dg.openFile),
			fyne.NewMenuItemWithIcon("Save result", theme.DocumentSaveIcon(), dg.saveResult),
		)),
		dg.menuButton("Navigate", theme.MenuIcon(), fyne.NewMenu("Navigate",
			fyne.NewMenuItemWithIcon("Home", theme.HomeIcon(), func() { dg.selectTab(0) }),
			fyne.NewMenuItemWithIcon("Examples", theme.HistoryIcon(), func() { dg.selectTab(1) }),
			fyne.NewMenuItemWithIcon("Plugins", theme.SearchIcon(), func() { dg.selectTab(2) }),
			fyne.NewMenuItemWithIcon("About", theme.InfoIcon(), func() { dg.selectTab(3) }),
		)),
		dg.menuButton("Chain", theme.FileTextIcon(), fyne.NewMenu("Chain",
			fyne.NewMenuItemWithIcon("Open chain", theme.FileTextIcon(), dg.openChain),
			fyne.NewMenuItemWithIcon("Save chain", theme.DocumentCreateIcon(), dg.saveChain),
			fyne.NewMenuItemWithIcon("Copy command", theme.MailForwardIcon(), dg.copyCommand),
		)),
		dg.menuButton("Workflow", theme.HistoryIcon(), fyne.NewMenu("Workflow",
			fyne.NewMenuItemWithIcon("Presets", theme.HistoryIcon(), dg.showPresets),
			fyne.NewMenuItemWithIcon("Compare", theme.ViewFullScreenIcon(), dg.showCompare),
			fyne.NewMenuItemWithIcon("Undo", theme.NavigateBackIcon(), dg.undo),
			fyne.NewMenuItemWithIcon("Redo", theme.NavigateNextIcon(), dg.redo),
			fyne.NewMenuItemWithIcon("Clear", theme.ContentClearIcon(), dg.clear),
		)),
	)
	return widget.NewCard("", "", container.NewPadded(menuBar))
}

func (dg *DeenGUI) menuButton(label string, icon fyne.Resource, menu *fyne.Menu) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon(label, icon, nil)
	btn.Importance = widget.LowImportance
	btn.OnTapped = func() {
		widget.ShowPopUpMenuAtRelativePosition(menu, dg.window.Canvas(), fyne.NewPos(0, btn.Size().Height), btn)
	}
	return btn
}

func (dg *DeenGUI) selectTab(index int) {
	if index < 0 || index > 3 || dg.activeTab == index || dg.tabContent == nil {
		return
	}
	dg.activeTab = index
	for i, tab := range dg.tabButtons {
		tab.setActive(i == index)
	}

	dg.tabContent.Objects = []fyne.CanvasObject{dg.cachedTab(index)}
	dg.tabContent.Refresh()
}

func (dg *DeenGUI) cachedTab(index int) fyne.CanvasObject {
	if dg.tabViews[index] != nil {
		return dg.tabViews[index]
	}
	var content fyne.CanvasObject
	switch index {
	case 1:
		content = dg.examplesTab()
	case 2:
		content = dg.pluginsTab()
	case 3:
		content = dg.aboutTab()
	default:
		content = dg.homeTab()
	}
	dg.tabViews[index] = compactMinWidth(content)
	return dg.tabViews[index]
}

func (dg *DeenGUI) setWorking(label string, working bool) {
	if dg.workStatus == nil {
		return
	}
	if label == "" {
		label = "Processing"
	}
	if working {
		dg.workStatus.SetText(label + "...")
		dg.workStatus.Show()
	} else {
		dg.workStatus.Hide()
	}
	dg.workStatus.Refresh()
}

func (dg *DeenGUI) runPipelineWork(label string, work func() error, done func()) {
	if dg.working {
		return
	}
	dg.working = true
	dg.setWorking(label, true)
	go func() {
		err := work()
		fyne.Do(func() {
			defer func() {
				dg.working = false
				dg.setWorking("", false)
			}()
			if err != nil {
				dialog.ShowError(err, dg.window)
				return
			}
			if done != nil {
				done()
			}
		})
	}()
}

func (dg *DeenGUI) homeActions() fyne.CanvasObject {
	open := widget.NewButtonWithIcon("Open file", theme.FolderOpenIcon(), dg.openFile)
	open.Importance = widget.HighImportance
	save := widget.NewButtonWithIcon("Save result", theme.DocumentSaveIcon(), dg.saveResult)
	copyResult := widget.NewButtonWithIcon("Copy result", theme.ContentCopyIcon(), dg.copyResult)
	undo := widget.NewButtonWithIcon("Undo", theme.NavigateBackIcon(), dg.undo)
	redo := widget.NewButtonWithIcon("Redo", theme.NavigateNextIcon(), dg.redo)
	clear := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), dg.clear)
	stepLayoutLabel := "Expand steps"
	stepLayoutIcon := theme.MenuExpandIcon()
	if dg.stepsExpanded {
		stepLayoutLabel = "Compact steps"
		stepLayoutIcon = theme.ViewRestoreIcon()
	}
	stepLayout := widget.NewButtonWithIcon(stepLayoutLabel, stepLayoutIcon, func() {
		dg.stepsExpanded = !dg.stepsExpanded
		dg.rebuild()
	})

	openChain := widget.NewButtonWithIcon("Open chain", theme.FileTextIcon(), dg.openChain)
	saveChain := widget.NewButtonWithIcon("Save chain", theme.DocumentCreateIcon(), dg.saveChain)
	presets := widget.NewButtonWithIcon("Presets", theme.HistoryIcon(), dg.showPresets)
	copyCommand := widget.NewButtonWithIcon("Copy command", theme.MailForwardIcon(), dg.copyCommand)

	compare := widget.NewButtonWithIcon("Compare", theme.ViewFullScreenIcon(), dg.showCompare)

	return container.NewVBox(
		actionGroup("Result", copyResult, save, open),
		actionGroup("Chain", openChain, saveChain, copyCommand),
		actionGroup("Workflow", presets, compare, stepLayout, undo, redo, clear),
	)
}

func actionGroup(title string, objects ...fyne.CanvasObject) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Importance = widget.LowImportance
	return container.NewVBox(label, container.NewGridWithColumns(3, objects...))
}

type chainRowLayout struct{}

func (chainRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}
		min := obj.MinSize()
		y := float32(0)
		if size.Height > min.Height {
			y = (size.Height - min.Height) / 2
		}
		obj.Move(fyne.NewPos(x, y))
		obj.Resize(min)
		x += min.Width + theme.Padding()
	}
}

func (chainRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width, height := float32(0), float32(0)
	visible := 0
	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}
		min := obj.MinSize()
		width += min.Width
		if min.Height > height {
			height = min.Height
		}
		visible++
	}
	if visible > 1 {
		width += theme.Padding() * float32(visible-1)
	}
	return fyne.NewSize(width, height)
}

func (dg *DeenGUI) newActionBar() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	toggleLabel := "Hide"
	toggleIcon := theme.MenuDropUpIcon()
	if !dg.actionsOpen {
		toggleLabel = "Show"
		toggleIcon = theme.MenuDropDownIcon()
	}
	toggle := widget.NewButtonWithIcon(toggleLabel, toggleIcon, func() {
		dg.actionsOpen = !dg.actionsOpen
		dg.rebuild()
	})
	toggle.Importance = widget.LowImportance
	header := container.NewBorder(nil, nil, title, toggle)
	if !dg.actionsOpen {
		return widget.NewCard("", "", header)
	}
	return widget.NewCard("", "", container.NewVBox(header, dg.homeActions()))
}

// mainMenu builds the window menu (theme switching).
func (dg *DeenGUI) mainMenu() *fyne.MainMenu {
	setTheme := func(t fyne.Theme) func() { return func() { dg.app.Settings().SetTheme(t) } }
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItemWithIcon("Open file", theme.FolderOpenIcon(), dg.openFile),
		fyne.NewMenuItemWithIcon("Save result", theme.DocumentSaveIcon(), dg.saveResult),
	)
	navigateMenu := fyne.NewMenu("Navigate",
		fyne.NewMenuItemWithIcon("Home", theme.HomeIcon(), func() { dg.selectTab(0) }),
		fyne.NewMenuItemWithIcon("Examples", theme.HistoryIcon(), func() { dg.selectTab(1) }),
		fyne.NewMenuItemWithIcon("Plugins", theme.SearchIcon(), func() { dg.selectTab(2) }),
		fyne.NewMenuItemWithIcon("About", theme.InfoIcon(), func() { dg.selectTab(3) }),
	)
	chainMenu := fyne.NewMenu("Chain",
		fyne.NewMenuItemWithIcon("Open chain", theme.FileTextIcon(), dg.openChain),
		fyne.NewMenuItemWithIcon("Save chain", theme.DocumentCreateIcon(), dg.saveChain),
		fyne.NewMenuItemWithIcon("Copy command", theme.MailForwardIcon(), dg.copyCommand),
	)
	workflowMenu := fyne.NewMenu("Workflow",
		fyne.NewMenuItemWithIcon("Presets", theme.HistoryIcon(), dg.showPresets),
		fyne.NewMenuItemWithIcon("Compare", theme.ViewFullScreenIcon(), dg.showCompare),
		fyne.NewMenuItemWithIcon("Undo", theme.NavigateBackIcon(), dg.undo),
		fyne.NewMenuItemWithIcon("Redo", theme.NavigateNextIcon(), dg.redo),
		fyne.NewMenuItemWithIcon("Clear", theme.ContentClearIcon(), dg.clear),
	)
	themeMenu := fyne.NewMenu("Theme",
		fyne.NewMenuItemWithIcon("Dark", theme.VisibilityIcon(), setTheme(newAdversecTheme(theme.VariantDark))),
		fyne.NewMenuItemWithIcon("Light", theme.VisibilityOffIcon(), setTheme(newAdversecTheme(theme.VariantLight))),
		fyne.NewMenuItemWithIcon("System", theme.SettingsIcon(), setTheme(theme.DefaultTheme())),
	)
	help := fyne.NewMenu("Help",
		fyne.NewMenuItemWithIcon("How to use", theme.HelpIcon(), dg.showHelp),
		fyne.NewMenuItemWithIcon("Examples", theme.HistoryIcon(), func() { dg.selectTab(1) }),
		fyne.NewMenuItemWithIcon("Plugin catalog", theme.SearchIcon(), func() { dg.selectTab(2) }),
		fyne.NewMenuItemWithIcon("About", theme.InfoIcon(), func() { dg.selectTab(3) }),
	)
	return fyne.NewMainMenu(fileMenu, navigateMenu, chainMenu, workflowMenu, themeMenu, help)
}

func (dg *DeenGUI) homeTab() fyne.CanvasObject {
	return container.NewVScroll(dg.stepsBox)
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
	dg.selectTab(3)
}

func (dg *DeenGUI) aboutTab() fyne.CanvasObject {
	version := core.Version()
	if b := core.Branch(); b != "" {
		version += " (" + b + ")"
	}
	docsURL, _ := url.Parse("https://deen.adversec.com")
	repoURL, _ := url.Parse("https://github.com/takeshixx/go-deen")

	content := container.NewVBox(
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

	load := widget.NewButtonWithIcon("Load example", theme.MediaPlayIcon(), func() {
		dg.runPipelineWork("Loading example", func() error {
			dg.pipe.ApplyExample(example)
			dg.stepsExpanded = false
			return nil
		}, func() {
			dg.rebuild()
			dg.selectTab(0)
		})
	})
	previewSlot := container.NewVBox()
	var preview *widget.Button
	preview = widget.NewButtonWithIcon("Preview data", theme.VisibilityIcon(), func() {
		previewSlot.RemoveAll()
		result, err := pipeline.ExampleResult(example)
		if err != nil {
			errLabel := widget.NewLabel("Output error: " + err.Error())
			errLabel.Importance = widget.DangerImportance
			errLabel.Wrapping = fyne.TextWrapBreak
			previewSlot.Add(errLabel)
			previewSlot.Refresh()
			return
		}
		outputSummary := widget.NewLabel("Output: " + pipeline.DataMetadata(result, len(example.Source)).Summary())
		outputSummary.Importance = widget.LowImportance
		outputSummary.Wrapping = fyne.TextWrapBreak
		inputEntry := multilineEntry(5)
		inputEntry.SetText(exampleDataText(example.Source))
		inputEntry.Disable()
		outputEntry := multilineEntry(5)
		outputEntry.SetText(exampleDataText(result))
		outputEntry.Disable()
		dataGrid := container.NewGridWithColumns(2,
			widget.NewCard("Input data", "", exampleDataObject(example.Source, inputEntry)),
			widget.NewCard("Output result", "", exampleDataObject(result, outputEntry)),
		)
		preview.Disable()
		previewSlot.Add(outputSummary)
		previewSlot.Add(dataGrid)
		previewSlot.Refresh()
	})

	body := container.NewVBox(desc, source, chain)
	if example.WantContains != "" {
		want := widget.NewLabel("Expected result contains: " + example.WantContains)
		want.Importance = widget.LowImportance
		want.Wrapping = fyne.TextWrapBreak
		body.Add(want)
	}
	body.Add(load)
	body.Add(preview)
	body.Add(previewSlot)
	details := container.New(cappedMinWidthLayout{width: compactControlMinWidth}, body)
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

func exampleDataObject(data []byte, fallback fyne.CanvasObject) fyne.CanvasObject {
	img := canvas.NewImageFromReader(bytes.NewReader(data), "example-output")
	if img == nil || img.Image == nil {
		if preview, spans, ok := pipeline.HighlightedPreview(data); ok {
			grid := newPreviewGrid()
			setPreviewText(grid, preview, spans)
			return grid
		}
		return fallback
	}
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(180, 180))
	return container.NewCenter(img)
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
	dg.setWorking("Rendering pipeline", true)
	defer dg.setWorking("", false)
	dg.stepsBox.RemoveAll()
	dg.cards = dg.cards[:0]

	dg.stepsBox.Add(dg.homeMenuBar())
	dg.stepsBox.Add(dg.newSourceCard())
	dg.stepsBox.Add(dg.newChainOverview())
	for i := range dg.pipe.Steps() {
		c := dg.newStepCard(i)
		dg.cards = append(dg.cards, c)
		dg.stepsBox.Add(c.container)
	}
	dg.stepsBox.Add(dg.newAddSlot())
	dg.stepsBox.Refresh()
	dg.updateHistory()
}

func (dg *DeenGUI) newChainOverview() fyne.CanvasObject {
	scroll := container.NewHScroll(dg.history)
	scroll.SetMinSize(fyne.NewSize(compactControlMinWidth, 54))
	dg.chainView = scroll
	return container.NewMax(scroll)
}

// updateHistory redraws the horizontal transformer chain overview.
func (dg *DeenGUI) updateHistory() {
	dg.history.RemoveAll()
	shown := 0
	for i, s := range dg.pipe.Steps() {
		if s.Disabled {
			continue
		}
		if shown > 0 {
			arrow := canvas.NewText("→", theme.Color(theme.ColorNamePlaceHolder))
			arrow.TextStyle = fyne.TextStyle{Bold: true}
			arrow.TextSize = 18
			dg.history.Add(arrow)
		}
		dg.history.Add(guiChainStepPill(i, s))
		shown++
	}
	if dg.chainView != nil {
		if shown == 0 {
			dg.chainView.Hide()
		} else {
			dg.chainView.Show()
		}
	}
	dg.history.Refresh()
}

func guiChainStepPill(i int, step *pipeline.Step) fyne.CanvasObject {
	displayCol := accent(i)
	if step.Disabled {
		displayCol = disabledAccent()
	}
	name := plugins.PluginLabel(step.Plugin)
	if name == "" {
		name = "(none)"
	}
	if step.Unprocess {
		name = "." + name
	}
	title := canvas.NewText(name, displayCol)
	title.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	metaParts := make([]string, 0, len(step.Options))
	for k, v := range step.Options {
		metaParts = append(metaParts, k+"="+v)
	}
	sort.Strings(metaParts)
	var meta fyne.CanvasObject
	if len(metaParts) > 0 {
		text := canvas.NewText(strings.Join(metaParts, ", "), theme.Color(theme.ColorNamePlaceHolder))
		text.TextStyle = fyne.TextStyle{Monospace: true}
		text.TextSize = 12
		meta = text
	}
	return guiChainPill(title, meta, displayCol, step.Disabled)
}

func guiChainPill(title fyne.CanvasObject, meta fyne.CanvasObject, accentColor color.NRGBA, disabled bool) fyne.CanvasObject {
	bg := canvas.NewRectangle(tint(accentColor))
	if disabled {
		bg.FillColor = color.NRGBA{R: accentColor.R, G: accentColor.G, B: accentColor.B, A: 0x18}
	}
	bg.StrokeColor = accentColor
	bg.StrokeWidth = 1
	bg.CornerRadius = 6
	body := container.NewVBox(title)
	if meta != nil {
		body.Add(meta)
	}
	return container.NewStack(bg, container.NewPadded(body))
}

// refreshFrom updates the displayed output of every card from index `from`
// downward without recreating widgets.
func (dg *DeenGUI) refreshFrom(from int) {
	dg.setWorking("Refreshing output", true)
	defer dg.setWorking("", false)
	if dg.sourceMeta != nil {
		dg.sourceMeta.SetText(dg.sourceMetadataSummary())
	}
	rawNeedsFull, hexNeedsFull, stringsNeedsFull := dg.sourceNeedsFull()
	if !rawNeedsFull {
		dg.sourceFullRaw = false
	}
	if !hexNeedsFull {
		dg.sourceFullHex = false
	}
	if !stringsNeedsFull {
		dg.sourceFullStrings = false
	}
	if dg.sourceEntry != nil {
		text, textCapped := guiTextDisplayMode(dg.pipe.Source(), dg.sourceFullRaw)
		dg.setText(dg.sourceEntry, text)
		if textCapped {
			dg.sourceEntry.Disable()
		} else {
			dg.sourceEntry.Enable()
		}
	}
	if dg.sourceHex != nil {
		hexText, _ := guiHexDisplayMode(dg.pipe.Source(), dg.sourceFullHex)
		dg.setText(dg.sourceHex, hexText)
	}
	if dg.sourceStrings != nil {
		stringsText, _ := guiStringsDisplayMode(dg.pipe.Source(), dg.sourceFullStrings)
		dg.setText(dg.sourceStrings, stringsText)
	}
	dg.refreshSourceFullControls(rawNeedsFull, hexNeedsFull, stringsNeedsFull)
	for i := from; i < len(dg.cards); i++ {
		if dg.cards[i].collapsed {
			continue
		}
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
		name := rc.URI().Name()
		dg.sourceName = name
		dg.clearSourceFullViews()
		dg.runPipelineWork("Processing file", func() error {
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}
			dg.pipe.SetSourceOwned(data)
			return nil
		}, func() {
			dg.rebuild()
		})
	}, dg.window)
}

func (dg *DeenGUI) sourceMetadataSummary() string {
	return metadataSummary(dg.sourceName, pipeline.DataMetadata(dg.pipe.Source(), 0))
}

func metadataSummary(source string, meta pipeline.Metadata) string {
	var lines []string
	if strings.TrimSpace(source) != "" {
		lines = append(lines, "source: "+source)
	}
	for _, field := range meta.Fields() {
		lines = append(lines, field.Label+": "+field.Value)
	}
	return strings.Join(lines, "\n")
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
		dg.runPipelineWork("Importing chain", func() error {
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}
			dg.clearSourceFullViews()
			if err := dg.pipe.ImportJSON(data); err != nil {
				return err
			}
			dg.stepsExpanded = false
			return nil
		}, dg.rebuild)
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
	if pipeline.IsLargeData(dg.pipe.Result()) {
		dialog.ShowInformation("Copy result", "Result is too large to copy safely from the GUI. Use Save result instead.", dg.window)
		return
	}
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
		text, _ := guiHexDisplay(data)
		return text
	case "base64":
		if pipeline.IsLargeData(data) {
			return pipeline.LargeDataPlaceholder(data) + "\n\nBase64 preview disabled for large data."
		}
		return base64.StdEncoding.EncodeToString(data)
	default:
		text, _ := guiTextDisplay(data)
		return text
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
		leftMeta.SetText(metadataSummary("", pipeline.DataMetadata(left, 0)))
		rightMeta.SetText(metadataSummary("", pipeline.DataMetadata(right, 0)))
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
	var suggestions []pipeline.Suggestion
	dg.runPipelineWork("Detecting transforms", func() error {
		suggestions = pipeline.Suggestions(dg.pipe.Result())
		return nil
	}, func() {
		dg.showSuggestionsDialog(suggestions)
	})
}

func (dg *DeenGUI) showSuggestionsDialog(suggestions []pipeline.Suggestion) {
	list := container.NewVBox()
	if len(suggestions) == 0 {
		list.Add(widget.NewLabel("No likely transforms detected."))
		dialog.ShowCustom("Suggested transforms", "Close", list, dg.window)
		return
	}
	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	var d dialog.Dialog
	for _, s := range suggestions {
		s := s
		detail := s.Reason
		if s.Confidence > 0 {
			detail += fmt.Sprintf(" Confidence: %d%%.", s.Confidence)
		}
		actionLabel := "Add"
		if len(s.Steps) > 1 {
			actionLabel = "Apply chain"
		}
		action := widget.NewButton(actionLabel, func() {
			if d != nil {
				d.Hide()
			}
			dg.runPipelineWork("Processing", func() error {
				dg.pipe.AddSuggestion(s)
				return nil
			}, dg.rebuild)
		})
		itemContent := container.NewVBox(widget.NewLabelWithStyle(s.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		if s.Reason != "" {
			itemContent.Add(widget.NewLabel(detail))
		}
		if s.Preview != "" {
			itemContent.Add(widget.NewLabel(guiSafeSuggestionPreview(s.Preview)))
		}
		list.Add(container.NewBorder(nil, nil, nil, action, itemContent))
	}
	d = dialog.NewCustom("Suggested transforms", "Close", container.NewVScroll(list), dg.window)
	d.Resize(fyne.NewSize(560, 360))
	d.Show()
}

func guiSafeSuggestionPreview(preview string) string {
	var b strings.Builder
	for _, r := range preview {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteRune(r)
		case strconv.IsPrint(r):
			b.WriteRune(r)
		case r <= 0xff:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r <= 0xffff:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	return b.String()
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
				if d != nil {
					d.Hide()
				}
				dg.runPipelineWork("Processing", func() error {
					dg.pipe.AddStep(info.Name, false)
					return nil
				}, dg.rebuild)
			})
			actions := container.NewHBox(addEncode)
			if info.CanDecode {
				actions.Add(widget.NewButton("Add decode", func() {
					if d != nil {
						d.Hide()
					}
					dg.runPipelineWork("Processing", func() error {
						dg.pipe.AddStep(info.Name, true)
						return nil
					}, dg.rebuild)
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
			if d != nil {
				d.Hide()
			}
			dg.runPipelineWork("Processing", func() error {
				dg.pipe.ApplyPreset(preset)
				return nil
			}, dg.rebuild)
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
	dg.runPipelineWork("Undoing", func() error {
		dg.pipe.Undo()
		return nil
	}, dg.rebuild)
}

func (dg *DeenGUI) redo() {
	dg.runPipelineWork("Redoing", func() error {
		dg.pipe.Redo()
		return nil
	}, dg.rebuild)
}

func (dg *DeenGUI) clear() {
	dg.sourceName = ""
	dg.runPipelineWork("Clearing", func() error {
		dg.pipe.Clear()
		return nil
	}, dg.rebuild)
}
