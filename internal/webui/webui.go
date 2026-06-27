//go:build js && wasm

// Package webui renders the deen browser interface: a Burp Decoder-style chain
// of plugin transforms, mirroring the desktop GUI, built on the shared
// internal/pipeline model via syscall/js.
package webui

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"syscall/js"
	"unicode/utf8"

	"github.com/takeshixx/deen/internal/core"
	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

var palette = []string{"#4285f4", "#0f9d58", "#f4b400", "#db4437", "#ab47bc", "#00acc1"}

func accent(i int) string { return palette[i%len(palette)] }

var (
	doc          js.Value
	pipe         *pipeline.Pipeline
	contentEl    js.Value
	tabBtns      []js.Value
	stepsEl      js.Value
	historyEl    js.Value
	sourceEl     js.Value
	updating     bool
	callbacks    []js.Func
	staticCBs    []js.Func
	cards        []*cardRef
	activeTab    = "home"
	sourceName   string
	navCollapsed bool
)

type cardRef struct {
	index        int
	output       js.Value
	hexOutput    js.Value
	preview      js.Value
	imageEnabled bool
	imagePanel   js.Value
	image        js.Value
	imageMsg     js.Value
	meta         js.Value
	errEl        js.Value
}

// Run builds the UI and blocks so the event callbacks stay alive.
func Run() {
	doc = js.Global().Get("document")
	pipe = pipeline.New()
	loadChainFromHash()
	buildLayout()
	select {}
}

// --- small DOM helpers ---

func el(tag string) js.Value { return doc.Call("createElement", tag) }

func svgEl(tag string) js.Value {
	return doc.Call("createElementNS", "http://www.w3.org/2000/svg", tag)
}

func div(class string) js.Value {
	d := el("div")
	d.Set("className", class)
	return d
}

func textNode(s string) js.Value { return doc.Call("createTextNode", s) }

func appendChildren(parent js.Value, kids ...js.Value) {
	for _, k := range kids {
		parent.Call("appendChild", k)
	}
}

func on(elem js.Value, event string, fn func()) {
	cb := js.FuncOf(func(js.Value, []js.Value) any { fn(); return nil })
	callbacks = append(callbacks, cb)
	elem.Call("addEventListener", event, cb)
}

func onEvent(elem js.Value, event string, fn func(js.Value)) {
	cb := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			fn(args[0])
		}
		return nil
	})
	callbacks = append(callbacks, cb)
	elem.Call("addEventListener", event, cb)
}

func onStatic(elem js.Value, event string, fn func()) {
	cb := js.FuncOf(func(js.Value, []js.Value) any { fn(); return nil })
	staticCBs = append(staticCBs, cb)
	elem.Call("addEventListener", event, cb)
}

func releaseCallbacks() {
	for _, cb := range callbacks {
		cb.Release()
	}
	callbacks = callbacks[:0]
}

func textarea(value string) js.Value {
	return textareaWithMax(value, 512)
}

func textareaWithMax(value string, maxHeight int) js.Value {
	t := el("textarea")
	t.Set("className", "io")
	t.Set("value", value)
	t.Call("setAttribute", "data-max-height", strconv.Itoa(maxHeight))
	autoSizeTextarea(t)
	on(t, "input", func() { autoSizeTextarea(t) })
	return t
}

func autoSizeTextarea(t js.Value) {
	t.Get("style").Set("height", "auto")
	maxHeight := 512
	if attr := t.Call("getAttribute", "data-max-height"); attr.Truthy() {
		if n, err := strconv.Atoi(attr.String()); err == nil && n > 0 {
			maxHeight = n
		}
	}
	scrollHeight := t.Get("scrollHeight").Int() + 2
	if scrollHeight <= 2 {
		return
	}
	height := scrollHeight
	if height > maxHeight {
		height = maxHeight
		t.Get("style").Set("overflowY", "auto")
	} else {
		t.Get("style").Set("overflowY", "hidden")
	}
	t.Get("style").Set("height", fmt.Sprintf("%dpx", height))
}

func autoSizeTextareaSoon(t js.Value) {
	autoSizeTextarea(t)
	raf := js.Global().Get("requestAnimationFrame")
	if !raf.Truthy() {
		return
	}
	var cb js.Func
	cb = js.FuncOf(func(js.Value, []js.Value) any {
		autoSizeTextarea(t)
		cb.Release()
		return nil
	})
	raf.Invoke(cb)
}

func autoSizeOutputTextareas(c *cardRef) {
	autoSizeTextareaSoon(c.output)
	autoSizeTextareaSoon(c.hexOutput)
}

func previewBox() js.Value {
	p := el("pre")
	p.Set("className", "syntax-preview")
	return p
}

func selectEl(placeholder string, options []string, selected string) js.Value {
	s := el("select")
	opt0 := el("option")
	opt0.Set("value", "")
	opt0.Set("textContent", placeholder)
	s.Call("appendChild", opt0)
	for _, o := range options {
		op := el("option")
		op.Set("value", o)
		op.Set("textContent", o)
		if o == selected {
			op.Set("selected", true)
		}
		s.Call("appendChild", op)
	}
	return s
}

type selectOption struct {
	value string
	label string
}

func selectOptionsEl(placeholder string, options []selectOption, selected string) js.Value {
	s := el("select")
	opt0 := el("option")
	opt0.Set("value", "")
	opt0.Set("textContent", placeholder)
	s.Call("appendChild", opt0)
	for _, o := range options {
		op := el("option")
		op.Set("value", o.value)
		op.Set("textContent", o.label)
		if o.value == selected {
			op.Set("selected", true)
		}
		s.Call("appendChild", op)
	}
	return s
}

func pluginSelectOptions(category string) []selectOption {
	names := plugins.InCategory(category)
	opts := make([]selectOption, 0, len(names))
	for _, name := range names {
		opts = append(opts, selectOption{value: name, label: plugins.PluginLabel(name)})
	}
	sort.Slice(opts, func(i, j int) bool {
		return strings.ToLower(opts[i].label) < strings.ToLower(opts[j].label)
	})
	return opts
}

func checkbox(label string, checked bool) (wrap, input js.Value) {
	wrap = el("label")
	input = el("input")
	input.Set("type", "checkbox")
	input.Set("checked", checked)
	appendChildren(wrap, input, textNode(" "+label))
	return wrap, input
}

