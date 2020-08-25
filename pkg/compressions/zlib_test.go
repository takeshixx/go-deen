package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var zlibTestData = []byte("deenZlIBTestData333 13123 wijdasd kanjsdn kajs as")
var zlibTestDataCompressed = "789c4a494dcd8bcaf1740a492d2e71492c493436365630343634325628cfcc4a492c4e51c84ecccb2a4ec953c84ecc2a56482c06040000ffff992a1087"
var zlibTestDataCompressedLevel1 = "7801003100ceff6465656e5a6c494254657374446174613333332031333132332077696a64617364206b616e6a73646e206b616a73206173010000ffff992a1087"

func TestPluginZlibProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(zlibTestData)
	plugin := NewPluginZlib()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(zlibTestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginZlibProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), zlibTestDataCompressed)
		}
	}
}

func TestPluginZlibProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(zlibTestData)
	plugin := NewPluginZlib()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-level", "1"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	compressedData, err := hex.DecodeString(zlibTestDataCompressedLevel1)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginZlibProcessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), zlibTestDataCompressedLevel1)
		}
	}
}

func TestPluginZlibUnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(zlibTestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginZlib()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(zlibTestData)); c != 0 {
			t.Errorf("TestPluginZlibUnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), zlibTestData)
		}
	}
}

func TestPluginZlibUnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(zlibTestDataCompressedLevel1)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginZlib()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(zlibTestData)); c != 0 {
			t.Errorf("TestPluginZlibUnprocessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), zlibTestData)
		}
	}
}

func TestPluginZlibAddDefaultCliFunc(t *testing.T) {
	p := NewPluginZlib()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
