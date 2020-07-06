package codecs

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

var b85InputData = "asd1239999"
var b85InputDataProcessed = []byte("@<5s61,CpN3B7")

func TestNewPluginBase85(t *testing.T) {
	p := NewPluginBase85()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginBase85: %s", reflect.TypeOf(p))
	}
}

func TestPluginBase85Process(t *testing.T) {
	p := NewPluginBase85()
	r := strings.NewReader(b85InputData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("Base85Process failed: %s", e)
	}
	if c := bytes.Compare(d, b85InputDataProcessed); c != 0 {
		t.Errorf("Base85Process data wrong: %s", d)
	}
}

func TestPluginBase85Unprocess(t *testing.T) {
	p := NewPluginBase85()
	r := bytes.NewReader(b85InputDataProcessed)
	d, e := p.UnprocessStreamFunc(r)
	if e != nil {
		t.Errorf("Base85Unprocess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte(b85InputData)); c != 0 {
		t.Errorf("Base85Unprocess data wrong: %s", d)
	}
}