func button(className, label string, fn func()) js.Value {
	b := el("button")
	b.Set("className", className)
	b.Set("type", "button")
	b.Set("textContent", label)
	on(b, "click", fn)
	return b
}

func iconButton(className, icon, label string, fn func()) js.Value {
	b := el("button")
	b.Set("className", className)
	b.Set("type", "button")
	b.Set("aria-label", label)
	setIconButtonContent(b, icon, label)
	on(b, "click", fn)
	return b
}

func setIconButtonContent(b js.Value, icon, label string) {
	className := b.Get("className").String()
	if !strings.Contains(" "+className+" ", " icon-label ") {
		b.Set("className", strings.TrimSpace(className+" icon-label"))
	}
	b.Set("aria-label", label)
	text := el("span")
	text.Set("textContent", label)
	appendChildren(b, iconGraphic(icon), text)
}

func iconGraphic(name string) js.Value {
	svg := svgEl("svg")
	svg.Call("setAttribute", "class", "button-icon")
	svg.Call("setAttribute", "width", "16")
	svg.Call("setAttribute", "height", "16")
	svg.Call("setAttribute", "viewBox", "0 0 24 24")
	svg.Call("setAttribute", "aria-hidden", "true")
	svg.Call("setAttribute", "focusable", "false")
	svg.Call("setAttribute", "fill", "none")
	svg.Call("setAttribute", "stroke", "currentColor")
	svg.Call("setAttribute", "stroke-width", "2")
	svg.Call("setAttribute", "stroke-linecap", "round")
	svg.Call("setAttribute", "stroke-linejoin", "round")
	for _, d := range iconPaths[name] {
		path := svgEl("path")
		path.Call("setAttribute", "d", d)
		svg.Call("appendChild", path)
	}
	return svg
}

var iconPaths = map[string][]string{
	"copy":     {"M8 8h11v11H8z", "M5 16H4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1"},
	"download": {"M12 3v12", "M7 10l5 5 5-5", "M5 21h14"},
	"upload":   {"M12 21V9", "M7 14l5-5 5 5", "M5 3h14"},
	"link":     {"M10 13a5 5 0 0 0 7.1 0l2-2a5 5 0 0 0-7.1-7.1l-1.2 1.2", "M14 11a5 5 0 0 0-7.1 0l-2 2A5 5 0 0 0 12 20.1l1.2-1.2"},
	"terminal": {"M4 17l6-5-6-5", "M12 19h8"},
	"star":     {"M12 3l2.7 5.5 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.8 1-6.1-4.4-4.3 6.1-.9z"},
	"compare":  {"M8 6h13", "M8 12h13", "M8 18h13", "M3 6h.01", "M3 12h.01", "M3 18h.01"},
	"undo":     {"M9 14l-5-5 5-5", "M4 9h10a6 6 0 0 1 0 12h-1"},
	"redo":     {"M15 14l5-5-5-5", "M20 9H10a6 6 0 0 0 0 12h1"},
	"clear":    {"M18 6L6 18", "M6 6l12 12"},
	"collapse": {"M15 18l-6-6 6-6"},
	"expand":   {"M9 18l6-6-6-6"},
}

func toolbarGroup(kids ...js.Value) js.Value {
	group := div("toolbar-group")
	appendChildren(group, kids...)
	return group
}

func textInput(placeholder, value string) js.Value {
	input := el("input")
	input.Set("type", "text")
	input.Set("placeholder", placeholder)
	input.Set("value", value)
	return input
}

func showModal(title string, content js.Value) func() {
	overlay := div("modal-overlay")
	modal := div("modal")
	header := div("modal-header")
	h := el("h2")
	h.Set("textContent", title)
	closeBtn := el("button")
	closeBtn.Set("type", "button")
	closeBtn.Set("className", "icon")
	closeBtn.Set("textContent", "✕")
	body := div("modal-body")

	appendChildren(header, h, closeBtn)
	appendChildren(body, content)
	appendChildren(modal, header, body)
	overlay.Call("appendChild", modal)
	doc.Get("body").Call("appendChild", overlay)

	close := func() {
		if overlay.Truthy() {
			overlay.Call("remove")
		}
	}
	on(closeBtn, "click", close)
	return close
}

// --- layout ---

func buildLayout() {
	body := doc.Get("body")
	body.Set("innerHTML", "")

	shell := div("shell")
	header := div("app-header")
	brand := div("brand")
	logo := el("img")
	logo.Set("className", "brand-logo")
	logo.Set("src", "favicon.svg")
	logo.Set("alt", "")
	logo.Set("aria-hidden", "true")
	heading := el("h1")
	heading.Set("textContent", "deen")
	appendChildren(brand, logo, heading)
	tabs := div("tabs")
	appendChildren(tabs, tabButton("home", "Home"), tabButton("examples", "Examples"), tabButton("plugins", "Plugins"), tabButton("about", "About"))
	appendChildren(header, brand, tabs)

	contentEl = div("tab-content")
	appendChildren(shell, header, contentEl)
	body.Call("appendChild", shell)

	renderActiveTab()
}

func tabButton(tab, label string) js.Value {
	b := el("button")
	b.Set("type", "button")
	b.Set("className", "tab")
	b.Set("textContent", label)
	tabBtns = append(tabBtns, b)
	onStatic(b, "click", func() {
		activeTab = tab
		renderActiveTab()
	})
	return b
}

func renderActiveTab() {
	releaseCallbacks()
	contentEl.Set("innerHTML", "")
	cards = cards[:0]
	for _, b := range tabBtns {
		if strings.EqualFold(b.Get("textContent").String(), activeTab) {
			b.Set("className", "tab active")
		} else {
			b.Set("className", "tab")
		}
	}
	switch activeTab {
	case "examples":
		renderExamples()
	case "plugins":
		renderPlugins()
	case "about":
		renderAbout()
	default:
		renderHome()
	}
}

// rebuild reconstructs the source card, step cards and add card.
func rebuild() {
	releaseCallbacks()
	contentEl.Set("innerHTML", "")
	cards = cards[:0]
	renderHome()
}

