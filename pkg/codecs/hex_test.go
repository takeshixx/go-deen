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

var hexInputData = "asd1239999"
var hexInputDataProcessed = []byte("61736431323339393939")

func TestNewPluginHex(t *testing.T) {
	p := NewPluginHex()
	if reflect.TypeOf(p) != reflect.TypeOf(types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginHex: %s", reflect.TypeOf(p))
	}
}

func TestPluginHexProcessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(hexInputData)
	plugin := NewPluginHex()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), hexInputDataProcessed); c != 0 {
			t.Errorf("TestPluginBase85ProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), hexInputDataProcessed)
		}
	}
}

func TestPluginHexUnprocessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(hexInputDataProcessed)
	plugin := NewPluginHex()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(hexInputData)); c != 0 {
			t.Errorf("TestPluginBase85ProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), []byte(hexInputData))
		}
	}
}

func TestPluginHexUsage(t *testing.T) {
	p := NewPluginHex()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
