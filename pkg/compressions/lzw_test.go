package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var lzwTestData = []byte("deenLZWLZWLzwlzwtest data aiaosdoji OOAJDonaoasdoi asi")
var lzwCompressedData = "64ca9471c344cb158308f5dc61b3904e99397440900943270c883069c2bc9943e68d9a34209e3c09a284c81b371bc3747c1372659a80"
var lzwCompressedDataFlags = "6465656e4c5a5786887a776c8b746573742064617461206169616f73646f6a69204f4f414a446f6e9b619d6fa1ab6981"

func TestPluginLzwProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(lzwTestData)
	plugin := NewPluginLzw()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(lzwCompressedData)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginLzwProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzwCompressedData)
		}
	}
}

func TestPluginLzwProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(lzwTestData)
	plugin := NewPluginLzw()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(&plugin, flags, []string{"-order", "1", "-lit-width", "7"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	compressedData, err := hex.DecodeString(lzwCompressedDataFlags)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginLzwProcessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzwCompressedDataFlags)
		}
	}
}

func TestPluginLzwUnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(lzwCompressedData)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginLzw()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(lzwTestData)); c != 0 {
			t.Errorf("TestPluginLzwUnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzwTestData)
		}
	}
}

func TestPluginLzwUnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(lzwCompressedDataFlags)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginLzw()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(&plugin, flags, []string{"-order", "1", "-lit-width", "7"})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(lzwTestData)); c != 0 {
			t.Errorf("TestPluginLzwUnprocessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), lzwTestData)
		}
	}
}

func TestPluginLzwAddDefaultCliFunc(t *testing.T) {
	p := NewPluginLzw()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