func renderHome() {
	app := div("app")
	if navCollapsed {
		app.Set("className", "app nav-collapsed")
	}
	main := div("main")
	side := div("side")

	stepsEl = div("steps")

	stepsEl.Call("appendChild", sourceCard())
	for i := range pipe.Steps() {
		stepsEl.Call("appendChild", stepCard(i))
	}
	stepsEl.Call("appendChild", addCard())

	sideHeader := div("side-header")
	sideTitle := el("h2")
	sideTitle.Set("textContent", "Navigation")
	toggleLabel := "Collapse navigation"
	toggleIcon := "collapse"
	if navCollapsed {
		toggleLabel = "Expand navigation"
		toggleIcon = "expand"
	}
	toggleNav := iconButton("nav-toggle", toggleIcon, toggleLabel, func() {
		navCollapsed = !navCollapsed
		rebuild()
	})
	appendChildren(sideHeader, sideTitle, toggleNav)
	chainTitle := el("h2")
	chainTitle.Set("className", "chain-heading")
	chainTitle.Set("textContent", "Transformer Chain")
	historyEl = div("history")
	sideActions := div("side-actions")
	appendChildren(sideActions,
		sideActionGroup("Result",
			iconButton("primary", "copy", "Copy result", copyResult),
			iconButton("", "download", "Download", downloadResult),
			filePicker(),
		),
		sideActionGroup("Chain",
			chainPicker(),
			iconButton("", "upload", "Export chain", exportChain),
			iconButton("", "link", "Copy link", copyShareLink),
			iconButton("", "terminal", "Copy command", copyCommand),
		),
		sideActionGroup("Workflow",
			iconButton("", "star", "Presets", showPresets),
			iconButton("", "compare", "Compare", showCompare),
			iconButton("", "undo", "Undo", func() {
				if pipe.Undo() {
					rebuild()
				}
			}),
			iconButton("", "redo", "Redo", func() {
				if pipe.Redo() {
					rebuild()
				}
			}),
			iconButton("", "clear", "Clear", func() {
				sourceName = ""
				pipe.Clear()
				rebuild()
			}),
		),
	)

	appendChildren(main, stepsEl)
	appendChildren(side, sideHeader, sideActions, chainTitle, historyEl)
	appendChildren(app, main, side)
	contentEl.Call("appendChild", app)
	updateHistory()
}

func sideActionGroup(title string, kids ...js.Value) js.Value {
	group := div("side-action-group")
	label := el("div")
	label.Set("className", "side-action-title")
	label.Set("textContent", title)
	appendChildren(group, label)
	for _, kid := range kids {
		group.Call("appendChild", kid)
	}
	return group
}

func renderAbout() {
	version := core.Version()
	if b := core.Branch(); b != "" {
		version += " (" + b + ")"
	}
	page := div("info-page")
	h := el("h2")
	h.Set("textContent", "deen")
	desc := el("p")
	desc.Set("textContent", "deen encodes, decodes, hashes, compresses and formats data through a configurable chain of plugins.")
	versionEl := el("p")
	versionEl.Set("textContent", "Version: "+version)
	built := el("p")
	built.Set("textContent", "Built with Go, Fyne for desktop, and WebAssembly for this browser interface. Processing runs locally in the current app surface.")
	docs := link("Documentation", "https://deen.adversec.com")
	repo := link("Source", "https://github.com/takeshixx/go-deen")
	appendChildren(page, h, desc, versionEl, built, docs, repo)
	contentEl.Call("appendChild", page)
}

func renderExamples() {
	page := div("catalog examples-page")
	header := div("examples-header")
	title := el("h2")
	title.Set("textContent", "Examples")
	query := textInput("Search examples", "")
	appendChildren(header, title, query)
	results := div("examples-grid")
	appendChildren(page, header, results)

	render := func() {
		results.Set("innerHTML", "")
		matches := 0
		for _, example := range pipeline.BuiltinExamples() {
			if !pipeline.ExampleMatches(example, query.Get("value").String()) {
				continue
			}
			matches++
			results.Call("appendChild", exampleCard(example))
		}
		if matches == 0 {
			empty := div("empty")
			empty.Set("textContent", "No examples found.")
			results.Call("appendChild", empty)
		}
	}
	on(query, "input", render)
	render()
	contentEl.Call("appendChild", page)
}

func exampleCard(example pipeline.Example) js.Value {
	card := el("details")
	card.Set("className", "plugin-card example-card")
	summary := el("summary")
	title := el("h3")
	title.Set("textContent", example.Name)
	summary.Call("appendChild", title)

	desc := el("p")
	desc.Set("textContent", example.Description)
	result, err := pipeline.ExampleResult(example)
	chain := exampleChain(example.Steps)

	data := div("example-data-grid")
	inputPanel := exampleDataPanel("Input data", example.Source)
	outputPanel := exampleDataPanel("Output result", result)
	appendChildren(data, inputPanel, outputPanel)

	actions := div("modal-actions")
	load := el("button")
	load.Set("type", "button")
	load.Set("textContent", "Load example")
	on(load, "click", func() {
		sourceName = ""
		pipe.ApplyExample(example)
		activeTab = "home"
		renderActiveTab()
	})
	actions.Call("appendChild", load)
	appendChildren(card, summary, desc, chain)
	if err != nil {
		output := div("error")
		output.Set("textContent", "Output error: "+err.Error())
		card.Call("appendChild", output)
	}
	if example.WantContains != "" {
		want := div("plugin-meta")
		want.Set("textContent", "Expected result contains: "+example.WantContains)
		card.Call("appendChild", want)
	}
	card.Call("appendChild", data)
	card.Call("appendChild", actions)
	return card
}

func exampleDataPanel(title string, data []byte) js.Value {
	panel := div("example-data-panel")
	label := el("strong")
	label.Set("textContent", title)
	body := el("pre")
	body.Set("className", "syntax-preview example-data-code")
	text := exampleDataText(data)
	renderHighlightedText(body, text, exampleSyntaxSpans(text))
	appendChildren(panel, label, body)
	return panel
}

func exampleSyntaxSpans(text string) []pipeline.SyntaxSpan {
	if !json.Valid([]byte(strings.TrimSpace(text))) {
		return nil
	}
	return pipeline.JSONSyntaxSpans(text)
}

func exampleChain(steps []pipeline.PresetStep) js.Value {
	wrap := div("example-chain")
	label := el("span")
	label.Set("className", "example-chain-label")
	label.Set("textContent", "Transformer chain")
	wrap.Call("appendChild", label)
	for i, step := range steps {
		if i > 0 {
			arrow := el("span")
			arrow.Set("className", "chain-arrow")
			arrow.Set("textContent", "->")
			wrap.Call("appendChild", arrow)
		}
		wrap.Call("appendChild", exampleStepPill(step))
	}
	return wrap
}

