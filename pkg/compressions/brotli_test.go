package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var brotliTestData = "someinterestingdata to compressssssss"
var brotliTestDataCompressedHex = "1b240000c46d6c5df70f898639346731859046b213752c7ec5617abf5d8ba831962f00"
var brotliTestDataModifiedHex = "a120010000509b4bd55c657b86411d9e21520043a7a2562336ea5d5fb76d3b9d7f040000"

func TestPluginBrotliProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(brotliTestData)
	plugin := NewPluginBrotli()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(brotliTestDataCompressedHex)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginBrotliProcessDeenTaskFunc data wrong: %s != %s", destWriter.Bytes(), compressedData)
		}
	}
}

func TestPluginBrotliProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(brotliTestData)
	plugin := NewPluginBrotli()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-level", "3", "-lgwin", "2"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	compressedData, err := hex.DecodeString(brotliTestDataModifiedHex)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginBrotliProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), compressedData)
		}
	}
}

func TestPluginBrotliUnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(brotliTestDataModifiedHex)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginBrotli()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-level", "3", "-lgwin", "2"})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(brotliTestData)); c != 0 {
			t.Errorf("TestPluginBrotliUnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), []byte(brotliTestData))
		}
	}
}

func TestPluginBrotliUnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(brotliTestDataCompressedHex)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginBrotli()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(brotliTestData)); c != 0 {
			t.Errorf("TestPluginBrotliUnprocessDeenTaskFunc data wrong: %s != %s", destWriter.Bytes(), []byte(brotliTestData))
		}
	}
}

func TestPluginBrotliAddDefaultCliFunc(t *testing.T) {
	p := NewPluginBrotli()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
