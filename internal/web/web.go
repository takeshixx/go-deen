// +build js,wasm

package web

import (
	"fmt"

	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/takeshixx/deen/internal/plugins"
)

func (dw *DeenWeb) Render() vecty.ComponentOrHTML {
	var encoders vecty.List
	fmt.Printf("Rendering main with EncoderWidgets: %v\n", dw.EncoderWidgets)
	for _, a := range dw.EncoderWidgets {
		b := a
		encoders = append(encoders, b)
	}
	return elem.Body(
		elem.Div(
			vecty.Markup(
				vecty.Style("width", "100%"),
			),
			encoders,
		),
	)
}

func (dw *DeenWeb) SetCurrentEncoder(e *EncoderWidget) {
	fmt.Printf("Setting current encoder: %v\n", e)
	for i, x := range dw.EncoderWidgets {
		enc := x
		if e == enc {
			dw.EncoderWidgets[i].Border = true
			dw.currentEncoder = e
		} else {
			dw.EncoderWidgets[i].Border = false
		}
		vecty.Rerender(enc)
	}
}

func (dw *DeenWeb) CurrentEncoder() (e *EncoderWidget) {
	return dw.currentEncoder
}

func (dw *DeenWeb) NextEncoder() (e *EncoderWidget) {
	return dw.NextAfterEncoder(dw.currentEncoder)
}

func (dw *DeenWeb) NextAfterEncoder(a *EncoderWidget) (e *EncoderWidget) {
	for i, enc := range dw.EncoderWidgets {
		if enc == a {
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
	dw.SetCurrentEncoder(e)
	dw.EncoderWidgets = append(dw.EncoderWidgets, e)
	return
}

func (dw *DeenWeb) RemoveEncoder(e *EncoderWidget) {
	if e == dw.EncoderWidgets[0] {
		fmt.Println("Clearing root encoder")
		dw.EncoderWidgets = nil
		dw.AddEncoder()
	} else {
		// If enc is not the root widget, we have a previous widget
		previous := dw.PreviousEncoder()
		if e == dw.EncoderWidgets[len(dw.EncoderWidgets)-1] {
			fmt.Println("Last encoder")
			// Remove the last encoder
			dw.EncoderWidgets = dw.EncoderWidgets[:len(dw.EncoderWidgets)-1]
			// And clear the plugin of the previous encoder
			previous.Plugin = nil
		} else {
			fmt.Println("Not last encoder")
			for i, enc := range dw.EncoderWidgets {
				if e == enc {
					dw.EncoderWidgets = append(dw.EncoderWidgets[:i], dw.EncoderWidgets[i+1:]...)
					// Transfer plugin to previous widget
					dw.EncoderWidgets[i-1].Plugin = e.Plugin
					break
				}
			}
		}
	}
	vecty.Rerender(dw)
}

func (dw *DeenWeb) RunPlugin(pluginCmd string) {
	plugin := plugins.GetForCmd(pluginCmd)
	if plugin == nil {
		return
	}
	ce := dw.CurrentEncoder()
	ce.Plugin = plugin
	dw.RunChainFrom(ce)
}

func (dw *DeenWeb) RunChain() {
	dw.RunChainFrom(dw.EncoderWidgets[0])
}

func (dw *DeenWeb) RunChainFrom(e *EncoderWidget) {
	encodersIndex := 0
	if e != dw.EncoderWidgets[0] {
		// We are not starting at the root widget
		for i, enc := range dw.EncoderWidgets {
			if enc == e {
				encodersIndex = i
				break
			}
		}
	}

	var processed []byte
	var err error
	var nextEnc *EncoderWidget
	for i, e := range dw.EncoderWidgets {
		if i < encodersIndex {
			// Skip encoders before the current one
			continue
		}
		processed, err = e.Process()
		if err != nil {
			return
		}
		if len(processed) < 1 {
			if len(dw.EncoderWidgets) > 1 {
				for _, de := range dw.EncoderWidgets[encodersIndex:] {
					fmt.Printf("Removing encoder: %v\n", de)
					dw.RemoveEncoder(de)
				}
			}
			break
		}
		e.Render()
		nextEnc = dw.NextAfterEncoder(e)
		nextEnc.SetContent(string(processed))
	}
	vecty.Rerender(dw)
}