func exampleStepPill(step pipeline.PresetStep) js.Value {
	pill := div("chain-pill")
	name := plugins.PluginLabel(step.Plugin)
	if step.Unprocess {
		name = "." + name
	}
	title := el("span")
	title.Set("className", "chain-plugin")
	title.Set("textContent", name)
	pill.Call("appendChild", title)
	if len(step.Options) > 0 {
		opts := make([]string, 0, len(step.Options))
		for k, v := range step.Options {
			opts = append(opts, k+"="+v)
		}
		sort.Strings(opts)
		meta := el("span")
		meta.Set("className", "chain-options")
		meta.Set("textContent", strings.Join(opts, ", "))
		pill.Call("appendChild", meta)
	}
	return pill
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
	return strings.Join(parts, " -> ")
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

func renderPlugins() {
	page := div("catalog")
	current := ""
	for _, info := range plugins.UICatalog() {
		if info.Category != current {
			current = info.Category
			h := el("h2")
			h.Set("textContent", plugins.CategoryLabel(current))
			page.Call("appendChild", h)
		}
		page.Call("appendChild", pluginCard(info))
	}
	contentEl.Call("appendChild", page)
}

func link(label, href string) js.Value {
	a := el("a")
	a.Set("href", href)
	a.Set("textContent", label)
	a.Set("target", "_blank")
	a.Set("rel", "noopener noreferrer")
	return a
}

func pluginCard(info plugins.UIPluginInfo) js.Value {
	card := div("plugin-card")
	title := el("h3")
	title.Set("textContent", info.Label)

	direction := "Encode only"
	if info.CanDecode {
		direction = "Encode and decode"
	}
	metaParts := []string{plugins.CategoryLabel(info.Category), direction}
	if info.Label != info.Name {
		metaParts = append(metaParts, "Command: "+info.Name)
	}
	if len(info.Aliases) > 0 {
		metaParts = append(metaParts, "Aliases: "+strings.Join(info.Aliases, ", "))
	}
	meta := div("plugin-meta")
	meta.Set("textContent", strings.Join(metaParts, " · "))

	desc := el("p")
	desc.Set("textContent", info.Description)
	use := el("p")
	use.Set("textContent", "Use for: "+info.UseFor)
	appendChildren(card, title, meta, desc, use)

	for _, ex := range info.Examples {
		example := div("example")
		label := el("strong")
		label.Set("textContent", "Example: "+ex.Label)
		input := el("pre")
		input.Set("textContent", "Input: "+ex.Input)
		output := el("pre")
		output.Set("textContent", "Output: "+ex.Output)
		appendChildren(example, label, input, output)
		card.Call("appendChild", example)
	}

	if len(info.References) > 0 {
		refs := div("refs")
		label := el("span")
		label.Set("textContent", "References: ")
		refs.Call("appendChild", label)
		for i, ref := range info.References {
			if i > 0 {
				refs.Call("appendChild", textNode(" · "))
			}
			refs.Call("appendChild", link(ref.Label, ref.URL))
		}
		card.Call("appendChild", refs)
	}
	return card
}

func copyResult() {
	clipboard := js.Global().Get("navigator").Get("clipboard")
	if clipboard.Truthy() {
		clipboard.Call("writeText", string(pipe.Result()))
	}
}

func copyCommand() {
	command := pipe.CommandLine()
	if command == "" {
		alert("No enabled transforms to export.")
		return
	}
	clipboard := js.Global().Get("navigator").Get("clipboard")
	if clipboard.Truthy() {
		clipboard.Call("writeText", command)
		alert("Command copied.")
		return
	}
	js.Global().Call("prompt", "Command line:", command)
}

func downloadResult() {
	downloadBytes(resultDownloadName(), pipe.Result())
}

func exportChain() {
	data, err := pipe.ExportJSON()
	if err != nil {
		alert(err.Error())
		return
	}
	downloadBytes("deen-chain.json", data)
}

func copyShareLink() {
	data, err := pipe.ExportJSONWithoutSource()
	if err != nil {
		alert(err.Error())
		return
	}
	link := shareLink(data)
	clipboard := js.Global().Get("navigator").Get("clipboard")
	if clipboard.Truthy() {
		clipboard.Call("writeText", link)
		alert("Share link copied. Source input was not included.")
		return
	}
	js.Global().Get("location").Set("hash", "chain="+base64.RawURLEncoding.EncodeToString(data))
	alert("Share link added to the address bar. Source input was not included.")
}

func showSuggestions() {
	suggestions := pipeline.Suggestions(pipe.Result())
	if len(suggestions) == 0 {
		alert("No likely transforms detected.")
		return
	}
	list := div("modal-list")
	var close func()
	for _, suggestion := range suggestions {
		s := suggestion
		item := div("modal-item")
		title := el("strong")
		title.Set("textContent", s.Label)
		reason := el("p")
		reason.Set("textContent", s.Reason)
		action := el("button")
		action.Set("type", "button")
		action.Set("textContent", "Add")
		on(action, "click", func() {
			pipe.AddStepWithOptions(s.Plugin, s.Unprocess, s.Options)
			if close != nil {
				close()
			}
			rebuild()
		})
		appendChildren(item, title, reason, action)
		list.Call("appendChild", item)
	}
	close = showModal("Suggested transforms", list)
}

func searchPlugins() {
	wrap := div("search-panel")
	query := textInput("Search transformers", "")
	results := div("modal-list")
	appendChildren(wrap, query, results)

	var close func()
	const maxMatches = 24
	render := func() {
		results.Set("innerHTML", "")
		matches := plugins.SearchUICatalog(query.Get("value").String())
		if len(matches) > maxMatches {
			matches = matches[:maxMatches]
		}
		if len(matches) == 0 {
			empty := div("empty")
			empty.Set("textContent", "No transformers found.")
			results.Call("appendChild", empty)
			return
		}
		for _, match := range matches {
			info := match
			item := div("modal-item")
			title := el("strong")
			title.Set("textContent", plugins.CategoryLabel(info.Category)+" / "+info.Label)
			metaParts := []string{plugins.CategoryLabel(info.Category)}
			if info.Label != info.Name {
				metaParts = append(metaParts, "Command: "+info.Name)
			}
			if len(info.Aliases) > 0 {
				metaParts = append(metaParts, "Aliases: "+strings.Join(info.Aliases, ", "))
			}
			meta := div("plugin-meta")
			meta.Set("textContent", strings.Join(metaParts, " · "))
			desc := el("p")
			desc.Set("textContent", info.Description)
			actions := div("modal-actions")
			add := el("button")
			add.Set("type", "button")
			if info.CanDecode {
				add.Set("textContent", "Add encode")
			} else {
				add.Set("textContent", "Add")
			}
			on(add, "click", func() {
				pipe.AddStep(info.Name, false)
				if close != nil {
					close()
				}
				rebuild()
			})
			actions.Call("appendChild", add)
			if info.CanDecode {
				decode := el("button")
				decode.Set("type", "button")
				decode.Set("textContent", "Add decode")
				on(decode, "click", func() {
					pipe.AddStep(info.Name, true)
					if close != nil {
						close()
					}
					rebuild()
				})
				actions.Call("appendChild", decode)
			}
			appendChildren(item, title, meta, desc, actions)
			results.Call("appendChild", item)
		}
	}
	on(query, "input", render)
	render()
	close = showModal("Search transformers", wrap)
	query.Call("focus")
}

func showPresets() {
	list := div("modal-list")
	var close func()
	for _, preset := range pipeline.BuiltinPresets() {
		p := preset
		item := div("modal-item")
		title := el("strong")
		title.Set("textContent", p.Name)
		desc := el("p")
		desc.Set("textContent", p.Description)
		actions := div("modal-actions")
		apply := el("button")
		apply.Set("type", "button")
		apply.Set("textContent", "Apply")
		on(apply, "click", func() {
			pipe.ApplyPreset(p)
			if close != nil {
				close()
			}
			rebuild()
		})
		actions.Call("appendChild", apply)
		appendChildren(item, title, desc, actions)
		list.Call("appendChild", item)
	}
	close = showModal("Presets", list)
}

type comparePoint struct {
	label string
	data  []byte
}

func comparePoints() []comparePoint {
	points := []comparePoint{{label: "Input", data: pipe.Source()}}
	for i := range pipe.Steps() {
		points = append(points, comparePoint{
			label: fmt.Sprintf("Step %d output", i+1),
			data:  pipe.Output(i),
		})
	}
	return points
}

func compareLabels(points []comparePoint) []string {
	labels := make([]string, len(points))
	for i, point := range points {
		labels[i] = point.label
	}
	return labels
}

func compareData(points []comparePoint, label string) []byte {
	for _, point := range points {
		if point.label == label {
			return point.data
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

func showCompare() {
	points := comparePoints()
	labels := compareLabels(points)
	wrap := div("compare-panel")
	controls := div("compare-controls")
	leftSelect := selectEl("", labels, labels[0])
	rightSelect := selectEl("", labels, labels[len(labels)-1])
	modeSelect := selectEl("", []string{"text", "hex", "base64"}, "text")
	appendChildren(controls, leftSelect, rightSelect, modeSelect)

	grid := div("compare-grid")
	left := div("compare-pane")
	right := div("compare-pane")
	leftMeta := div("meta")
	rightMeta := div("meta")
	leftBody := textarea("")
	leftBody.Set("readOnly", true)
	rightBody := textarea("")
	rightBody.Set("readOnly", true)
	appendChildren(left, leftMeta, leftBody)
	appendChildren(right, rightMeta, rightBody)
	appendChildren(grid, left, right)
	appendChildren(wrap, controls, grid)

	refresh := func() {
		mode := modeSelect.Get("value").String()
		leftData := compareData(points, leftSelect.Get("value").String())
		rightData := compareData(points, rightSelect.Get("value").String())
		leftMeta.Set("textContent", pipeline.DataMetadata(leftData, 0).Summary())
		rightMeta.Set("textContent", pipeline.DataMetadata(rightData, 0).Summary())
		leftBody.Set("value", formatCompareData(leftData, mode))
		rightBody.Set("value", formatCompareData(rightData, mode))
		autoSizeTextarea(leftBody)
		autoSizeTextarea(rightBody)
	}
	on(leftSelect, "change", refresh)
	on(rightSelect, "change", refresh)
	on(modeSelect, "change", refresh)
	refresh()
	showModal("Compare pipeline data", wrap)
}

func shareLink(data []byte) string {
	loc := js.Global().Get("location")
	return loc.Get("origin").String() +
		loc.Get("pathname").String() +
		loc.Get("search").String() +
		"#chain=" + base64.RawURLEncoding.EncodeToString(data)
}

func loadChainFromHash() {
	hash := js.Global().Get("location").Get("hash").String()
	hash = strings.TrimPrefix(hash, "#")
	if !strings.HasPrefix(hash, "chain=") {
		return
	}
	encoded := strings.TrimPrefix(hash, "chain=")
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		alert("Could not decode chain from URL: " + err.Error())
		return
	}
	if err := pipe.ImportJSON(data); err != nil {
		alert("Could not import chain from URL: " + err.Error())
	}
}

func downloadBytes(name string, data []byte) {
	array := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(array, data)
	blob := js.Global().Get("Blob").New([]any{array})
	urlObj := js.Global().Get("URL")
	u := urlObj.Call("createObjectURL", blob)
	a := el("a")
	a.Set("href", u)
	a.Set("download", name)
	doc.Get("body").Call("appendChild", a)
	a.Call("click")
	a.Call("remove")
	urlObj.Call("revokeObjectURL", u)
}

func resultDownloadName() string {
	if sourceName == "" {
		return "deen-result.bin"
	}
	name := strings.TrimSpace(sourceName)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	if name == "" || name == "." || name == ".." {
		return "deen-result.bin"
	}
	if dot := strings.LastIndex(name, "."); dot > 0 && dot < len(name)-1 {
		return name[:dot] + ".deen-result" + name[dot:]
	}
	return name + ".deen-result"
}

func alert(message string) {
	js.Global().Call("alert", message)
}

func loadSourceFile(file js.Value) {
	reader := js.Global().Get("FileReader").New()
	var loadCB js.Func
	loadCB = js.FuncOf(func(js.Value, []js.Value) any {
		array := js.Global().Get("Uint8Array").New(reader.Get("result"))
		buf := make([]byte, array.Get("byteLength").Int())
		js.CopyBytesToGo(buf, array)
		sourceName = file.Get("name").String()
		pipe.SetSource(buf)
		rebuild()
		loadCB.Release()
		return nil
	})
	reader.Call("addEventListener", "load", loadCB)
	reader.Call("readAsArrayBuffer", file)
}

func filePicker() js.Value {
	wrap := div("file-action")
	openBtn := el("button")
	openBtn.Set("type", "button")
	setIconButtonContent(openBtn, "upload", "Open file")
	input := el("input")
	input.Set("type", "file")
	input.Get("style").Set("display", "none")

	on(openBtn, "click", func() {
		input.Call("click")
	})
	on(input, "change", func() {
		files := input.Get("files")
		if files.Get("length").Int() == 0 {
			return
		}
		loadSourceFile(files.Call("item", 0))
	})
	appendChildren(wrap, openBtn, input)
	return wrap
}

func chainPicker() js.Value {
	wrap := div("file-action")
	openBtn := el("button")
	openBtn.Set("type", "button")
	setIconButtonContent(openBtn, "upload", "Import chain")
	input := el("input")
	input.Set("type", "file")
	input.Set("accept", "application/json,.json")
	input.Get("style").Set("display", "none")

	on(openBtn, "click", func() {
		input.Call("click")
	})
	on(input, "change", func() {
		files := input.Get("files")
		if files.Get("length").Int() == 0 {
			return
		}
		file := files.Call("item", 0)
		reader := js.Global().Get("FileReader").New()
		var loadCB js.Func
		loadCB = js.FuncOf(func(js.Value, []js.Value) any {
			data := []byte(reader.Get("result").String())
			if err := pipe.ImportJSON(data); err != nil {
				alert(err.Error())
			} else {
				sourceName = ""
				rebuild()
			}
			loadCB.Release()
			return nil
		})
		reader.Call("addEventListener", "load", loadCB)
		reader.Call("readAsText", file)
	})
	appendChildren(wrap, openBtn, input)
	return wrap
}

func sourceCard() js.Value {
	card := div("card source")
	label := el("div")
	label.Set("className", "card-title")
	label.Set("textContent", "Input")

	ta := textarea(string(pipe.Source()))
	sourceEl = ta
	meta := div("meta")
	meta.Set("textContent", sourceMetadataSummary())
	on(ta, "input", func() {
		if updating {
			return
		}
		sourceName = ""
		pipe.SetSource([]byte(ta.Get("value").String()))
		meta.Set("textContent", sourceMetadataSummary())
		refreshOutputs(0)
	})
	onEvent(card, "dragover", func(ev js.Value) {
		ev.Call("preventDefault")
		card.Set("className", "card source drag-over")
	})
	onEvent(card, "dragleave", func(ev js.Value) {
		ev.Call("preventDefault")
		card.Set("className", "card source")
	})
	onEvent(card, "drop", func(ev js.Value) {
		ev.Call("preventDefault")
		card.Set("className", "card source")
		files := ev.Get("dataTransfer").Get("files")
		if files.Get("length").Int() == 0 {
			return
		}
		loadSourceFile(files.Call("item", 0))
	})
	appendChildren(card, label, ta, meta)
	autoSizeTextarea(ta)
	return card
}

func sourceMetadataSummary() string {
	summary := pipeline.DataMetadata(pipe.Source(), 0).Summary()
	if strings.TrimSpace(sourceName) == "" {
		return summary
	}
	return "source: " + sourceName + " · " + summary
}

func stepCard(i int) js.Value {
	step := pipe.Steps()[i]
	col := accent(i)
	ref := &cardRef{index: i}
	cards = append(cards, ref)

	card := div("card")
	card.Get("style").Set("borderLeft", "5px solid "+col)

	// Header: collapse, title, summary, controls.
	header := div("card-header")
	collapse := el("button")
	collapse.Set("className", "icon")
	collapse.Set("textContent", "▾")
	title := el("span")
	title.Set("className", "title")
	title.Set("textContent", fmt.Sprintf("Step %d", i+1))
	title.Get("style").Set("color", col)
	summary := el("span")
	summary.Set("className", "summary")
	summary.Set("textContent", summaryText(i))
	remove := el("button")
	remove.Set("className", "icon")
	remove.Set("textContent", "✕")
	on(remove, "click", func() { pipe.RemoveStep(i); rebuild() })
	moveUp := el("button")
	moveUp.Set("className", "icon")
	moveUp.Set("textContent", "↑")
	on(moveUp, "click", func() {
		pipe.MoveStep(i, i-1)
		rebuild()
	})
	moveDown := el("button")
	moveDown.Set("className", "icon")
	moveDown.Set("textContent", "↓")
	on(moveDown, "click", func() {
		pipe.MoveStep(i, i+1)
		rebuild()
	})
	duplicate := el("button")
	duplicate.Set("className", "icon")
	duplicate.Set("textContent", "⧉")
	on(duplicate, "click", func() {
		pipe.DuplicateStep(i)
		rebuild()
	})
	enabledWrap, enabledInput := checkbox("enabled", !step.Disabled)
	enabledWrap.Set("className", "header-toggle")
	enabledInput.Set("checked", !step.Disabled)
	on(enabledInput, "change", func() {
		pipe.SetStepDisabled(i, !enabledInput.Get("checked").Bool())
		rebuild()
	})

	spacer := div("spacer")
	appendChildren(header, collapse, title, summary, spacer, enabledWrap, moveUp, moveDown, duplicate, remove)

	detail := div("card-detail")

	// decode toggle (referenced by the category selectors).
	canDecode := plugins.CanDecode(step.Plugin)
	decodeWrap, decodeInput := checkbox("decode", step.Unprocess && canDecode)
	decodeInput.Set("checked", step.Unprocess && canDecode)

	// One dropdown per category.
	selRow := div("selectors")
	for _, cat := range plugins.PluginCategories {
		selected := ""
		if plugins.CategoryOf(step.Plugin) == cat {
			selected = step.Plugin
		}
		sel := selectOptionsEl(plugins.CategoryLabel(cat), pluginSelectOptions(cat), selected)
		on(sel, "change", func() {
			name := sel.Get("value").String()
			if name == "" {
				return
			}
			decode := decodeInput.Get("checked").Bool() && plugins.CanDecode(name)
			pipe.SetPlugin(i, name, decode)
			rebuild()
		})
		selRow.Call("appendChild", sel)
	}

	on(decodeInput, "change", func() {
		pipe.SetPlugin(i, step.Plugin, decodeInput.Get("checked").Bool() && plugins.CanDecode(step.Plugin))
		refreshOutputs(i)
	})

	toggles := div("toggles")
	if canDecode {
		toggles.Call("appendChild", decodeWrap)
	}

	options := div("options")
	buildOptions(options, i)

	ref.output = textareaWithMax("", 960)
	on(ref.output, "input", func() {
		if updating {
			return
		}
		pipe.EditOutput(i, []byte(ref.output.Get("value").String()))
		refreshOutputs(i + 1)
	})
	ref.hexOutput = textareaWithMax("", 960)
	ref.hexOutput.Set("readOnly", true)
	ref.preview = previewBox()
	ref.imageEnabled = stepGeneratesImage(step)
	if ref.imageEnabled {
		ref.imagePanel, ref.image, ref.imageMsg = imagePreviewBox()
	}
	viewer := outputViewer(ref)

	ref.meta = div("meta")
	ref.errEl = div("error")
	ref.errEl.Get("style").Set("display", "none")

	appendChildren(detail, selRow, toggles, options, viewer, ref.meta, ref.errEl)
	on(collapse, "click", func() {
		if detail.Get("style").Get("display").String() == "none" {
			detail.Get("style").Set("display", "block")
			collapse.Set("textContent", "▾")
			autoSizeOutputTextareas(ref)
		} else {
			detail.Get("style").Set("display", "none")
			collapse.Set("textContent", "▸")
		}
	})

	appendChildren(card, header, detail)
	renderOutput(ref)
	return card
}

func outputViewer(ref *cardRef) js.Value {
	viewer := div("viewer")
	tabs := div("viewer-tabs")
	panels := div("viewer-panels")

	panelItems := []struct {
		label string
		node  js.Value
	}{
		{"Text", ref.output},
		{"Hex", ref.hexOutput},
		{"Preview", ref.preview},
	}
	if ref.imageEnabled {
		panelItems = append(panelItems, struct {
			label string
			node  js.Value
		}{"Image", ref.imagePanel})
	}
	buttons := make([]js.Value, 0, len(panelItems))
	panelEls := make([]js.Value, 0, len(panelItems))
	var activate func(int)
	for idx, item := range panelItems {
		panel := div("viewer-panel")
		if idx != 0 {
			panel.Set("className", "viewer-panel hidden")
		}
		panel.Call("appendChild", item.node)
		panelEls = append(panelEls, panel)
		btn := el("button")
		btn.Set("type", "button")
		if idx == 0 {
			btn.Set("className", "viewer-tab active")
		} else {
			btn.Set("className", "viewer-tab")
		}
		btn.Set("textContent", item.label)
		i := idx
		on(btn, "click", func() { activate(i) })
		buttons = append(buttons, btn)
		tabs.Call("appendChild", btn)
		panels.Call("appendChild", panel)
	}
	activate = func(active int) {
		for i, btn := range buttons {
			if i == active {
				btn.Set("className", "viewer-tab active")
				panelEls[i].Set("className", "viewer-panel")
				autoSizeViewerPanel(panelEls[i])
			} else {
				btn.Set("className", "viewer-tab")
				panelEls[i].Set("className", "viewer-panel hidden")
			}
		}
	}

	appendChildren(viewer, tabs, panels)
	return viewer
}

func stepGeneratesImage(step *pipeline.Step) bool {
	return step.Plugin == "qr" && !step.Unprocess
}

func autoSizeViewerPanel(panel js.Value) {
	areas := panel.Call("querySelectorAll", "textarea.io")
	for i := 0; i < areas.Get("length").Int(); i++ {
		autoSizeTextareaSoon(areas.Call("item", i))
	}
}

func buildOptions(container js.Value, i int) {
	step := pipe.Steps()[i]
	opts := pipeline.PluginOptions(step.Plugin)
	if len(opts) == 0 {
		return
	}
	checkGroup, checkBody := optionGroup("Checkboxes")
	fieldGroup, fieldBody := optionGroup("Inputs")
	var hasChecks, hasFields bool
	for _, opt := range opts {
		opt := opt
		row := div("option")
		row.Set("title", opt.Usage)
		row.Get("classList").Call("toggle", "option-multiline", opt.Multiline)
		if opt.IsBool {
			wrap, input := checkbox(opt.Label, step.Options[opt.Name] == "true")
			on(input, "change", func() {
				val := "false"
				if input.Get("checked").Bool() {
					val = "true"
				}
				pipe.SetOption(i, opt.Name, val)
				refreshOutputs(i)
			})
			appendChildren(row, wrap, optionHelp(opt))
			checkBody.Call("appendChild", row)
			hasChecks = true
		} else {
			label := el("label")
			label.Set("className", "option-label")
			label.Set("textContent", opt.Label)
			var input js.Value
			current := opt.Default
			if v, ok := step.Options[opt.Name]; ok {
				current = v
			}
			if opt.Kind == "select" {
				input = selectEl("", opt.Choices, current)
				input.Set("title", opt.Description)
				on(input, "change", func() {
					pipe.SetOption(i, opt.Name, input.Get("value").String())
					refreshOutputs(i)
				})
			} else {
				if opt.Multiline {
					input = el("textarea")
					input.Set("rows", 3)
					input.Set("placeholder", optionPlaceholder(opt))
					input.Set("value", current)
					input.Set("title", opt.Description)
				} else {
					input = el("input")
					inputType := "text"
					if opt.Kind == "number" {
						inputType = "number"
					}
					if opt.Kind == "secret" || opt.Secret {
						inputType = "password"
					}
					input.Set("type", inputType)
					input.Set("placeholder", optionPlaceholder(opt))
					input.Set("title", opt.Description)
					if v, ok := step.Options[opt.Name]; ok {
						input.Set("value", v)
					}
				}
				on(input, "input", func() {
					pipe.SetOption(i, opt.Name, input.Get("value").String())
					refreshOutputs(i)
				})
			}
			appendChildren(row, label, input, optionHelp(opt))
			fieldBody.Call("appendChild", row)
			hasFields = true
		}
	}
	if hasChecks {
		container.Call("appendChild", checkGroup)
	}
	if hasFields {
		container.Call("appendChild", fieldGroup)
	}
}

func optionGroup(title string) (js.Value, js.Value) {
	group := div("option-group")
	heading := div("option-group-title")
	heading.Set("textContent", title)
	body := div("option-group-body")
	appendChildren(group, heading, body)
	return group, body
}

func optionPlaceholder(opt pipeline.Option) string {
	if opt.Default == "" {
		return opt.Label
	}
	return "default: " + opt.Default
}

func optionHelp(opt pipeline.Option) js.Value {
	help := div("option-help")
	if opt.Description != "" {
		desc := el("span")
		desc.Set("textContent", opt.Description)
		help.Call("appendChild", desc)
	}
	if opt.HelpURL != "" {
		link := el("a")
		link.Set("href", opt.HelpURL)
		link.Set("target", "_blank")
		link.Set("rel", "noopener noreferrer")
		if opt.HelpLabel != "" {
			link.Set("textContent", opt.HelpLabel)
		} else {
			link.Set("textContent", "Reference")
		}
		help.Call("appendChild", link)
	}
	return help
}

func addCard() js.Value {
	card := div("card add")
	header := div("add-header")
	titleBlock := div("add-title")
	label := el("div")
	label.Set("className", "card-title")
	label.Set("textContent", "Add transformer step")
	subtitle := div("add-subtitle")
	subtitle.Set("textContent", "Choose a transformer by category or search the catalog.")
	appendChildren(titleBlock, label, subtitle)
	actions := div("add-actions")
	appendChildren(actions,
		button("", "Search transformers", searchPlugins),
		button("", "Detect next", showSuggestions),
	)
	appendChildren(header, titleBlock, actions)
	selRow := div("add-selectors")
	for _, cat := range plugins.PluginCategories {
		group := div("add-picker")
		pickerLabel := el("label")
		pickerLabel.Set("textContent", plugins.CategoryLabel(cat))
		sel := selectOptionsEl(plugins.CategorySelectLabel(cat), pluginSelectOptions(cat), "")
		on(sel, "change", func() {
			name := sel.Get("value").String()
			if name == "" {
				return
			}
			pipe.AddStep(name, false)
			rebuild()
		})
		appendChildren(group, pickerLabel, sel)
		selRow.Call("appendChild", group)
	}
	appendChildren(card, header, selRow)
	return card
}

func updateHistory() {
	historyEl.Set("innerHTML", "")
	input := div("hist")
	input.Set("textContent", "Input")
	historyEl.Call("appendChild", input)
	for i, s := range pipe.Steps() {
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
		line := div("hist")
		line.Set("textContent", fmt.Sprintf("%d. %s (%s)", i+1, name, dir))
		line.Get("style").Set("color", accent(i))
		historyEl.Call("appendChild", line)
	}
}

func refreshOutputs(from int) {
	for _, c := range cards {
		if c.index >= from {
			renderOutput(c)
		}
	}
	updateHistory()
}

func renderOutput(c *cardRef) {
	out := pipe.Output(c.index)
	inputBytes := len(pipe.Source())
	if c.index > 0 {
		inputBytes = len(pipe.Output(c.index - 1))
	}
	summary := pipeline.DataMetadata(out, inputBytes).Summary()
	c.meta.Set("textContent", summary)
	if err := pipe.Err(c.index); err != nil {
		c.errEl.Set("textContent", "error: "+err.Error())
		c.errEl.Get("style").Set("display", "block")
	} else {
		c.errEl.Get("style").Set("display", "none")
	}
	updating = true
	c.output.Set("value", string(out))
	c.output.Set("readOnly", false)
	c.hexOutput.Set("value", hex.Dump(out))
	autoSizeOutputTextareas(c)
	if c.imageEnabled {
		renderImageOutput(c, out)
	}
	preview, ok := pipeline.StructuredPreview(out)
	if !ok {
		preview = "No structured preview available."
		renderHighlightedText(c.preview, preview, nil)
	} else {
		preview, spans, _ := pipeline.HighlightedPreview(out)
		renderHighlightedText(c.preview, preview, spans)
	}
	updating = false
}

func imagePreviewBox() (wrap, img, msg js.Value) {
	wrap = div("image-preview")
	img = el("img")
	img.Set("alt", "Output image preview")
	msg = div("image-preview-message")
	appendChildren(wrap, img, msg)
	return wrap, img, msg
}

func renderImageOutput(c *cardRef, data []byte) {
	imageData, mime := imagePreviewPayload(data)
	if mime == "" {
		c.image.Get("style").Set("display", "none")
		c.image.Set("src", "")
		c.imageMsg.Set("textContent", "No image preview available.")
		return
	}
	c.image.Set("src", "data:"+mime+";base64,"+base64.StdEncoding.EncodeToString(imageData))
	c.image.Get("style").Set("display", "block")
	c.imageMsg.Set("textContent", mime)
}

func imagePreviewPayload(data []byte) ([]byte, string) {
	if mime := imageMIME(data); mime != "" {
		return data, mime
	}
	compact := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, string(data))
	if len(compact) < 8 {
		return nil, ""
	}
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := enc.DecodeString(compact)
		if err != nil {
			continue
		}
		if mime := imageMIME(decoded); mime != "" {
			return decoded, mime
		}
	}
	return nil, ""
}

func imageMIME(data []byte) string {
	switch {
	case len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n":
		return "image/png"
	case len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff:
		return "image/jpeg"
	case len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a"):
		return "image/gif"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp"
	case looksLikeSVG(data):
		return "image/svg+xml"
	default:
		return ""
	}
}

func looksLikeSVG(data []byte) bool {
	s := strings.TrimSpace(string(data))
	return strings.HasPrefix(s, "<svg ") || strings.HasPrefix(s, "<svg>")
}

func renderHighlightedText(node js.Value, text string, spans []pipeline.SyntaxSpan) {
	node.Set("innerHTML", "")
	pos := 0
	for _, span := range spans {
		if span.Start < pos || span.End > len(text) || span.Start >= span.End {
			continue
		}
		if span.Start > pos {
			node.Call("appendChild", textNode(text[pos:span.Start]))
		}
		part := el("span")
		part.Set("className", "syntax-"+string(span.Kind))
		part.Set("textContent", text[span.Start:span.End])
		node.Call("appendChild", part)
		pos = span.End
	}
	if pos < len(text) {
		node.Call("appendChild", textNode(text[pos:]))
	}
}

func summaryText(i int) string {
	step := pipe.Steps()[i]
	if step.Plugin == "" {
		return "(no transform)"
	}
	dir := "encode"
	if step.Unprocess {
		dir = "decode"
	}
	if step.Disabled {
		dir += " · disabled"
	}
	return fmt.Sprintf("%s / %s · %s", plugins.CategoryLabel(plugins.CategoryOf(step.Plugin)), plugins.PluginLabel(step.Plugin), dir)
}
