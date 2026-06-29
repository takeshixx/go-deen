//go:build js && wasm

// Package webui renders the deen browser interface: a Burp Decoder-style chain
// of plugin transforms, mirroring the desktop GUI, built on the shared
// internal/pipeline model via syscall/js.
package webui

import (
	"encoding/base64"
	"fmt"
	"net/url"
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
	doc                js.Value
	pipe               *pipeline.Pipeline
	contentEl          js.Value
	tabBtns            []js.Value
	stepsEl            js.Value
	historyEl          js.Value
	sourceEl           js.Value
	busyEl             js.Value
	updating           bool
	callbacks          []js.Func
	staticCBs          []js.Func
	cards              []*cardRef
	activeTab          = "home"
	aboutPage          js.Value
	aboutPageBuilt     bool
	pluginsPage        js.Value
	pluginsPageBuilt   bool
	examplesPage       js.Value
	examplesPageBuilt  bool
	examplePreviews    = map[string]examplePreview{}
	applyExamplesRoute func(webRoute)
	applyPluginsRoute  func(webRoute)
	currentRoute       webRoute
	updatingRoute      bool
	sourceName         string
	sourceFullRaw      bool
	sourceFullHex      bool
	sourceFullStrings  bool
)

type cardRef struct {
	index             int
	output            js.Value
	hexOutput         js.Value
	stringsOutput     js.Value
	preview           js.Value
	rawButton         js.Value
	hexButton         js.Value
	stringsButton     js.Value
	previewButton     js.Value
	rawPanel          js.Value
	hexPanel          js.Value
	stringsPanel      js.Value
	previewPanel      js.Value
	activeOutputView  string
	activateOutput    func(string)
	imageEnabled      bool
	imageButton       js.Value
	imagePanel        js.Value
	image             js.Value
	imageMsg          js.Value
	fullControls      js.Value
	rawFullButton     js.Value
	hexFullButton     js.Value
	stringsFullButton js.Value
	fullWarning       js.Value
	meta              js.Value
	errEl             js.Value
	fullRaw           bool
	fullHex           bool
	fullStrings       bool
}

type examplePreview struct {
	result []byte
	err    error
}

type exampleCardRef struct {
	example pipeline.Example
	card    js.Value
}

type pluginCardRef struct {
	info plugins.UIPluginInfo
	card js.Value
}

type webRoute struct {
	tab    string
	item   string
	search string
	chain  string
}

// Run builds the UI and blocks so the event callbacks stay alive.
func Run() {
	doc = js.Global().Get("document")
	pipe = pipeline.New()
	currentRoute = parseRoute()
	loadChainFromRoute(currentRoute)
	if currentRoute.tab != "" {
		activeTab = currentRoute.tab
	}
	buildLayout()
	onStatic(js.Global(), "hashchange", applyRouteFromHash)
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
	return textareaWithMaxAttr(value, strconv.Itoa(maxHeight))
}

func sourceTextarea(value string) js.Value {
	t := textareaWithMaxAttr(value, "window")
	t.Get("classList").Call("add", "source-io")
	return t
}

func textareaWithMaxAttr(value, maxHeight string) js.Value {
	t := el("textarea")
	t.Set("className", "io")
	t.Set("value", value)
	t.Call("setAttribute", "data-max-height", maxHeight)
	autoSizeTextarea(t)
	on(t, "input", func() { autoSizeTextarea(t) })
	return t
}

