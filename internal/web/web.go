package web

import (
	"syscall/js"
)

var document = js.Global().Get("document")

// Run is the main function of a deen web instance.
func (dw *DeenWeb) Run() (err error) {
	/* 	document.Set("title", "deen web")
	   	body := document.Get("body")
	   	body.Set("innerHTML", template) */
	return
}

// NewDeenWeb creates a new deen web instance.
func NewDeenWeb() (dw *DeenWeb) {
	dw = &DeenWeb{}
	return
}
