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

func TestPluginBase85ProcessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b85InputData)
	plugin := NewPluginBase85()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b85InputDataProcessed); c != 0 {
			t.Errorf("TestPluginBase85ProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), b85InputDataProcessed)
		}
	}
}

func TestPluginBase85UnprocessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b85InputDataProcessed)
	plugin := NewPluginBase85()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b85InputData)); c != 0 {
			t.Errorf("TestPluginBase85UnprocessDeenTask data wrong: %s != %s", destWriter.Bytes(), []byte(b85InputData))
		}
	}
}
