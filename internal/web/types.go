// +build js,wasm

package web

import (
	"github.com/gopherjs/vecty"
)

type DeenWeb struct {
	vecty.Core
	Body           vecty.ComponentOrHTML
	EncoderWidgets []*EncoderWidget
	currentEncoder *EncoderWidget
}

// Run is the main function of a deen web instance.
func (dw *DeenWeb) Run() (err error) {
	vecty.SetTitle("deen web")
	vecty.RenderBody(dw)
	return
}

// NewDeenWeb creates a new deen web instance.
func NewDeenWeb() (dw *DeenWeb) {
	dw = &DeenWeb{}
	dw.AddEncoder()
	return
}
