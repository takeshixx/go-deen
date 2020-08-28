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

var b64InputData = "asd123<<<<>>>>deentestdata23xxxx"
var b64InputDataProcessed = []byte("YXNkMTIzPDw8PD4+Pj5kZWVudGVzdGRhdGEyM3h4eHg=")
var b64InputDataProcessedURL = []byte("YXNkMTIzPDw8PD4-Pj5kZWVudGVzdGRhdGEyM3h4eHg=")

func TestNewPluginBase64(t *testing.T) {
	p := NewPluginBase64()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginBase64: %s", reflect.TypeOf(p))
	}
}

func TestPluginBase64ProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b64InputData)
	plugin := NewPluginBase64()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b64InputDataProcessed); c != 0 {
			t.Errorf("TestPluginBase64ProcessDeenTaskFunc data wrong: %s != %s", destWriter.Bytes(), b64InputDataProcessed)
		}
	}
}

func TestPluginBase64UnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b64InputDataProcessed)
	plugin := NewPluginBase64()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b64InputData)); c != 0 {
			t.Errorf("TestPluginBase64UnprocessDeenTaskFunc data wrong: %s != %s", destWriter.Bytes(), b64InputData)
		}
	}
}

func TestPluginBase64ProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b64InputData)
	plugin := NewPluginBase64()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{""})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b64InputDataProcessed); c != 0 {
			t.Errorf("TestPluginBase64ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), b64InputDataProcessed)
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b64InputData)
	plugin = NewPluginBase64()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-url"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b64InputDataProcessedURL); c != 0 {
			t.Errorf("TestPluginBase64ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), b64InputDataProcessedURL)
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b64InputData)
	plugin = NewPluginBase64()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-raw"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), bytes.ReplaceAll(b64InputDataProcessed, []byte("="), []byte(""))); c != 0 {
			t.Errorf("TestPluginBase64ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), bytes.ReplaceAll(b64InputDataProcessed, []byte("="), []byte("")))
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b64InputData)
	plugin = NewPluginBase64()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-raw", "-url"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), bytes.ReplaceAll(b64InputDataProcessedURL, []byte("="), []byte(""))); c != 0 {
			t.Errorf("TestPluginBase64ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), bytes.ReplaceAll(b64InputDataProcessedURL, []byte("="), []byte("")))
		}
	}
}

func TestPluginBase64UnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b64InputDataProcessed)
	plugin := NewPluginBase64()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{""})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b64InputData)); c != 0 {
			t.Errorf("TestPluginBase64UnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), []byte(b64InputData))
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b64InputDataProcessedURL)
	plugin = NewPluginBase64()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-url"})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b64InputData)); c != 0 {
			t.Errorf("TestPluginBase64UnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), []byte(b64InputData))
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(bytes.ReplaceAll(b64InputDataProcessed, []byte("="), []byte("")))
	plugin = NewPluginBase64()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-raw"})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b64InputData)); c != 0 {
			t.Errorf("TestPluginBase64UnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), []byte(b64InputData))
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(bytes.ReplaceAll(b64InputDataProcessedURL, []byte("="), []byte("")))
	plugin = NewPluginBase64()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-raw", "-url"})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b64InputData)); c != 0 {
			t.Errorf("TestPluginBase64UnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), []byte(b64InputData))
		}
	}
}

func TestPluginBase64Usage(t *testing.T) {
	p := NewPluginBase64()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
