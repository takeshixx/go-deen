package gui

import (
	"reflect"
	"testing"

	"fyne.io/fyne/test"
)

func TestNewDeenGUI(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("deen test")
	g, err := NewDeenGUI(a, w)
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(g) != reflect.TypeOf(&DeenGUI{}) {
		t.Errorf("Invalid type returned: %v", g)
	}
}
