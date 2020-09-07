package gui

import (
	"testing"

	"fyne.io/fyne/test"
)

func TestSetUpGUI(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("deen test")
	setUpGUI(a, w)
}
