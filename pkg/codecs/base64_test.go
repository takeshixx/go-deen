package codecs

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

var b64InputData = "asd123<<<<>>>>"
var b64InputDataProcessed = []byte("YXNkMTIzPDw8PD4+Pj4=")
var b64InputDataProcessedURL = []byte("YXNkMTIzPDw8PD4-Pj4=")

func TestNewPluginBase64(t *testing.T) {
	p := NewPluginBase64()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginBase64: %s", reflect.TypeOf(p))
	}
}

func TestPluginBase64Process(t *testing.T) {
	p := NewPluginBase64()
	r := strings.NewReader(b64InputData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("Base64Process failed: %s", e)
	}
	if c := bytes.Compare(d, b64InputDataProcessed); c != 0 {
		t.Errorf("Base64Process data wrong: %s", d)
	}
}

func TestPluginBase64Unprocess(t *testing.T) {
	p := NewPluginBase64()
	r := bytes.NewReader(b64InputDataProcessed)
	d, e := p.UnprocessStreamFunc(r)
	if e != nil {
		t.Errorf("Base64Unprocess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte(b64InputData)); c != 0 {
		t.Errorf("Base64Unprocess data wrong: %s", d)
	}
}
