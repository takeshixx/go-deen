package formatters

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var jsonTestData = "{\"version\": \"v1.0\", \"enable\": true}"
var jsonTestDataProcessed = `{
    "enable": true,
    "version": "v1.0"
}
`
var jsonTestDataProcessedColor = `{
    "enable": true,
    "version": "v1.0"
}`

func TestNewPluginJSONFormat(t *testing.T) {
	p := NewPluginJSONFormatter()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginJSONFormatter: %s", reflect.TypeOf(p))
	}
}

func TestPluginJSONFormatProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(jsonTestData)
	plugin := NewPluginJSONFormatter()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(jsonTestDataProcessed)); c != 0 {
			t.Errorf("jsonTestDataProcessed data wrong: %s != %s", destWriter.Bytes(), jsonTestDataProcessed)
		}
	}
}

func TestPluginJSONFormatProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(jsonTestData)
	plugin := NewPluginJSONFormatter()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{""})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(jsonTestDataProcessedColor)); c != 0 {
			t.Errorf("TestPluginJSONFormatProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), jsonTestDataProcessedColor)
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(jsonTestData)
	plugin = NewPluginJSONFormatter()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-no-color"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(jsonTestDataProcessed)); c != 0 {
			t.Errorf("TestPluginJSONFormatProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), jsonTestDataProcessed)
		}
	}
}

func TestPluginJSONFormatUsage(t *testing.T) {
	p := NewPluginJSONFormatter()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
