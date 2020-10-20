// +build gui

package gui

import (
	"reflect"
	"testing"
)

func TestNewDeenGUI(t *testing.T) {
	g, err := NewDeenGUI()
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(g) != reflect.TypeOf(&DeenGUI{}) {
		t.Errorf("Invalid type returned: %v", g)
	}
}
