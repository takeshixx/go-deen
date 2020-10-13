// +build js,wasm

package web

import (
	"github.com/gopherjs/vecty"
)

// Run is the main function of a deen web instance.
func (dw *DeenWeb) Run() (err error) {
	vecty.SetTitle("deen")
	vecty.RenderBody(dw)
	return
}
