package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var lzmaTestData = []byte("deenLzMa data onlyUUUSEfor Testing PURPOSES")
var lzmaTestDataCompressed = "5d00008000ffffffffffffffff0032196cdc4c8bd28be7a9dfcdd6221406244c46751fdf763f23209273abaaae7c90b600db9c28ab9958805e4cb94bcffffcf27000"
var lzma2TestDataCompressed = "01002a6465656e4c7a4d612064617461206f6e6c795555555345666f722054657374696e6720505552504f53455300"

func TestPluginLZMAProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(lzmaTestData)
	plugin := NewPluginLZMA()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(lzmaTestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginLZMAProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzmaTestDataCompressed)
		}
	}
}

func TestPluginLZMAUnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(lzmaTestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginLZMA()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(lzmaTestData)); c != 0 {
			t.Errorf("TestPluginLZMAUnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzmaTestData)
		}
	}
}

func TestPluginLZMA2ProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(lzmaTestData)
	plugin := NewPluginLZMA2()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(lzma2TestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginLZMA2ProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzma2TestDataCompressed)
		}
	}
}

func TestPluginLZMA2UnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(lzma2TestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginLZMA2()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(lzmaTestData)); c != 0 {
			t.Errorf("TestPluginLZMA2UnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzmaTestData)
		}
	}
}

func TestPluginLZMAAddDefaultCliFunc(t *testing.T) {
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w

	p := NewPluginLZMA()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	flags.Usage()

	p = NewPluginLZMA2()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	flags.Usage()
}
