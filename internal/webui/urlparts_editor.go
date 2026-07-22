//go:build js && wasm

package webui

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"syscall/js"

	"github.com/takeshixx/deen/pkg/formatters"
)

type webURLPartsEditor struct {
	card        *cardRef
	root        js.Value
	message     js.Value
	rebuilt     js.Value
	copy        js.Value
	doc         *formatters.URLPartsDocument
	lastJSON    string
	showRaw     bool
	initialized bool
	callbacks   []js.Func
}

func newWebURLPartsEditor(card *cardRef, data []byte) *webURLPartsEditor {
	e := &webURLPartsEditor{
		card: card,
		root: div("urlparts-editor"),
	}
	e.root.Call("setAttribute", "data-testid", "urlparts-editor")
	e.refresh(data)
	return e
}

func (e *webURLPartsEditor) refresh(data []byte) {
	if e == nil || e.initialized && string(data) == e.lastJSON {
		return
	}
	e.initialized = true
	e.lastJSON = string(data)
	doc, err := formatters.DecodeURLPartsJSON(bytes.NewReader(data))
	e.releaseCallbacks()
	e.root.Set("innerHTML", "")
	if err != nil {
		e.doc = nil
		message := div("urlparts-message invalid")
		message.Set("role", "status")
		message.Set("textContent", "Structured editor unavailable: "+err.Error())
		e.root.Call("appendChild", message)
		return
	}
	e.doc = doc
	e.build()
}

func (e *webURLPartsEditor) build() {
	e.root.Set("innerHTML", "")
	e.message = div("urlparts-message")
	e.message.Set("role", "status")
	e.message.Set("aria-live", "polite")
	e.message.Set("textContent", "Edits update this step's JSON output and all downstream steps.")

	preview := div("urlparts-preview")
	previewTitle := el("strong")
	previewTitle.Set("textContent", "Rebuilt URL")
	e.rebuilt = el("textarea")
	e.rebuilt.Set("className", "urlparts-rebuilt")
	e.rebuilt.Set("readOnly", true)
	e.rebuilt.Set("rows", 3)
	e.rebuilt.Call("setAttribute", "aria-label", "Rebuilt URL")
	e.copy = e.actionButton("copy", "Copy URL", func() {
		if e.rebuilt.Get("value").String() == "" {
			return
		}
		value := e.rebuilt.Get("value").String()
		clipboard := js.Global().Get("navigator").Get("clipboard")
		if clipboard.Truthy() {
			clipboard.Call("writeText", value)
		} else {
			js.Global().Call("prompt", "Copy rebuilt URL:", value)
		}
	})
	previewBody := div("urlparts-preview-body")
	appendChildren(previewBody, e.rebuilt, e.copy)
	appendChildren(preview, previewTitle, previewBody)

	rawWrap, rawInput := checkbox("Show original encoded values", e.showRaw)
	rawWrap.Set("className", "urlparts-raw-toggle")
	rawInput.Call("setAttribute", "aria-label", "Show original encoded values")
	e.on(rawInput, "change", func() {
		e.showRaw = rawInput.Get("checked").Bool()
		e.setRawVisible()
	})

	appendChildren(e.root, e.message, preview, rawWrap)
	e.buildAuthority()
	e.buildPath()
	e.buildQuery()
	e.buildFragment()
	e.setRawVisible()
	e.refreshRebuiltURL()
}

func (e *webURLPartsEditor) buildAuthority() {
	section, body := urlPartsSection("Authority")
	grid := div("urlparts-field-grid")
	e.field(grid, "Scheme", "URL scheme", "text", e.doc.Scheme, func(value string) {
		e.doc.Scheme = value
		e.commit(false)
	})
	e.field(grid, "Hostname", "URL hostname", "text", e.doc.Hostname, func(value string) {
		e.doc.Hostname = value
		e.commit(false)
	})
	e.field(grid, "Port", "URL port", "text", e.doc.Port, func(value string) {
		e.doc.Port = value
		e.commit(false)
	})
	body.Call("appendChild", grid)

	userinfoWrap, userinfoInput := checkbox("URL contains user information", e.doc.Userinfo != nil)
	userinfoInput.Call("setAttribute", "aria-label", "URL contains user information")
	e.on(userinfoInput, "change", func() {
		if userinfoInput.Get("checked").Bool() {
			if e.doc.Userinfo == nil {
				e.doc.Userinfo = &formatters.URLPartsUserinfo{}
			}
		} else {
			e.doc.Userinfo = nil
		}
		e.commit(true)
	})
	body.Call("appendChild", userinfoWrap)
	if e.doc.Userinfo != nil {
		userGrid := div("urlparts-field-grid")
		e.field(userGrid, "Username", "URL username", "text", e.doc.Userinfo.Username, func(value string) {
			e.doc.Userinfo.Username = value
			e.commit(false)
		})
		password := e.field(userGrid, "Password", "URL password", "password", e.doc.Userinfo.Password, func(value string) {
			e.doc.Userinfo.Password = value
			e.doc.Userinfo.PasswordSet = true
			e.commit(false)
		})
		if !e.doc.Userinfo.PasswordSet {
			password.Set("placeholder", "No password in URL")
		}
		body.Call("appendChild", userGrid)
	}
	if e.doc.Raw != nil {
		e.rawValue(body, "Encoded host: "+e.doc.Raw.Host)
	}
	e.root.Call("appendChild", section)
}