func autoSizeTextarea(t js.Value) {
	t.Get("style").Set("height", "auto")
	maxHeight := 512
	if attr := t.Call("getAttribute", "data-max-height"); attr.Truthy() {
		if attr.String() == "window" {
			rect := t.Call("getBoundingClientRect")
			top := rect.Get("top").Float()
			available := (js.Global().Get("innerHeight").Float() - top - 24) * 0.85
			if available < 160 {
				available = 160
			}
			maxHeight = int(available)
		} else if n, err := strconv.Atoi(attr.String()); err == nil && n > 0 {
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
	} else if t.Get("classList").Call("contains", "source-io").Bool() {
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

func autoSizeSourceTextareasSoon() {
	areas := contentEl.Call("querySelectorAll", ".source textarea.source-io")
	for i := 0; i < areas.Get("length").Int(); i++ {
		autoSizeTextareaSoon(areas.Call("item", i))
	}
}

func autoSizeOutputTextareas(c *cardRef) {
	autoSizeTextareaSoon(c.output)
	autoSizeTextareaSoon(c.hexOutput)
	autoSizeTextareaSoon(c.stringsOutput)
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
	b.Set("title", label)
	b.Set("aria-label", label)
	setIconButtonContent(b, icon, label)
	on(b, "click", fn)
	return b
}

func iconOnlyButton(icon, label string, fn func()) js.Value {
	b := el("button")
	b.Set("className", "icon")
	b.Set("type", "button")
	setIconOnlyButtonContent(b, icon, label)
	if fn != nil {
		on(b, "click", fn)
	}
	return b
}

func setIconOnlyButtonContent(b js.Value, icon, label string) {
	b.Set("title", label)
	b.Set("aria-label", label)
	b.Set("textContent", "")
	b.Call("appendChild", iconGraphic(icon))
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
	"copy":         {"M8 8h11v11H8z", "M5 16H4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1"},
	"download":     {"M12 3v12", "M7 10l5 5 5-5", "M5 21h14"},
	"upload":       {"M12 21V9", "M7 14l5-5 5 5", "M5 3h14"},
	"link":         {"M10 13a5 5 0 0 0 7.1 0l2-2a5 5 0 0 0-7.1-7.1l-1.2 1.2", "M14 11a5 5 0 0 0-7.1 0l-2 2A5 5 0 0 0 12 20.1l1.2-1.2"},
	"terminal":     {"M4 17l6-5-6-5", "M12 19h8"},
	"star":         {"M12 3l2.7 5.5 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.8 1-6.1-4.4-4.3 6.1-.9z"},
	"compare":      {"M8 6h13", "M8 12h13", "M8 18h13", "M3 6h.01", "M3 12h.01", "M3 18h.01"},
	"undo":         {"M9 14l-5-5 5-5", "M4 9h10a6 6 0 0 1 0 12h-1"},
	"redo":         {"M15 14l5-5-5-5", "M20 9H10a6 6 0 0 0 0 12h1"},
	"clear":        {"M18 6L6 18", "M6 6l12 12"},
	"collapse":     {"M15 18l-6-6 6-6"},
	"expand":       {"M9 18l6-6-6-6"},
	"visible":      {"M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z", "M12 9a3 3 0 1 1 0 6 3 3 0 0 1 0-6z"},
	"hidden":       {"M3 3l18 18", "M10.6 10.6a2 2 0 0 0 2.8 2.8", "M9.9 4.2A10.7 10.7 0 0 1 12 4c6.5 0 10 8 10 8a18.4 18.4 0 0 1-3.2 4.4", "M6.1 6.1C3.5 8 2 12 2 12s3.5 8 10 8a9.8 9.8 0 0 0 4.1-.9"},
	"arrow-up":     {"M12 19V5", "M5 12l7-7 7 7"},
	"arrow-down":   {"M12 5v14", "M19 12l-7 7-7-7"},
	"trash":        {"M3 6h18", "M8 6V4h8v2", "M6 6l1 15h10l1-15"},
	"chevron-down": {"M6 9l6 6 6-6"},
	"folder":       {"M3 6h7l2 2h9v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"},
	"home":         {"M3 11l9-8 9 8", "M5 10v10h14V10", "M9 20v-6h6v6"},
	"examples":     {"M4 19.5A2.5 2.5 0 0 1 6.5 17H20", "M4 4.5A2.5 2.5 0 0 1 6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5z"},
	"plugins":      {"M9 3v5", "M15 3v5", "M6 8h12", "M7 8v4a5 5 0 0 0 10 0V8", "M12 17v4"},
	"info":         {"M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20z", "M12 16v-4", "M12 8h.01"},
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
	busyEl = div("busy-overlay")
	busyEl.Set("role", "status")
	busyEl.Set("aria-live", "polite")
	busyEl.Set("aria-hidden", "true")
	spinner := div("busy-spinner")
	label := div("busy-label")
	label.Set("textContent", "Processing")
	appendChildren(busyEl, spinner, label)
	body.Call("appendChild", busyEl)

	renderActiveTab()
}

func setBusy(label string, show bool) {
	if !busyEl.Truthy() {
		return
	}
	if label == "" {
		label = "Processing"
	}
	busyEl.Call("querySelector", ".busy-label").Set("textContent", label)
	if show {
		busyEl.Set("className", "busy-overlay visible")
		busyEl.Set("aria-hidden", "false")
	} else {
		busyEl.Set("className", "busy-overlay")
		busyEl.Set("aria-hidden", "true")
	}
}

func runBusy(label string, fn func()) {
	setBusy(label, true)
	afterPaint(func() {
		defer setBusy("", false)
		fn()
	})
}

func runMaybeBusy(label string, show bool, fn func()) {
	if show {
		runBusy(label, fn)
		return
	}
	fn()
}

func afterPaint(fn func()) {
	raf := js.Global().Get("requestAnimationFrame")
	if !raf.Truthy() {
		var cb js.Func
		cb = js.FuncOf(func(js.Value, []js.Value) any {
			defer cb.Release()
			fn()
			return nil
		})
		js.Global().Call("setTimeout", cb, 75)
		return
	}

	var first js.Func
	first = js.FuncOf(func(js.Value, []js.Value) any {
		defer first.Release()
		var second js.Func
		second = js.FuncOf(func(js.Value, []js.Value) any {
			defer second.Release()
			fn()
			return nil
		})
		raf.Invoke(second)
		return nil
	})
	raf.Invoke(first)
}

func navigateTo(route webRoute) {
	if route.tab == "" {
		route.tab = "home"
	}
	currentRoute = route
	activeTab = route.tab
	hash := routeHash(route)
	if js.Global().Get("location").Get("hash").String() != hash {
		updatingRoute = true
		js.Global().Get("location").Set("hash", strings.TrimPrefix(hash, "#"))
		updatingRoute = false
	}
	renderActiveTab()
}

func replaceRoute(route webRoute) {
	if route.tab == "" {
		route.tab = activeTab
	}
	currentRoute = route
	history := js.Global().Get("history")
	if history.Truthy() {
		history.Call("replaceState", nil, "", routeHash(route))
		return
	}
	js.Global().Get("location").Set("hash", strings.TrimPrefix(routeHash(route), "#"))
}

func applyRouteFromHash() {
	if updatingRoute {
		return
	}
	route := parseRoute()
	currentRoute = route
	if route.chain != "" {
		loadChainFromRoute(route)
	}
	if route.tab == "" {
		route.tab = "home"
	}
	activeTab = route.tab
	renderActiveTab()
}

func applyCurrentRouteToTab() {
	switch activeTab {
	case "examples":
		if applyExamplesRoute != nil {
			applyExamplesRoute(currentRoute)
		}
	case "plugins":
		if applyPluginsRoute != nil {
			applyPluginsRoute(currentRoute)
		}
	}
}

func parseRoute() webRoute {
	hash := strings.TrimPrefix(js.Global().Get("location").Get("hash").String(), "#")
	if hash == "" {
		return webRoute{tab: "home"}
	}
	if strings.HasPrefix(hash, "chain=") {
		return webRoute{tab: "home", chain: strings.TrimPrefix(hash, "chain=")}
	}
	path, rawQuery, _ := strings.Cut(hash, "?")
	values, _ := url.ParseQuery(rawQuery)
	path = strings.Trim(path, "/")
	if path == "" {
		return webRoute{tab: "home", search: values.Get("search"), chain: values.Get("chain")}
	}
	parts := strings.Split(path, "/")
	route := webRoute{
		tab:    routeTab(parts[0]),
		search: values.Get("search"),
		chain:  values.Get("chain"),
	}
	if len(parts) > 1 {
		item, err := url.PathUnescape(strings.Join(parts[1:], "/"))
		if err == nil {
			route.item = item
		}
	}
	return route
}

func routeTab(tab string) string {
	switch strings.ToLower(strings.TrimSpace(tab)) {
	case "home", "examples", "plugins", "about":
		return strings.ToLower(strings.TrimSpace(tab))
	default:
		return "home"
	}
}

func routeHash(route webRoute) string {
	if route.chain != "" && route.tab == "home" && route.item == "" && route.search == "" {
		return "#chain=" + route.chain
	}
	tab := routeTab(route.tab)
	hash := "#" + tab
	if route.item != "" {
		hash += "/" + url.PathEscape(route.item)
	}
	params := url.Values{}
	if route.search != "" {
		params.Set("search", route.search)
	}
	if route.chain != "" {
		params.Set("chain", route.chain)
	}
	if encoded := params.Encode(); encoded != "" {
		hash += "?" + encoded
	}
	return hash
}

func routeURL(route webRoute) string {
	loc := js.Global().Get("location")
	return loc.Get("origin").String() +
		loc.Get("pathname").String() +
		loc.Get("search").String() +
		routeHash(route)
}

func copyRouteLink(route webRoute, label string) {
	link := routeURL(route)
	clipboard := js.Global().Get("navigator").Get("clipboard")
	if clipboard.Truthy() {
		clipboard.Call("writeText", link)
		alert(label + " link copied.")
		return
	}
	js.Global().Get("location").Set("hash", strings.TrimPrefix(routeHash(route), "#"))
	js.Global().Call("prompt", label+" link:", link)
}

func itemSlug(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func highlightRouteTarget(card js.Value) {
	card.Get("classList").Call("add", "route-highlight")
	card.Call("scrollIntoView", map[string]any{"block": "center", "behavior": "smooth"})
}

func menuBar() js.Value {
	bar := div("menu-bar")
	appendChildren(bar,
		menu("File",
			menuFilePicker("folder", "Open file"),
			menuItem("download", "Download result", downloadResult),
		),
		menu("Chain",
			menuChainPicker("upload", "Import chain"),
			menuItem("download", "Export chain", exportChain),
			menuItem("link", "Copy link", copyShareLink),
			menuItem("terminal", "Copy command", copyCommand),
		),
		menu("Workflow",
			menuItem("star", "Presets", showPresets),
			menuItem("compare", "Compare", showCompare),
			menuItem("undo", "Undo", func() {
				if pipe.Undo() {
					rebuild()
				}
			}),
			menuItem("redo", "Redo", func() {
				if pipe.Redo() {
					rebuild()
				}
			}),
			menuItem("clear", "Clear", func() {
				sourceName = ""
				clearSourceFullViews()
				pipe.Clear()
				rebuild()
			}),
		),
	)
	return bar
}

func tabButton(tab, label string) js.Value {
	b := el("button")
	b.Set("type", "button")
	b.Set("className", "tab")
	b.Set("textContent", label)
	tabBtns = append(tabBtns, b)
	onStatic(b, "click", func() {
		navigateTo(webRoute{tab: tab})
	})
	return b
}

func menu(label string, items ...js.Value) js.Value {
	wrap := div("menu")
	trigger := el("button")
	trigger.Set("type", "button")
	trigger.Set("className", "menu-trigger")
	appendChildren(trigger, textNode(label), iconGraphic("chevron-down"))
	dropdown := div("menu-dropdown")
	dropdown.Set("role", "menu")
	appendChildren(dropdown, items...)
	appendChildren(wrap, trigger, dropdown)
	return wrap
}

func menuItem(icon, label string, fn func()) js.Value {
	b := el("button")
	b.Set("type", "button")
	b.Set("className", "menu-item")
	b.Set("role", "menuitem")
	appendChildren(b, iconGraphic(icon), textNode(label))
	if fn != nil {
		on(b, "click", fn)
	}
	return b
}

func menuTabItem(icon, label, tab string) js.Value {
	b := menuItem(icon, label, func() {
		navigateTo(webRoute{tab: tab})
	})
	b.Call("setAttribute", "data-tab", tab)
	return b
}

func menuFilePicker(icon, label string) js.Value {
	wrap := div("menu-file-item")
	item := menuItem(icon, label, nil)
	input := el("input")
	input.Set("type", "file")
	input.Get("style").Set("display", "none")
	on(item, "click", func() { input.Call("click") })
	on(input, "change", func() {
		files := input.Get("files")
		if files.Get("length").Int() == 0 {
			return
		}
		loadSourceFile(files.Call("item", 0))
	})
	appendChildren(wrap, item, input)
	return wrap
}

func menuChainPicker(icon, label string) js.Value {
	wrap := div("menu-file-item")
	item := menuItem(icon, label, nil)
	input := el("input")
	input.Set("type", "file")
	input.Set("accept", "application/json,.json")
	input.Get("style").Set("display", "none")
	on(item, "click", func() { input.Call("click") })
	on(input, "change", func() {
		files := input.Get("files")
		if files.Get("length").Int() == 0 {
			return
		}
		loadChainFile(files.Call("item", 0))
	})
	appendChildren(wrap, item, input)
	return wrap
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
	applyCurrentRouteToTab()
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

	stepsEl = div("steps")

	stepsEl.Call("appendChild", menuBar())
	stepsEl.Call("appendChild", sourceCard())
	historyEl = div("history chain-overview")
	stepsEl.Call("appendChild", historyEl)
	for i := range pipe.Steps() {
		stepsEl.Call("appendChild", stepCard(i))
	}
	stepsEl.Call("appendChild", addCard())

	appendChildren(main, stepsEl)
	appendChildren(app, main)
	contentEl.Call("appendChild", app)
	updateHistory()
	autoSizeSourceTextareasSoon()
}

func commandBar() js.Value {
	bar := div("command-bar")
	appendChildren(bar,
		commandGroup("Result",
			iconButton("primary", "copy", "Copy result", copyResult),
			iconButton("", "download", "Download", downloadResult),
			filePicker(),
		),
		commandGroup("Chain",
			chainPicker(),
			iconButton("", "upload", "Export chain", exportChain),
			iconButton("", "link", "Copy link", copyShareLink),
			iconButton("", "terminal", "Copy command", copyCommand),
		),
		commandGroup("Workflow",
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
	return bar
}

func commandGroup(title string, kids ...js.Value) js.Value {
	group := div("command-group")
	label := el("div")
	label.Set("className", "command-title")
	label.Set("textContent", title)
	appendChildren(group, label)
	for _, kid := range kids {
		group.Call("appendChild", kid)
	}
	return group
}

func renderAbout() {
	if aboutPageBuilt {
		contentEl.Call("appendChild", aboutPage)
		return
	}
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
	built.Set("textContent", "This web interface runs deen as WebAssembly in your browser. Transform chains process data client side, so pasted text and loaded files do not need to be sent back to the server.")
	privacy := el("p")
	privacy.Set("textContent", "That client-side model is privacy focused for sensitive payloads: you can inspect tokens, logs, binaries, certificates and encoded data without handing the raw content to a hosted decoding service.")
	local := el("p")
	local.Set("textContent", "For fully local workflows, run the deen web server on your own machine, or use the native GUI and CLI builds. The CLI is also useful for repeatable processing, shell pipelines and workflow automation.")
	docs := link("Documentation", "https://deen.adversec.com")
	repo := link("Source", "https://github.com/takeshixx/go-deen")
	appendChildren(page, h, desc, versionEl, built, privacy, local, docs, repo)
	aboutPage = page
	aboutPageBuilt = true
	contentEl.Call("appendChild", page)
}

func renderExamples() {
	if examplesPageBuilt {
		contentEl.Call("appendChild", examplesPage)
		return
	}
	page := div("catalog examples-page")
	header := div("examples-header")
	title := el("h2")
	title.Set("textContent", "Examples")
	query := textInput("Search examples", "")
	appendChildren(header, title, query)
	results := div("examples-grid")
	appendChildren(page, header, results)

	refs := make([]exampleCardRef, 0, len(pipeline.BuiltinExamples()))
	empty := div("empty")
	empty.Set("textContent", "No examples found.")
	empty.Get("style").Set("display", "none")
	for _, example := range pipeline.BuiltinExamples() {
		card := exampleCard(example)
		refs = append(refs, exampleCardRef{example: example, card: card})
		results.Call("appendChild", card)
	}
	results.Call("appendChild", empty)

	render := func() {
		q := query.Get("value").String()
		matches := 0
		for _, ref := range refs {
			if !pipeline.ExampleMatches(ref.example, q) {
				ref.card.Get("style").Set("display", "none")
				continue
			}
			ref.card.Get("style").Set("display", "")
			matches++
		}
		if matches == 0 {
			empty.Get("style").Set("display", "")
		} else {
			empty.Get("style").Set("display", "none")
		}
	}
	applyExamplesRoute = func(route webRoute) {
		query.Set("value", route.search)
		render()
		for _, ref := range refs {
			ref.card.Get("classList").Call("remove", "route-highlight")
		}
		if route.item == "" {
			return
		}
		for _, ref := range refs {
			if routeItemMatches(route.item, ref.example.Name) {
				ref.card.Set("open", true)
				highlightRouteTarget(ref.card)
				return
			}
		}
	}
	onStatic(query, "input", func() {
		render()
		replaceRoute(webRoute{tab: "examples", search: query.Get("value").String()})
	})
	render()
	examplesPage = page
	examplesPageBuilt = true
	contentEl.Call("appendChild", page)
}

func exampleCard(example pipeline.Example) js.Value {
	card := el("details")
	card.Set("className", "plugin-card example-card")
	summary := el("summary")
	title := el("h3")
	title.Set("textContent", example.Name)
	summary.Call("appendChild", title)
	card.Call("appendChild", summary)

	detailsLoaded := false
	onStatic(card, "toggle", func() {
		if detailsLoaded || !card.Get("open").Bool() {
			return
		}
		detailsLoaded = true
		populateExampleDetails(card, example)
	})

	return card
}

func populateExampleDetails(card js.Value, example pipeline.Example) {
	desc := el("p")
	desc.Set("textContent", example.Description)
	source := div("plugin-meta")
	source.Set("textContent", "Input: "+pipeline.DataMetadata(example.Source, 0).Summary())
	chain := exampleChain(example.Steps)

	previewSlot := div("example-preview-slot")
	previewHint := div("plugin-meta")
	previewHint.Set("textContent", "Preview data is loaded on demand.")
	previewSlot.Call("appendChild", previewHint)

	actions := div("modal-actions")
	previewLoaded := false
	preview := el("button")
	preview.Set("type", "button")
	preview.Set("textContent", "Preview data")
	onStatic(preview, "click", func() {
		if previewLoaded {
			return
		}
		previewLoaded = true
		preview.Set("disabled", true)
		runBusy("Previewing example", func() {
			renderExamplePreview(example, previewSlot)
		})
	})
	load := el("button")
	load.Set("type", "button")
	load.Set("textContent", "Load example")
	onStatic(load, "click", func() {
		runBusy("Loading example", func() {
			sourceName = ""
			pipe.ApplyExample(example)
			activeTab = "home"
			renderActiveTab()
		})
	})
	copyLink := el("button")
	copyLink.Set("type", "button")
	copyLink.Set("textContent", "Copy link")
	onStatic(copyLink, "click", func() {
		copyRouteLink(webRoute{tab: "examples", item: itemSlug(example.Name)}, "Example")
	})
	appendChildren(actions, preview, load, copyLink)
	appendChildren(card, desc, source, chain)
	if example.WantContains != "" {
		want := div("plugin-meta")
		want.Set("textContent", "Expected result contains: "+example.WantContains)
		card.Call("appendChild", want)
	}
	appendChildren(card, previewSlot, actions)
}

func renderExamplePreview(example pipeline.Example, slot js.Value) {
	slot.Set("innerHTML", "")
	preview := cachedExamplePreview(example)
	if preview.err != nil {
		output := div("error")
		output.Set("textContent", "Output error: "+preview.err.Error())
		slot.Call("appendChild", output)
		return
	}
	outputSummary := div("plugin-meta")
	outputSummary.Set("textContent", "Output: "+pipeline.DataMetadata(preview.result, len(example.Source)).Summary())
	data := div("example-data-grid")
	inputPanel := exampleDataPanel("Input data", example.Source)
	outputPanel := exampleDataPanel("Output result", preview.result)
	appendChildren(data, inputPanel, outputPanel)
	appendChildren(slot, outputSummary, data)
}

func cachedExamplePreview(example pipeline.Example) examplePreview {
	key := example.Name
	if preview, ok := examplePreviews[key]; ok {
		return preview
	}
	result, err := pipeline.ExampleResult(example)
	preview := examplePreview{result: result, err: err}
	examplePreviews[key] = preview
	return preview
}

func exampleDataPanel(title string, data []byte) js.Value {
	panel := div("example-data-panel")
	label := el("strong")
	label.Set("textContent", title)
	if imageData, mime := imagePreviewPayload(data); mime != "" {
		wrap, img, msg := imagePreviewBox()
		img.Set("src", "data:"+mime+";base64,"+base64.StdEncoding.EncodeToString(imageData))
		img.Get("style").Set("display", "block")
		msg.Set("textContent", mime)
		appendChildren(panel, label, wrap)
		return panel
	}
	body := el("pre")
	body.Set("className", "syntax-preview example-data-code")
	if preview, spans, ok := pipeline.HighlightedPreview(data); ok {
		renderHighlightedText(body, preview, spans)
	} else {
		renderHighlightedText(body, exampleDataText(data), nil)
	}
	appendChildren(panel, label, body)
	return panel
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
			arrow.Set("textContent", "→")
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
	if pluginsPageBuilt {
		contentEl.Call("appendChild", pluginsPage)
		return
	}
	page := div("catalog")
	header := div("examples-header")
	title := el("h2")
	title.Set("textContent", "Plugins")
	query := textInput("Search plugins", "")
	appendChildren(header, title, query)
	page.Call("appendChild", header)
	current := ""
	refs := make([]pluginCardRef, 0, len(plugins.UICatalog()))
	for _, info := range plugins.UICatalog() {
		if info.Category != current {
			current = info.Category
			h := el("h2")
			h.Set("textContent", plugins.CategoryLabel(current))
			page.Call("appendChild", h)
		}
		card := pluginCard(info)
		refs = append(refs, pluginCardRef{info: info, card: card})
		page.Call("appendChild", card)
	}
	render := func() {
		q := query.Get("value").String()
		for _, ref := range refs {
			if pluginMatches(ref.info, q) {
				ref.card.Get("style").Set("display", "")
			} else {
				ref.card.Get("style").Set("display", "none")
			}
		}
	}
	applyPluginsRoute = func(route webRoute) {
		query.Set("value", route.search)
		render()
		for _, ref := range refs {
			ref.card.Get("classList").Call("remove", "route-highlight")
		}
		if route.item == "" {
			return
		}
		for _, ref := range refs {
			if routeItemMatches(route.item, ref.info.Name) || routeItemMatches(route.item, ref.info.Label) {
				ref.card.Get("style").Set("display", "")
				highlightRouteTarget(ref.card)
				return
			}
		}
	}
	onStatic(query, "input", func() {
		render()
		replaceRoute(webRoute{tab: "plugins", search: query.Get("value").String()})
	})
	render()
	pluginsPage = page
	pluginsPageBuilt = true
	contentEl.Call("appendChild", page)
}

func pluginMatches(info plugins.UIPluginInfo, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	fields := []string{
		info.Name,
		info.Label,
		info.Category,
		plugins.CategoryLabel(info.Category),
		info.Description,
		info.UseFor,
		strings.Join(info.Aliases, " "),
	}
	return strings.Contains(strings.ToLower(strings.Join(fields, " ")), query)
}

func routeItemMatches(item, name string) bool {
	item = strings.TrimSpace(item)
	return strings.EqualFold(item, name) || itemSlug(item) == itemSlug(name)
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
	actions := div("modal-actions")
	copyLink := el("button")
	copyLink.Set("type", "button")
	copyLink.Set("textContent", "Copy link")
	onStatic(copyLink, "click", func() {
		copyRouteLink(webRoute{tab: "plugins", item: itemSlug(info.Name)}, "Plugin")
	})
	actions.Call("appendChild", copyLink)
	appendChildren(card, title, meta, desc, use, actions)

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
			if close != nil {
				close()
			}
			runBusy("Processing", func() {
				pipe.AddStepWithOptions(s.Plugin, s.Unprocess, s.Options)
				rebuild()
			})
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
				if close != nil {
					close()
				}
				runBusy("Processing", func() {
					pipe.AddStep(info.Name, false)
					rebuild()
				})
			})
			actions.Call("appendChild", add)
			if info.CanDecode {
				decode := el("button")
				decode.Set("type", "button")
				decode.Set("textContent", "Add decode")
				on(decode, "click", func() {
					if close != nil {
						close()
					}
					runBusy("Processing", func() {
						pipe.AddStep(info.Name, true)
						rebuild()
					})
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
			if close != nil {
				close()
			}
			runBusy("Processing", func() {
				pipe.ApplyPreset(p)
				rebuild()
			})
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
		text, _ := pipeline.HexDisplay(data)
		return text
	case "base64":
		if pipeline.IsLargeData(data) {
			return base64.StdEncoding.EncodeToString(data[:min(len(data), pipeline.HexPreviewLimit)]) + "\n\n... base64 preview truncated ..."
		}
		return base64.StdEncoding.EncodeToString(data)
	default:
		text, _ := pipeline.TextDisplay(data)
		return text
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

func loadChainFromRoute(route webRoute) {
	if route.chain == "" {
		return
	}
	data, err := base64.RawURLEncoding.DecodeString(route.chain)
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
		name := file.Get("name").String()
		runBusy("Processing file", func() {
			sourceName = name
			clearSourceFullViews()
			pipe.SetSourceOwned(buf)
			rebuild()
		})
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

func loadChainFile(file js.Value) {
	reader := js.Global().Get("FileReader").New()
	var loadCB js.Func
	loadCB = js.FuncOf(func(js.Value, []js.Value) any {
		data := []byte(reader.Get("result").String())
		runBusy("Importing chain", func() {
			if err := pipe.ImportJSON(data); err != nil {
				alert(err.Error())
			} else {
				sourceName = ""
				clearSourceFullViews()
				rebuild()
			}
		})
		loadCB.Release()
		return nil
	})
	reader.Call("addEventListener", "load", loadCB)
	reader.Call("readAsText", file)
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
		loadChainFile(files.Call("item", 0))
	})
	appendChildren(wrap, openBtn, input)
	return wrap
}

func sourceCard() js.Value {
	card := div("card source")
	label := el("div")
	label.Set("className", "card-title")
	label.Set("textContent", "Input")

	rawNeedsFull, hexNeedsFull, stringsNeedsFull := sourceNeedsFull()
	if !rawNeedsFull {
		sourceFullRaw = false
	}
	if !hexNeedsFull {
		sourceFullHex = false
	}
	if !stringsNeedsFull {
		sourceFullStrings = false
	}
	sourceText, textCapped := webTextDisplay(pipe.Source(), sourceFullRaw)
	ta := sourceTextarea(sourceText)
	if textCapped || sourceFullRaw {
		ta.Set("readOnly", true)
		ta.Set("title", pipeline.LargeDataPlaceholder(pipe.Source()))
	}
	sourceEl = ta
	var sourceHex js.Value
	var sourceStrings js.Value
	sourceHasHex := false
	var sourceView js.Value = ta
	if pipeline.IsBinaryData(pipe.Source()) {
		sourceHasHex = true
		sourceHex = sourceTextarea("")
		sourceHex.Set("readOnly", true)
		hexText, _ := webHexDisplay(pipe.Source(), sourceFullHex)
		sourceHex.Set("value", hexText)
		sourceStrings = sourceTextarea("")
		sourceStrings.Set("readOnly", true)
		stringsText, _ := webStringsDisplay(pipe.Source(), sourceFullStrings)
		sourceStrings.Set("value", stringsText)
		sourceView = sourceInputViewer(ta, sourceHex, sourceStrings, "hex")
	}
	fullControls := sourceFullViewControls(rawNeedsFull, hexNeedsFull, stringsNeedsFull)
	meta := div("meta")
	renderMetadata(meta, sourceName, pipeline.DataMetadata(pipe.Source(), 0))
	on(ta, "input", func() {
		if updating {
			return
		}
		value := ta.Get("value").String()
		runMaybeBusy("Processing", pipe.Len() > 0, func() {
			sourceName = ""
			clearSourceFullViews()
			pipe.SetSourceOwned([]byte(value))
			renderMetadata(meta, sourceName, pipeline.DataMetadata(pipe.Source(), 0))
			if sourceHasHex {
				hexText, _ := webHexDisplay(pipe.Source(), false)
				sourceHex.Set("value", hexText)
				autoSizeTextareaSoon(sourceHex)
				stringsText, _ := webStringsDisplay(pipe.Source(), false)
				sourceStrings.Set("value", stringsText)
				autoSizeTextareaSoon(sourceStrings)
			}
			refreshOutputs(0)
		})
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
	appendChildren(card, label, sourceView, fullControls, meta)
	autoSizeTextarea(ta)
	if sourceHasHex {
		autoSizeTextarea(sourceHex)
		autoSizeTextarea(sourceStrings)
	}
	return card
}

func sourceNeedsFull() (raw, hexView, stringsView bool) {
	_, raw = webTextDisplay(pipe.Source(), false)
	_, hexView = webHexDisplay(pipe.Source(), false)
	_, stringsView = webStringsDisplay(pipe.Source(), false)
	return raw, hexView, stringsView
}

func clearSourceFullViews() {
	sourceFullRaw = false
	sourceFullHex = false
	sourceFullStrings = false
}

func sourceFullViewControls(rawCapped, hexCapped, stringsCapped bool) js.Value {
	controls := div("full-view-controls")
	add := func(label string, visible bool, setFull func()) {
		if !visible {
			return
		}
		controls.Call("appendChild", button("", label, func() {
			ok := js.Global().Call("confirm", label+"?\n\nRendering the full input view can use a lot of memory and may make the interface slow for large files.").Bool()
			if !ok {
				return
			}
			setFull()
			rebuild()
		}))
	}
	add("Show full Raw", rawCapped && !sourceFullRaw, func() { sourceFullRaw = true })
	if pipeline.IsBinaryData(pipe.Source()) {
		add("Show full Hex", hexCapped && !sourceFullHex, func() { sourceFullHex = true })
		add("Show full Strings", stringsCapped && !sourceFullStrings, func() { sourceFullStrings = true })
	}
	if sourceFullRaw || sourceFullHex || sourceFullStrings {
		notice := el("span")
		notice.Set("className", "full-view-warning")
		notice.Set("textContent", "Full input view enabled; input is read-only.")
		controls.Call("appendChild", notice)
	}
	if controls.Get("childNodes").Get("length").Int() == 0 {
		controls.Get("style").Set("display", "none")
	}
	return controls
}

func sourceInputViewer(textInput, hexInput, stringsInput js.Value, activeView string) js.Value {
	if activeView == "" {
		activeView = "raw"
	}
	viewer := div("viewer")
	tabs := div("viewer-tabs")
	panels := div("viewer-panels")
	items := []struct {
		label string
		name  string
		node  js.Value
	}{
		{"Raw", "raw", textInput},
		{"Hex", "hex", hexInput},
		{"Strings", "strings", stringsInput},
	}
	buttons := make([]js.Value, 0, len(items))
	panelEls := make([]js.Value, 0, len(items))
	var activate func(int)
	for idx, item := range items {
		panel := div("viewer-panel")
		if item.name != activeView {
			panel.Set("className", "viewer-panel hidden")
		}
		panel.Call("appendChild", item.node)
		panelEls = append(panelEls, panel)
		btn := el("button")
		btn.Set("type", "button")
		if item.name == activeView {
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
		for i, panel := range panelEls {
			if i == active {
				panel.Set("className", "viewer-panel")
				buttons[i].Set("className", "viewer-tab active")
			} else {
				panel.Set("className", "viewer-panel hidden")
				buttons[i].Set("className", "viewer-tab")
			}
		}
		autoSizeTextareaSoon(textInput)
		autoSizeTextareaSoon(hexInput)
		autoSizeTextareaSoon(stringsInput)
	}
	appendChildren(viewer, tabs, panels)
	return viewer
}

func sourceMetadataSummary() string {
	summary := pipeline.DataMetadata(pipe.Source(), 0).Summary()
	if strings.TrimSpace(sourceName) == "" {
		return summary
	}
	return "source: " + sourceName + " · " + summary
}

func renderMetadata(node js.Value, source string, meta pipeline.Metadata) {
	node.Set("innerHTML", "")
	if strings.TrimSpace(source) != "" {
		item := div("meta-item meta-source")
		label := el("span")
		label.Set("className", "meta-label")
		label.Set("textContent", "source")
		value := el("span")
		value.Set("className", "meta-value")
		value.Set("textContent", source)
		appendChildren(item, label, value)
		node.Call("appendChild", item)
	}
	for _, field := range meta.Fields() {
		item := div("meta-item")
		label := el("span")
		label.Set("className", "meta-label")
		label.Set("textContent", field.Label)
		value := el("span")
		value.Set("className", "meta-value")
		value.Set("textContent", field.Value)
		appendChildren(item, label, value)
		node.Call("appendChild", item)
	}
}

func stepCard(i int) js.Value {
	step := pipe.Steps()[i]
	col := accent(i)
	ref := &cardRef{index: i}
	cards = append(cards, ref)

	card := div("card")
	if step.Disabled {
		card.Set("className", "card disabled")
	}
	card.Get("style").Set("borderLeft", "5px solid "+col)

	// Header: collapse, title, summary, controls.
	header := div("card-header")
	collapse := iconOnlyButton("collapse", "Collapse step", nil)
	title := el("span")
	title.Set("className", "title")
	title.Set("textContent", fmt.Sprintf("Step %d", i+1))
	title.Get("style").Set("color", col)
	summary := el("span")
	summary.Set("className", "summary")
	summary.Set("textContent", summaryText(i))
	remove := iconOnlyButton("trash", "Delete step", func() {
		runBusy("Processing", func() {
			pipe.RemoveStep(i)
			rebuild()
		})
	})
	moveUp := iconOnlyButton("arrow-up", "Move step up", func() {
		runBusy("Processing", func() {
			pipe.MoveStep(i, i-1)
			rebuild()
		})
	})
	moveDown := iconOnlyButton("arrow-down", "Move step down", func() {
		runBusy("Processing", func() {
			pipe.MoveStep(i, i+1)
			rebuild()
		})
	})
	duplicate := iconOnlyButton("copy", "Duplicate step", func() {
		runBusy("Processing", func() {
			pipe.DuplicateStep(i)
			rebuild()
		})
	})
	enableLabel := "Disable step"
	enableIcon := "visible"
	if step.Disabled {
		enableLabel = "Enable step"
		enableIcon = "hidden"
	}
	enabled := iconOnlyButton(enableIcon, enableLabel, func() {
		runBusy("Processing", func() {
			pipe.SetStepDisabled(i, !step.Disabled)
			rebuild()
		})
	})
	enabled.Set("aria-pressed", fmt.Sprintf("%t", !step.Disabled))

	spacer := div("spacer")
	appendChildren(header, collapse, title, summary, spacer, enabled, moveUp, moveDown, duplicate, remove)

	detail := div("card-detail")

	// decode toggle (referenced by the category selectors).
	canDecode := plugins.CanDecode(step.Plugin)
	decodeWrap, decodeInput := checkbox("decode", step.Unprocess && canDecode)
	styleStepToggle(decodeWrap, decodeInput, "mode-toggle", "Decode mode", step.Unprocess && canDecode)
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
			runBusy("Processing", func() {
				pipe.SetPlugin(i, name, decode)
				rebuild()
			})
		})
		selRow.Call("appendChild", sel)
	}

	on(decodeInput, "change", func() {
		decode := decodeInput.Get("checked").Bool() && plugins.CanDecode(step.Plugin)
		styleStepToggle(decodeWrap, decodeInput, "mode-toggle", "Decode mode", decode)
		runBusy("Processing", func() {
			pipe.SetPlugin(i, step.Plugin, decode)
			refreshOutputs(i)
		})
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
	ref.stringsOutput = textareaWithMax("", 960)
	ref.stringsOutput.Set("readOnly", true)
	ref.preview = previewBox()
	ref.imageEnabled = stepGeneratesImage(step)
	if ref.imageEnabled {
		ref.imagePanel, ref.image, ref.imageMsg = imagePreviewBox()
	}
	viewer := outputViewer(ref)

	ref.fullControls = div("full-view-controls")
	buildFullViewControls(ref)
	ref.meta = div("meta")
	ref.errEl = div("error")
	ref.errEl.Get("style").Set("display", "none")

	appendChildren(detail, selRow, toggles, options, viewer, ref.fullControls, ref.meta, ref.errEl)
	on(collapse, "click", func() {
		if detail.Get("style").Get("display").String() == "none" {
			detail.Get("style").Set("display", "block")
			setIconOnlyButtonContent(collapse, "collapse", "Collapse step")
			autoSizeOutputTextareas(ref)
		} else {
			detail.Get("style").Set("display", "none")
			setIconOnlyButtonContent(collapse, "expand", "Expand step")
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

	type panelItem struct {
		label string
		name  string
		node  js.Value
	}
	panelItems := []panelItem{
		{"Raw", "raw", ref.output},
		{"Hex", "hex", ref.hexOutput},
		{"Strings", "strings", ref.stringsOutput},
		{"Preview", "preview", ref.preview},
	}
	if ref.imageEnabled {
		panelItems = append(panelItems, panelItem{"Image", "image", ref.imagePanel})
	}
	activeView := "raw"
	if pipeline.IsBinaryData(pipe.Output(ref.index)) {
		activeView = "hex"
	} else if ref.imageEnabled {
		activeView = "image"
	}
	buttons := map[string]js.Value{}
	panelEls := map[string]js.Value{}
	hasPreview := pipeline.HasStructuredPreview(pipe.Output(ref.index))
	hasStrings := pipeline.IsBinaryData(pipe.Output(ref.index))
	if hasPreview {
		activeView = "preview"
	}
	ref.activeOutputView = activeView
	var activate func(string)
	for _, item := range panelItems {
		panel := div("viewer-panel")
		if item.name != activeView {
			panel.Set("className", "viewer-panel hidden")
		}
		panel.Call("appendChild", item.node)
		btn := el("button")
		btn.Set("type", "button")
		if item.name == activeView {
			btn.Set("className", "viewer-tab active")
		} else {
			btn.Set("className", "viewer-tab")
		}
		btn.Set("textContent", item.label)
		name := item.name
		on(btn, "click", func() { activate(name) })
		buttons[name] = btn
		panelEls[name] = panel
		switch item.name {
		case "raw":
			ref.rawButton, ref.rawPanel = btn, panel
		case "hex":
			ref.hexButton, ref.hexPanel = btn, panel
		case "strings":
			ref.stringsButton, ref.stringsPanel = btn, panel
			setOutputTabVisible(btn, panel, hasStrings)
		case "preview":
			ref.previewButton, ref.previewPanel = btn, panel
			setOutputTabVisible(btn, panel, hasPreview)
		case "image":
			ref.imageButton = btn
		}
		tabs.Call("appendChild", btn)
		panels.Call("appendChild", panel)
	}
	activate = func(active string) {
		if active == "strings" && !pipeline.IsBinaryData(pipe.Output(ref.index)) {
			return
		}
		if active == "preview" && !pipeline.HasStructuredPreview(pipe.Output(ref.index)) {
			return
		}
		for name, btn := range buttons {
			if name == active {
				btn.Set("className", "viewer-tab active")
				panelEls[name].Set("className", "viewer-panel")
				ref.activeOutputView = name
				autoSizeViewerPanel(panelEls[name])
			} else {
				btn.Set("className", "viewer-tab")
				panelEls[name].Set("className", "viewer-panel hidden")
			}
		}
	}
	ref.activateOutput = activate

	appendChildren(viewer, tabs, panels)
	return viewer
}

func setOutputTabVisible(button, panel js.Value, visible bool) {
	if visible {
		button.Get("style").Set("display", "")
		return
	}
	button.Get("style").Set("display", "none")
	panel.Set("className", "viewer-panel hidden")
}

func styleStepToggle(wrap, input js.Value, baseClass, title string, active bool) {
	className := baseClass + " step-toggle"
	if active {
		className += " active"
	}
	wrap.Set("className", className)
	wrap.Set("title", title)
	input.Set("checked", active)
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
				runBusy("Processing", func() {
					pipe.SetOption(i, opt.Name, val)
					refreshOutputs(i)
				})
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
					runBusy("Processing", func() {
						pipe.SetOption(i, opt.Name, input.Get("value").String())
						refreshOutputs(i)
					})
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
			runBusy("Processing", func() {
				pipe.AddStep(name, false)
				rebuild()
			})
		})
		appendChildren(group, pickerLabel, sel)
		selRow.Call("appendChild", group)
	}
	appendChildren(card, header, selRow)
	return card
}

func updateHistory() {
	historyEl.Set("innerHTML", "")
	historyEl.Get("style").Set("display", "none")
	shown := 0
	for i, s := range pipe.Steps() {
		if s.Disabled {
			continue
		}
		if shown > 0 {
			arrow := el("span")
			arrow.Set("className", "chain-arrow")
			arrow.Set("textContent", "→")
			historyEl.Call("appendChild", arrow)
		}
		historyEl.Call("appendChild", liveChainStepPill(i, s))
		shown++
	}
	if shown > 0 {
		historyEl.Get("style").Set("display", "flex")
	}
}

func liveChainStepPill(i int, step *pipeline.Step) js.Value {
	pill := div("chain-pill")
	if step.Disabled {
		pill.Set("className", "chain-pill disabled")
	}
	pill.Get("style").Set("borderColor", accent(i))
	name := plugins.PluginLabel(step.Plugin)
	if name == "" {
		name = "(none)"
	}
	if step.Unprocess {
		name = "." + name
	}
	title := el("span")
	title.Set("className", "chain-plugin")
	title.Set("textContent", name)
	pill.Call("appendChild", title)

	metaParts := make([]string, 0, len(step.Options))
	for k, v := range step.Options {
		metaParts = append(metaParts, k+"="+v)
	}
	if len(metaParts) > 0 {
		sort.Strings(metaParts)
		meta := el("span")
		meta.Set("className", "chain-options")
		meta.Set("textContent", strings.Join(metaParts, ", "))
		pill.Call("appendChild", meta)
	}
	return pill
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
	inputBytes := len(pipe.Input(c.index))
	renderMetadata(c.meta, "", pipeline.DataMetadata(out, inputBytes))
	if err := pipe.Err(c.index); err != nil {
		c.errEl.Set("textContent", "error: "+err.Error())
		c.errEl.Get("style").Set("display", "block")
	} else {
		c.errEl.Get("style").Set("display", "none")
	}
	updating = true
	_, rawNeedsFull := webTextDisplay(out, false)
	_, hexNeedsFull := webHexDisplay(out, false)
	_, stringsNeedsFull := webStringsDisplay(out, false)
	if !rawNeedsFull {
		c.fullRaw = false
	}
	if !hexNeedsFull {
		c.fullHex = false
	}
	if !stringsNeedsFull {
		c.fullStrings = false
	}
	text, textCapped := webTextDisplay(out, c.fullRaw)
	c.output.Set("value", text)
	c.output.Set("readOnly", textCapped || c.fullRaw)
	if textCapped {
		c.output.Set("title", pipeline.LargeDataPlaceholder(out))
	} else {
		c.output.Set("title", "")
	}
	hexText, _ := webHexDisplay(out, c.fullHex)
	c.hexOutput.Set("value", hexText)
	stringsText, _ := webStringsDisplay(out, c.fullStrings)
	c.stringsOutput.Set("value", stringsText)
	autoSizeOutputTextareas(c)
	if c.imageEnabled {
		renderImageOutput(c, out)
	}
	hasPreview := pipeline.HasStructuredPreview(out)
	hasStrings := pipeline.IsBinaryData(out)
	setOutputTabVisible(c.stringsButton, c.stringsPanel, hasStrings)
	if !hasStrings && c.activeOutputView == "strings" && c.activateOutput != nil {
		c.activateOutput("raw")
	}
	renderFullViewControls(c, rawNeedsFull, hexNeedsFull, stringsNeedsFull)
	previewWasVisible := c.previewButton.Get("style").Get("display").String() != "none"
	setOutputTabVisible(c.previewButton, c.previewPanel, hasPreview)
	if hasPreview {
		preview, spans, _ := pipeline.HighlightedPreview(out)
		renderHighlightedText(c.preview, preview, spans)
		if !previewWasVisible && c.activateOutput != nil {
			c.activateOutput("preview")
		}
	} else if c.activeOutputView == "preview" && c.activateOutput != nil {
		if pipeline.IsBinaryData(out) {
			c.activateOutput("hex")
		} else {
			c.activateOutput("raw")
		}
	}
	updating = false
}

func webTextDisplay(data []byte, full bool) (string, bool) {
	if full {
		return pipeline.TextDisplayFull(data), true
	}
	return pipeline.TextDisplay(data)
}

func webHexDisplay(data []byte, full bool) (string, bool) {
	if full {
		return pipeline.HexDisplayFull(data), true
	}
	return pipeline.HexDisplay(data)
}

func webStringsDisplay(data []byte, full bool) (string, bool) {
	if full {
		return pipeline.StringsDisplayFull(data), true
	}
	return pipeline.StringsDisplay(data)
}

func renderFullViewControls(c *cardRef, rawCapped, hexCapped, stringsCapped bool) {
	setVisible := func(node js.Value, visible bool) {
		if visible {
			node.Get("style").Set("display", "")
		} else {
			node.Get("style").Set("display", "none")
		}
	}
	showRaw := rawCapped && !c.fullRaw
	showHex := hexCapped && !c.fullHex
	showStrings := stringsCapped && !c.fullStrings
	showWarning := c.fullRaw || c.fullHex || c.fullStrings
	setVisible(c.rawFullButton, showRaw)
	setVisible(c.hexFullButton, showHex)
	setVisible(c.stringsFullButton, showStrings)
	setVisible(c.fullWarning, showWarning)
	if !showRaw && !showHex && !showStrings && !showWarning {
		c.fullControls.Get("style").Set("display", "none")
	} else {
		c.fullControls.Get("style").Set("display", "flex")
	}
}

func buildFullViewControls(c *cardRef) {
	warnAndSet := func(label string, setFull func()) func() {
		return func() {
			ok := js.Global().Call("confirm", label+"?\n\nRendering the full view can use a lot of memory and may make the interface slow for large binary data.").Bool()
			if !ok {
				return
			}
			setFull()
			renderOutput(c)
		}
	}
	c.rawFullButton = button("", "Show full Raw", warnAndSet("Show full Raw", func() { c.fullRaw = true }))
	c.hexFullButton = button("", "Show full Hex", warnAndSet("Show full Hex", func() { c.fullHex = true }))
	c.stringsFullButton = button("", "Show full Strings", warnAndSet("Show full Strings", func() { c.fullStrings = true }))
	c.fullWarning = el("span")
	c.fullWarning.Set("className", "full-view-warning")
	c.fullWarning.Set("textContent", "Full view enabled; output is read-only.")
	appendChildren(c.fullControls, c.rawFullButton, c.hexFullButton, c.stringsFullButton, c.fullWarning)
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
