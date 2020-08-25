package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var bzip2TestData = []byte("4e1260f08e8e1b7d0a829cf82fb23cf95cc1a919cf2187d0276f4c00807faec1")
var bzip2CompressedData = "425a6836314159265359e63fefa100000a89007fe03f002000488a9e794cd4d4c7aa83553f541feaa7ea8fcf50e76c316e577225deee168a1ec6aaa1a5b249a39118f4fcf91a262ba9c84ce177245385090e63fefa10"
var bzip2CompressedDataLevel1 = "425a6831314159265359e63fefa100000a89007fe03f002000488a9e794cd4d4c7aa83553f541feaa7ea8fcf50e76c316e577225deee168a1ec6aaa1a5b249a39118f4fcf91a262ba9c84ce177245385090e63fefa10"

func TestPluginBZip2ProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(bzip2TestData)
	plugin := NewPluginBzip2()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(bzip2CompressedData)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginBZip2ProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), bzip2CompressedData)
		}
	}
}

func TestPluginBZip2ProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(bzip2TestData)
	plugin := NewPluginBzip2()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(&plugin, flags, []string{"-level", "1"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	compressedData, err := hex.DecodeString(bzip2CompressedDataLevel1)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginBZip2ProcessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), bzip2CompressedDataLevel1)
		}
	}
}

func TestPluginBZip2UnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(bzip2CompressedData)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginBzip2()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(bzip2TestData)); c != 0 {
			t.Errorf("TestPluginBZip2UnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), bzip2TestData)
		}
	}
}

func TestPluginBZip2UnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(bzip2CompressedData)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginBzip2()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(&plugin, flags, []string{})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(bzip2TestData)); c != 0 {
			t.Errorf("TestPluginBZip2UnprocessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), bzip2TestData)
		}
	}
}

func TestPluginBZip2AddDefaultCliFunc(t *testing.T) {
	p := NewPluginBzip2()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