func (e *webURLPartsEditor) buildPath() {
	section, body := urlPartsSection("Path segments")
	rows := div("urlparts-rows")
	if len(e.doc.PathSegments) == 0 {
		empty := div("urlparts-empty")
		empty.Set("textContent", "No path segments")
		rows.Call("appendChild", empty)
	}
	for i, segment := range e.doc.PathSegments {
		i := i
		row := div("urlparts-row urlparts-path-row")
		row.Call("setAttribute", "data-index", strconv.Itoa(i))
		index := el("span")
		index.Set("className", "urlparts-index")
		index.Set("textContent", fmt.Sprintf("%d", i+1))
		input := el("input")
		input.Set("type", "text")
		input.Set("value", segment)
		input.Call("setAttribute", "aria-label", fmt.Sprintf("Path segment %d", i+1))
		e.on(input, "input", func() {
			if i >= len(e.doc.PathSegments) {
				return
			}
			e.doc.PathSegments[i] = input.Get("value").String()
			e.commit(false)
		})
		actions := e.rowActions("path segment", i, len(e.doc.PathSegments), func(delta int) {
			e.doc.PathSegments[i], e.doc.PathSegments[i+delta] = e.doc.PathSegments[i+delta], e.doc.PathSegments[i]
			e.commit(true)
		}, func() {
			e.doc.PathSegments = slices.Insert(e.doc.PathSegments, i+1, e.doc.PathSegments[i])
			e.commit(true)
		}, func() {
			e.doc.PathSegments = slices.Delete(e.doc.PathSegments, i, i+1)
			e.commit(true)
		})
		appendChildren(row, index, input, actions)
		rows.Call("appendChild", row)
	}
	body.Call("appendChild", rows)
	add := e.textButton("Add path segment", func() {
		e.doc.PathSegments = append(e.doc.PathSegments, "")
		e.commit(true)
	})
	add.Set("className", "urlparts-add")
	body.Call("appendChild", add)
	if e.doc.Raw != nil {
		e.rawValue(body, "Encoded path: "+e.doc.Raw.Path)
	}
	e.root.Call("appendChild", section)
}

