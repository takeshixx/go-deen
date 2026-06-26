//go:build js && wasm

// WebAssembly entry point: builds a minimal browser UI for deen's plugins.
package main

import (
	"bytes"
	"flag"
	"strings"
	"syscall/js"

	"github.com/takeshixx/deen/internal/plugins"
)

func main() {
	doc := js.Global().Get("document")
	body := doc.Get("body")

	create := func(tag string) js.Value { return doc.Call("createElement", tag) }

	container := create("div")
	container.Get("classList").Call("add", "deen")

	heading := create("h1")
	heading.Set("textContent", "deen")
	container.Call("appendChild", heading)

	// Plugin selector.
	controls := create("div")
	controls.Get("classList").Call("add", "controls")

	sel := create("select")
	for _, name := range plugins.Names() {
		opt := create("option")
		opt.Set("value", name)
		opt.Set("textContent", name)
		sel.Call("appendChild", opt)
	}
	controls.Call("appendChild", sel)

	decodeLabel := create("label")
	decode := create("input")
	decode.Set("type", "checkbox")
	decodeLabel.Call("appendChild", decode)
	decodeLabel.Call("appendChild", doc.Call("createTextNode", " decode"))
	controls.Call("appendChild", decodeLabel)

	runBtn := create("button")
	runBtn.Set("textContent", "Run")
	controls.Call("appendChild", runBtn)

	container.Call("appendChild", controls)

	input := create("textarea")
	input.Set("placeholder", "input")
	input.Get("classList").Call("add", "io")
	container.Call("appendChild", input)

	output := create("pre")
	output.Get("classList").Call("add", "io")
	container.Call("appendChild", output)

	body.Call("appendChild", container)

	run := js.FuncOf(func(this js.Value, args []js.Value) any {
		name := sel.Get("value").String()
		cmd := name
		if decode.Get("checked").Bool() {
			cmd = "." + name
		}
		plugin, unprocess, ok := plugins.Resolve(cmd)
		if !ok {
			output.Set("textContent", "unknown plugin: "+name)
			return nil
		}
		fn := plugin.Process
		if unprocess {
			if plugin.Unprocess == nil {
				output.Set("textContent", name+" does not support decoding")
				return nil
			}
			fn = plugin.Unprocess
		}
		fs := flag.NewFlagSet(name, flag.ContinueOnError)
		if plugin.RegisterFlags != nil {
			plugin.RegisterFlags(fs)
		}
		_ = fs.Parse(nil)

		var buf bytes.Buffer
		if err := fn(strings.NewReader(input.Get("value").String()), &buf, fs); err != nil {
			output.Set("textContent", "error: "+err.Error())
			return nil
		}
		output.Set("textContent", buf.String())
		return nil
	})
	runBtn.Call("addEventListener", "click", run)

	// Keep the Go program (and the click callback) alive.
	select {}
}
