package types

import (
	"reflect"
	"testing"
)

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()
	if reflect.TypeOf(p) != reflect.TypeOf(&DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPlugin: %s", reflect.TypeOf(p))
	}
}