func (e *webURLPartsEditor) buildQuery() {
	section, body := urlPartsSection("Query parameters")
	forceWrap, forceInput := checkbox("Keep ? when the query is empty", e.doc.ForceQuery)
	forceInput.Call("setAttribute", "aria-label", "Keep question mark when query is empty")
	e.on(forceInput, "change", func() {
		e.doc.ForceQuery = forceInput.Get("checked").Bool()
		e.commit(false)
	})
	body.Call("appendChild", forceWrap)

	rows := div("urlparts-rows")
	if len(e.doc.Query) == 0 {
		empty := div("urlparts-empty")
		empty.Set("textContent", "No query parameters")
		rows.Call("appendChild", empty)
	}
	for i := range e.doc.Query {
		i := i
		parameter := &e.doc.Query[i]
		row := div("urlparts-query-row")
		row.Call("setAttribute", "data-index", strconv.Itoa(i))
		fields := div("urlparts-query-fields")
		key := el("input")
		key.Set("type", "text")
		key.Set("placeholder", "Key")
		key.Set("value", parameter.Key)
		key.Call("setAttribute", "aria-label", fmt.Sprintf("Query key %d", i+1))
		value := el("input")
		value.Set("type", "text")
		value.Set("placeholder", "Value")
		value.Set("value", parameter.Value)
		value.Set("disabled", !parameter.HasValue)
		value.Call("setAttribute", "aria-label", fmt.Sprintf("Query value %d", i+1))
		e.on(key, "input", func() {
			if i >= len(e.doc.Query) {
				return
			}
			e.doc.Query[i].Key = key.Get("value").String()
			e.commit(false)
		})
		e.on(value, "input", func() {
			if i >= len(e.doc.Query) {
				return
			}
			e.doc.Query[i].Value = value.Get("value").String()
			e.commit(false)
		})
		appendChildren(fields, key, value)
		hasValueWrap, hasValueInput := checkbox("has value", parameter.HasValue)
		hasValueWrap.Set("className", "urlparts-has-value")
		hasValueInput.Call("setAttribute", "aria-label", fmt.Sprintf("Query parameter %d has value", i+1))
		e.on(hasValueInput, "change", func() {
			if i >= len(e.doc.Query) {
				return
			}
			hasValue := hasValueInput.Get("checked").Bool()
			e.doc.Query[i].HasValue = hasValue
			value.Set("disabled", !hasValue)
			if !hasValue {
				e.doc.Query[i].Value = ""
				value.Set("value", "")
			}
			e.commit(false)
		})
		actions := e.rowActions("query parameter", i, len(e.doc.Query), func(delta int) {
			e.doc.Query[i], e.doc.Query[i+delta] = e.doc.Query[i+delta], e.doc.Query[i]
			e.commit(true)
		}, func() {
			e.doc.Query = slices.Insert(e.doc.Query, i+1, e.doc.Query[i])
			e.commit(true)
		}, func() {
			e.doc.Query = slices.Delete(e.doc.Query, i, i+1)
			e.commit(true)
		})
		appendChildren(row, fields, hasValueWrap, actions)
		if e.doc.Query[i].HasValue {
			e.rawValue(row, fmt.Sprintf("Encoded: %s=%s", parameter.RawKey, parameter.RawValue))
		} else {
			e.rawValue(row, "Encoded: "+parameter.RawKey)
		}
		rows.Call("appendChild", row)
	}
	body.Call("appendChild", rows)
	add := e.textButton("Add query parameter", func() {
		e.doc.Query = append(e.doc.Query, formatters.URLQueryParameter{HasValue: true})
		e.commit(true)
	})
	add.Set("className", "urlparts-add")
	body.Call("appendChild", add)
	e.root.Call("appendChild", section)
}

func (e *webURLPartsEditor) buildFragment() {
	section, body := urlPartsSection("Fragment and advanced fields")
	grid := div("urlparts-field-grid")
	e.field(grid, "Fragment", "URL fragment", "text", e.doc.Fragment, func(value string) {
		e.doc.Fragment = value
		e.commit(false)
	})
	e.field(grid, "Opaque data", "URL opaque data", "text", e.doc.Opaque, func(value string) {
		e.doc.Opaque = value
		e.commit(false)
	})
	body.Call("appendChild", grid)
	omitWrap, omitInput := checkbox("Host is omitted", e.doc.OmitHost)
	omitInput.Call("setAttribute", "aria-label", "URL host is omitted")
	e.on(omitInput, "change", func() {
		e.doc.OmitHost = omitInput.Get("checked").Bool()
		e.commit(false)
	})
	body.Call("appendChild", omitWrap)
	if e.doc.Raw != nil {
		e.rawValue(body, "Encoded fragment: "+e.doc.Raw.Fragment)
	}
	e.root.Call("appendChild", section)
}

func (e *webURLPartsEditor) field(parent js.Value, labelText, ariaLabel, inputType, value string, apply func(string)) js.Value {
	wrap := el("label")
	wrap.Set("className", "urlparts-field")
	label := el("span")
	label.Set("textContent", labelText)
	input := el("input")
	input.Set("type", inputType)
	input.Set("value", value)
	input.Call("setAttribute", "aria-label", ariaLabel)
	e.on(input, "input", func() { apply(input.Get("value").String()) })
	appendChildren(wrap, label, input)
	parent.Call("appendChild", wrap)
	return input
}

