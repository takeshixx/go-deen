//go:build js && wasm

// Package webui renders the deen browser interface: a Burp Decoder-style chain
// of plugin transforms, mirroring the desktop GUI, built on the shared
// internal/pipeline model via syscall/js.
package webui

import (
	"encoding/hex"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/takeshixx/deen/internal/core"
	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

var palette = []string{"#4285f4", "#0f9d58", "#f4b400", "#db4437", "#ab47bc", "#00acc1"}

func accent(i int) string { return palette[i%len(palette)] }

var (
	doc       js.Value
	pipe      *pipeline.Pipeline
	contentEl js.Value
	tabBtns   []js.Value
	stepsEl   js.Value
	historyEl js.Value
	sourceEl  js.Value
	updating  bool
	callbacks []js.Func
	staticCBs []js.Func
	cards     []*cardRef
	activeTab = "home"
)

type cardRef struct {
	index   int
	output  js.Value
	errEl   js.Value
	hexView bool
}

// Run builds the UI and blocks so the event callbacks stay alive.
func Run() {
	doc = js.Global().Get("document")
	pipe = pipeline.New()
	buildLayout()
	select {}
}

// --- small DOM helpers ---

func el(tag string) js.Value { return doc.Call("createElement", tag) }

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
	t := el("textarea")
	t.Set("className", "io")
	t.Set("value", value)
	return t
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

// --- layout ---

func buildLayout() {
	body := doc.Get("body")
	body.Set("innerHTML", "")

	shell := div("shell")
	header := div("app-header")
	heading := el("h1")
	heading.Set("textContent", "deen")
	tabs := div("tabs")
	appendChildren(tabs, tabButton("home", "Home"), tabButton("plugins", "Plugins"), tabButton("about", "About"))
	appendChildren(header, heading, tabs)

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
	main := div("main")
	side := div("side")

	toolbar := div("toolbar")
	appendChildren(toolbar,
		button("primary", "Copy result", copyResult),
		button("", "Download", downloadResult),
		filePicker(),
		button("", "Clear", func() {
			pipe = pipeline.New()
			rebuild()
		}),
	)

	stepsEl = div("steps")

	stepsEl.Call("appendChild", sourceCard())
	for i := range pipe.Steps() {
		stepsEl.Call("appendChild", stepCard(i))
	}
	stepsEl.Call("appendChild", addCard())

	sideTitle := el("h2")
	sideTitle.Set("textContent", "History")
	historyEl = div("history")

	appendChildren(main, toolbar, stepsEl)
	appendChildren(side, sideTitle, historyEl)
	appendChildren(app, main, side)
	contentEl.Call("appendChild", app)
	updateHistory()
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
	title.Set("textContent", info.Name)

	direction := "Encode only"
	if info.CanDecode {
		direction = "Encode and decode"
	}
	metaParts := []string{info.Category, direction}
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

func downloadResult() {
	data := pipe.Result()
	array := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(array, data)
	blob := js.Global().Get("Blob").New([]any{array})
	urlObj := js.Global().Get("URL")
	u := urlObj.Call("createObjectURL", blob)
	a := el("a")
	a.Set("href", u)
	a.Set("download", "deen-result.bin")
	doc.Get("body").Call("appendChild", a)
	a.Call("click")
	a.Call("remove")
	urlObj.Call("revokeObjectURL", u)
}

func filePicker() js.Value {
	wrap := div("file-action")
	openBtn := el("button")
	openBtn.Set("type", "button")
	openBtn.Set("textContent", "Open file")
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
		file := files.Call("item", 0)
		reader := js.Global().Get("FileReader").New()
		var loadCB js.Func
		loadCB = js.FuncOf(func(js.Value, []js.Value) any {
			array := js.Global().Get("Uint8Array").New(reader.Get("result"))
			buf := make([]byte, array.Get("byteLength").Int())
			js.CopyBytesToGo(buf, array)
			pipe.SetSource(buf)
			rebuild()
			loadCB.Release()
			return nil
		})
		reader.Call("addEventListener", "load", loadCB)
		reader.Call("readAsArrayBuffer", file)
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
	on(ta, "input", func() {
		if updating {
			return
		}
		pipe.SetSource([]byte(ta.Get("value").String()))
		refreshOutputs(0)
	})
	appendChildren(card, label, ta)
	return card
}

func stepCard(i int) js.Value {
	step := pipe.Steps()[i]
	col := accent(i)
	ref := &cardRef{index: i}
	cards = append(cards, ref)

	card := div("card")
	card.Get("style").Set("borderLeft", "5px solid "+col)

	// Header: collapse, title, summary, remove.
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

	spacer := div("spacer")
	appendChildren(header, collapse, title, summary, spacer, remove)

	detail := div("card-detail")

	// decode toggle (referenced by the category selectors).
	decodeWrap, decodeInput := checkbox("decode", step.Unprocess)
	decodeInput.Set("checked", step.Unprocess)

	// One dropdown per category.
	selRow := div("selectors")
	for _, cat := range plugins.PluginCategories {
		selected := ""
		if plugins.CategoryOf(step.Plugin) == cat {
			selected = step.Plugin
		}
		sel := selectEl(cat, plugins.InCategory(cat), selected)
		on(sel, "change", func() {
			name := sel.Get("value").String()
			if name == "" {
				return
			}
			pipe.SetPlugin(i, name, decodeInput.Get("checked").Bool())
			rebuild()
		})
		selRow.Call("appendChild", sel)
	}

	on(decodeInput, "change", func() {
		pipe.SetPlugin(i, step.Plugin, decodeInput.Get("checked").Bool())
		refreshOutputs(i)
	})

	hexWrap, hexInput := checkbox("hex", false)
	on(hexInput, "change", func() {
		ref.hexView = hexInput.Get("checked").Bool()
		renderOutput(ref)
	})

	toggles := div("toggles")
	appendChildren(toggles, decodeWrap, hexWrap)

	options := div("options")
	buildOptions(options, i)

	ref.output = textarea("")
	on(ref.output, "input", func() {
		if updating || ref.hexView {
			return
		}
		pipe.EditOutput(i, []byte(ref.output.Get("value").String()))
		refreshOutputs(i + 1)
	})

	ref.errEl = div("error")
	ref.errEl.Get("style").Set("display", "none")

	appendChildren(detail, selRow, toggles, options, ref.output, ref.errEl)
	on(collapse, "click", func() {
		if detail.Get("style").Get("display").String() == "none" {
			detail.Get("style").Set("display", "block")
			collapse.Set("textContent", "▾")
		} else {
			detail.Get("style").Set("display", "none")
			collapse.Set("textContent", "▸")
		}
	})

	appendChildren(card, header, detail)
	renderOutput(ref)
	return card
}

func buildOptions(container js.Value, i int) {
	step := pipe.Steps()[i]
	for _, opt := range pipeline.PluginOptions(step.Plugin) {
		opt := opt
		row := div("option")
		if opt.IsBool {
			wrap, input := checkbox(opt.Name, step.Options[opt.Name] == "true")
			on(input, "change", func() {
				val := "false"
				if input.Get("checked").Bool() {
					val = "true"
				}
				pipe.SetOption(i, opt.Name, val)
				refreshOutputs(i)
			})
			row.Call("appendChild", wrap)
		} else {
			label := el("span")
			label.Set("textContent", opt.Name+": ")
			input := el("input")
			input.Set("type", "text")
			input.Set("placeholder", "default: "+opt.Default)
			if v, ok := step.Options[opt.Name]; ok {
				input.Set("value", v)
			}
			on(input, "input", func() {
				pipe.SetOption(i, opt.Name, input.Get("value").String())
				refreshOutputs(i)
			})
			appendChildren(row, label, input)
		}
		container.Call("appendChild", row)
	}
}

func addCard() js.Value {
	card := div("card add")
	label := el("div")
	label.Set("className", "card-title")
	label.Set("textContent", "Add transform")
	selRow := div("selectors")
	for _, cat := range plugins.PluginCategories {
		sel := selectEl(cat, plugins.InCategory(cat), "")
		on(sel, "change", func() {
			name := sel.Get("value").String()
			if name == "" {
				return
			}
			pipe.AddStep(name, false)
			rebuild()
		})
		selRow.Call("appendChild", sel)
	}
	appendChildren(card, label, selRow)
	return card
}

func updateHistory() {
	historyEl.Set("innerHTML", "")
	input := div("hist")
	input.Set("textContent", "Input")
	historyEl.Call("appendChild", input)
	for i, s := range pipe.Steps() {
		name := s.Plugin
		if name == "" {
			name = "(none)"
		}
		dir := "encode"
		if s.Unprocess {
			dir = "decode"
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
	if err := pipe.Err(c.index); err != nil {
		c.errEl.Set("textContent", "error: "+err.Error())
		c.errEl.Get("style").Set("display", "block")
	} else {
		c.errEl.Get("style").Set("display", "none")
	}
	updating = true
	if c.hexView {
		c.output.Set("value", hex.Dump(out))
		c.output.Set("readOnly", true)
	} else {
		c.output.Set("value", string(out))
		c.output.Set("readOnly", false)
	}
	updating = false
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
	return fmt.Sprintf("%s / %s · %s", plugins.CategoryOf(step.Plugin), step.Plugin, dir)
}
