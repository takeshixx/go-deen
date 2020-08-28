package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var flateTestData = []byte("deenflatetestdata123123192840912atestflatedeenflatedeentest")
var deflatedData = "4a494dcd4bcb492c492d492d2e49492c4934343206214b230b13034b43a344903858015c2588011205040000ffff"
var deflateDataLevel1 = "04c0b10dc0300844d195024e118f7312df9545c3edafbc823e57c68c4b56e48a5cb1f37b9f1d29333e57a6a0cf95296833fe030000ffff"

func TestPluginFlateProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(flateTestData)
	plugin := NewPluginFlate()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(deflatedData)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginFlateProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), deflatedData)
		}
	}
}

func TestPluginFlateProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(flateTestData)
	plugin := NewPluginFlate()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-level", "1"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	compressedData, err := hex.DecodeString(deflateDataLevel1)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginFlateProcessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), deflateDataLevel1)
		}
	}
}

func TestPluginFlateUnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(deflatedData)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginFlate()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(flateTestData)); c != 0 {
			t.Errorf("TestPluginFlateUnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), flateTestData)
		}
	}
}

func TestPluginFlateUnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(deflatedData)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginFlate()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(flateTestData)); c != 0 {
			t.Errorf("TestPluginFlateUnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), flateTestData)
		}
	}
}

func TestPluginFlateAddDefaultCliFunc(t *testing.T) {
	p := NewPluginFlate()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