func (e *webURLPartsEditor) rowActions(kind string, index, count int, move func(int), duplicate, remove func()) js.Value {
	actions := div("urlparts-row-actions")
	up := e.iconButton("arrow-up", fmt.Sprintf("Move %s %d up", kind, index+1), func() { move(-1) })
	down := e.iconButton("arrow-down", fmt.Sprintf("Move %s %d down", kind, index+1), func() { move(1) })
	copy := e.iconButton("copy", fmt.Sprintf("Duplicate %s %d", kind, index+1), duplicate)
	trash := e.iconButton("trash", fmt.Sprintf("Remove %s %d", kind, index+1), remove)
	up.Set("disabled", index == 0)
	down.Set("disabled", index == count-1)
	appendChildren(actions, up, down, copy, trash)
	return actions
}

func (e *webURLPartsEditor) rawValue(parent js.Value, value string) {
	raw := div("urlparts-raw")
	raw.Set("textContent", value)
	parent.Call("appendChild", raw)
}

func (e *webURLPartsEditor) setRawVisible() {
	items := e.root.Call("querySelectorAll", ".urlparts-raw")
	display := "none"
	if e.showRaw {
		display = "block"
	}
	for i := 0; i < items.Get("length").Int(); i++ {
		items.Call("item", i).Get("style").Set("display", display)
	}
}

func (e *webURLPartsEditor) refreshRebuiltURL() {
	if e.doc == nil || !e.rebuilt.Truthy() {
		return
	}
	rebuilt, err := formatters.RebuildURLParts(e.doc)
	if err != nil {
		e.rebuilt.Set("value", "")
		e.copy.Set("disabled", true)
		e.message.Set("className", "urlparts-message invalid")
		e.message.Set("textContent", "URL fields are not valid yet: "+err.Error())
		return
	}
	e.rebuilt.Set("value", rebuilt)
	e.copy.Set("disabled", false)
	e.message.Set("className", "urlparts-message")
	e.message.Set("textContent", "Edits update this step's JSON output and all downstream steps.")
}

func (e *webURLPartsEditor) commit(structural bool) {
	if e.doc == nil {
		return
	}
	var output bytes.Buffer
	if err := formatters.EncodeURLPartsJSON(&output, e.doc); err != nil {
		e.message.Set("className", "urlparts-message invalid")
		e.message.Set("textContent", "Could not update URL Parts JSON: "+err.Error())
		return
	}
	data := output.Bytes()
	e.lastJSON = string(data)
	pipe.EditOutput(e.card.index, data)
	if structural {
		runBusy("Updating URL", rebuild)
		return
	}
	e.refreshRebuiltURL()
	renderOutput(e.card)
	refreshOutputs(e.card.index + 1)
}

func (e *webURLPartsEditor) actionButton(icon, label string, fn func()) js.Value {
	b := el("button")
	b.Set("type", "button")
	b.Set("className", "urlparts-copy icon-label")
	b.Call("setAttribute", "aria-label", label)
	text := el("span")
	text.Set("textContent", label)
	appendChildren(b, iconGraphic(icon), text)
	e.on(b, "click", fn)
	return b
}

func (e *webURLPartsEditor) textButton(label string, fn func()) js.Value {
	b := el("button")
	b.Set("type", "button")
	b.Set("textContent", label)
	e.on(b, "click", fn)
	return b
}

func (e *webURLPartsEditor) iconButton(icon, label string, fn func()) js.Value {
	b := el("button")
	b.Set("type", "button")
	b.Set("className", "icon")
	setIconOnlyButtonContent(b, icon, label)
	b.Call("setAttribute", "aria-label", label)
	e.on(b, "click", fn)
	return b
}

func (e *webURLPartsEditor) on(node js.Value, event string, fn func()) {
	cb := js.FuncOf(func(js.Value, []js.Value) any {
		fn()
		return nil
	})
	e.callbacks = append(e.callbacks, cb)
	callbacks = append(callbacks, cb)
	node.Call("addEventListener", event, cb)
}

func (e *webURLPartsEditor) releaseCallbacks() {
	if len(e.callbacks) == 0 {
		return
	}
	kept := callbacks[:0]
	for _, callback := range callbacks {
		found := false
		for _, owned := range e.callbacks {
			if callback.Value.Equal(owned.Value) {
				found = true
				break
			}
		}
		if !found {
			kept = append(kept, callback)
		}
	}
	callbacks = kept
	for _, callback := range e.callbacks {
		callback.Release()
	}
	e.callbacks = nil
}

func urlPartsSection(title string) (section, body js.Value) {
	section = div("urlparts-section")
	heading := el("h4")
	heading.Set("textContent", title)
	body = div("urlparts-section-body")
	appendChildren(section, heading, body)
	return section, body
}
