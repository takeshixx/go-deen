package codecs

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

var hexInputData = "asd1239999"
var hexInputDataProcessed = []byte("61736431323339393939")

func TestNewPluginHex(t *testing.T) {
	p := NewPluginHex()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginHex: %s", reflect.TypeOf(p))
	}
}

func TestPluginHexProcess(t *testing.T) {
	p := NewPluginHex()
	r := strings.NewReader(hexInputData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("HexProcess failed: %s", e)
	}
	if c := bytes.Compare(d, hexInputDataProcessed); c != 0 {
		t.Errorf("HexProcess data wrong: %s", d)
	}
}

func TestPluginHexUnprocess(t *testing.T) {
	p := NewPluginHex()
	r := bytes.NewReader(hexInputDataProcessed)
	d, e := p.UnprocessStreamFunc(r)
	if e != nil {
		t.Errorf("HexUnprocess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte(hexInputData)); c != 0 {
		t.Errorf("HexUnprocess data wrong: %s", d)
	}
}
