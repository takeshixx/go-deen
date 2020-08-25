package compressions

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var gzipTestData = []byte("deenGzipTestdataaaa12333331 test data")
var gzipTestDataCompressed = "1f8b08000000000000ff4a494dcd73afca2c08492d2e49492c494c4c4c3434320601438592d4e212059020200000ffff3b35306525000000"
var gzipTestDataCompressedLevel1 = "1f8b08000000000004ff002500daff6465656e477a69705465737464617461616161313233333333333120746573742064617461010000ffff3b35306525000000"

func TestPluginGzipProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(gzipTestData)
	plugin := NewPluginGzip()
	plugin.ProcessDeenTaskFunc(task)
	compressedData, err := hex.DecodeString(gzipTestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginGzipProcessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), gzipTestDataCompressed)
		}
	}
}

func TestPluginGzipProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(gzipTestData)
	plugin := NewPluginGzip()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(&plugin, flags, []string{"-level", "1"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	compressedData, err := hex.DecodeString(gzipTestDataCompressedLevel1)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), compressedData); c != 0 {
			t.Errorf("TestPluginGzipProcessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), gzipTestDataCompressedLevel1)
		}
	}
}

func TestPluginGzipUnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(gzipTestDataCompressed)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginGzip()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(gzipTestData)); c != 0 {
			t.Errorf("TestPluginGzipUnprocessDeenTaskFunc data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), gzipTestData)
		}
	}
}

func TestPluginGzipUnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	compressedData, err := hex.DecodeString(gzipTestDataCompressedLevel1)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(compressedData)
	plugin := NewPluginGzip()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(&plugin, flags, []string{})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(gzipTestData)); c != 0 {
			t.Errorf("TestPluginGzipUnprocessDeenTaskWithFlags data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), gzipTestData)
		}
	}
}

func TestPluginGzipAddDefaultCliFunc(t *testing.T) {
	p := NewPluginGzip()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(&p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
