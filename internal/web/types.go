// +build js,wasm

package web

import (
	"github.com/gopherjs/vecty"
)

type DeenWeb struct {
	vecty.Core
	EncoderWidgets []*EncoderWidget
	currentEncoder *EncoderWidget
}

// Run is the main function of a deen web instance.
func (dw *DeenWeb) Run() {
	vecty.SetTitle("deen web")
	vecty.RenderBody(dw)
}

func (dw *DeenWeb) Reload() {
	vecty.Rerender(dw)
}

// NewDeenWeb creates a new deen web instance.
func NewDeenWeb() (dw *DeenWeb) {
	dw = &DeenWeb{}
	dw.AddEncoder()
	return
}
