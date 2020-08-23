package codecs

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var htmlInputData = "<h1>deen</h1>"
var htmlInputDataProcessed = []byte("&lt;h1&gt;deen&lt;/h1&gt;")

func TestNewPluginHtml(t *testing.T) {
	p := NewPluginHTML()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginHTML: %s", reflect.TypeOf(p))
	}
}

func TestPluginHtmlProcess(t *testing.T) {
	p := NewPluginHTML()
	r := strings.NewReader(htmlInputData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("HtmlProcess failed: %s", e)
	}
	if c := bytes.Compare(d, htmlInputDataProcessed); c != 0 {
		t.Errorf("HtmlProcess data wrong: %s", d)
	}
}

func TestPluginHtmlUnprocess(t *testing.T) {
	p := NewPluginHTML()
	r := bytes.NewReader(htmlInputDataProcessed)
	d, e := p.UnprocessStreamFunc(r)
	if e != nil {
		t.Errorf("HtmlUnprocess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte(htmlInputData)); c != 0 {
		t.Errorf("HtmlUnprocess data wrong: %s", d)
	}
}

func TestPluginHtmlUsage(t *testing.T) {
	p := NewPluginHTML()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
