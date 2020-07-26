package codecs

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

var b32InputData = "asdtest123"
var b32InputDataProcessed = []byte("MFZWI5DFON2DCMRT")
var b32InputDataProcessedHex = []byte("C5PM8T35EDQ32CHJ")

func TestNewPluginBase32(t *testing.T) {
	p := NewPluginBase32()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginBase32: %s", reflect.TypeOf(p))
	}
}

func TestPluginBase32Process(t *testing.T) {
	p := NewPluginBase32()
	r := strings.NewReader(b32InputData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("Base32Process failed: %s", e)
	}
	if c := bytes.Compare(d, b32InputDataProcessed); c != 0 {
		t.Errorf("Base32Process data wrong: %s", d)
	}
}

func TestPluginBase32Unprocess(t *testing.T) {
	p := NewPluginBase32()
	r := bytes.NewReader(b32InputDataProcessed)
	d, e := p.UnprocessStreamFunc(r)
	if e != nil {
		t.Errorf("Base32Unprocess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte(b32InputData)); c != 0 {
		t.Errorf("Base32Unprocess data wrong: %s", d)
	}
}
