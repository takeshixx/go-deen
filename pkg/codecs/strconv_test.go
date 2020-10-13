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

var strconvTestData = "☺"
var strconvTestDataProcessed = "\\u263a"

func TestNewPluginStrconv(t *testing.T) {
	p := NewPluginStrconv()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginStrconv: %s", reflect.TypeOf(p))
	}
}

func TestPluginStrconvProcessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(strconvTestData)
	plugin := NewPluginStrconv()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{""})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(strconvTestDataProcessed)); c != 0 {
			t.Errorf("TestPluginStrconvProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), strconvTestDataProcessed)
		}
	}
}

func TestPluginStrconvUnprocessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(strconvTestDataProcessed)
	plugin := NewPluginStrconv()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{""})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(strconvTestData)); c != 0 {
			t.Errorf("TestPluginStrconvUnprocessDeenTask data wrong: %s != %s", destWriter.Bytes(), []byte(strconvTestData))
		}
	}
}

func TestPluginStrconvUsage(t *testing.T) {
	p := NewPluginStrconv()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
