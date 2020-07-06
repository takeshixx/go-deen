package codecs

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

var urlInputData = "test?deen=true"
var urlInputDataProcessed = []byte("test%3Fdeen%3Dtrue")

func TestNewPluginURL(t *testing.T) {
	p := NewPluginURL()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginHTML: %s", reflect.TypeOf(p))
	}
}

func TestPluginURLProcess(t *testing.T) {
	p := NewPluginURL()
	r := strings.NewReader(urlInputData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("URLProcess failed: %s", e)
	}
	if c := bytes.Compare(d, urlInputDataProcessed); c != 0 {
		t.Errorf("URLProcess data wrong: %s", d)
	}
}

func TestPluginURLUnprocess(t *testing.T) {
	p := NewPluginURL()
	r := bytes.NewReader(urlInputDataProcessed)
	d, e := p.UnprocessStreamFunc(r)
	if e != nil {
		t.Errorf("URLUnprocess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte(urlInputData)); c != 0 {
		t.Errorf("URLUnprocess data wrong: %s", d)
	}
}
