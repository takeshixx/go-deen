package codecs

import (
	"bytes"
	"encoding/hex"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var hexInputData = "asd1239999"
var hexInputDataProcessed = []byte("61736431323339393939")
var hexBinData = "ad13285a5a48976ef51f18c601954a703e3c0c5a"

func TestNewPluginHex(t *testing.T) {
	p := NewPluginHex()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
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
			t.Errorf("TestPluginHexProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), hexInputDataProcessed)
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
			t.Errorf("TestPluginHexUnprocessDeenTask data wrong: %s != %s", destWriter.Bytes(), []byte(hexInputData))
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(hexBinData)
	plugin = NewPluginHex()
	plugin.UnprocessDeenTaskFunc(task)
	decoded, err := hex.DecodeString(hexBinData)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), decoded); c != 0 {
			t.Errorf("TestPluginHexUnprocessDeenTask data wrong: %s != %s", destWriter.Bytes(), decoded)
		}
	}
}

func TestPluginHexUsage(t *testing.T) {
	p := NewPluginHex()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
