// +build js,wasm

package web

import (
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/takeshixx/deen/internal/plugins"
)

type DeenWeb struct {
	vecty.Core
	Body           vecty.ComponentOrHTML
	EncoderWidgets []*EncoderWidget
	currentEncoder *EncoderWidget
}

func (dw *DeenWeb) Render() vecty.ComponentOrHTML {
	var e []vecty.MarkupOrChild
	e = append(e, vecty.Markup(
		vecty.Style("width", "100%"),
		vecty.Style("float", "left"),
	))
	for _, a := range dw.EncoderWidgets {
		e = append(e, a)
	}
	dw.Body = elem.Body(
		// Display a textarea on the right-hand side of the page.
		elem.Div(
			e...,
		),
	)
	return dw.Body
}

func (dw *DeenWeb) SetCurrentEncoder(e *EncoderWidget) {
	dw.currentEncoder = e
}

func (dw *DeenWeb) CurrentEncoder() (e *EncoderWidget) {
	return dw.currentEncoder
}

func (dw *DeenWeb) NextEncoder() (e *EncoderWidget) {
	for i, enc := range dw.EncoderWidgets {
		if enc == dw.currentEncoder {
			if len(dw.EncoderWidgets) > i+1 {
				return dw.EncoderWidgets[i+1]
			} else {
				return dw.AddEncoder()
			}
		}
	}
	return
}

func (dw *DeenWeb) PreviousEncoder() (e *EncoderWidget) {
	if dw.currentEncoder == dw.EncoderWidgets[0] {
		return dw.EncoderWidgets[0]
	} else {
		for i, enc := range dw.EncoderWidgets {
			if enc == dw.currentEncoder {
				return dw.EncoderWidgets[i-1]
			}
		}
	}
	return
}

func (dw *DeenWeb) AddEncoder() (e *EncoderWidget) {
	e = NewEncoderWidget(dw)
	dw.currentEncoder = e
	dw.EncoderWidgets = append(dw.EncoderWidgets, e)
	return
}

func (dw *DeenWeb) RunPlugin(pluginCmd string) {
	plugin := plugins.GetForCmd(pluginCmd)
	if plugin == nil {
		return
	}
	ce := dw.CurrentEncoder()
	ce.Plugin = plugin
	dw.RunChain()
}

func (dw *DeenWeb) RunChain() {
	var processed []byte
	var err error
	var nextEnc *EncoderWidget
	for _, e := range dw.EncoderWidgets {
		processed, err = e.Process()
		if err != nil {
			return
		}
		nextEnc = dw.NextEncoder()
		nextEnc.SetContent(string(processed))
	}
	vecty.Rerender(dw)
}

// NewDeenWeb creates a new deen web instance.
func NewDeenWeb() (dw *DeenWeb) {
	dw = &DeenWeb{}
	dw.AddEncoder()
	return
}
