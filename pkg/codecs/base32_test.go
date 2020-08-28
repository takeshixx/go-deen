package codecs

import (
	"bytes"
	"encoding/base32"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var b32InputData = "deentestdatastringextendedversion321xxx"
var b32InputDataProcessed = []byte("MRSWK3TUMVZXIZDBORQXG5DSNFXGOZLYORSW4ZDFMR3GK4TTNFXW4MZSGF4HQ6A=")
var b32InputDataProcessedHex = []byte("CHIMARJKCLPN8P31EHGN6T3ID5N6EPBOEHIMSP35CHR6ASJJD5NMSCPI65S7GU0=")

func TestNewPluginBase32(t *testing.T) {
	p := NewPluginBase32()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginBase32: %s", reflect.TypeOf(p))
	}
}

func TestPluginBase32ProcessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	plugin := NewPluginBase32()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b32InputDataProcessed); c != 0 {
			t.Errorf("Base32Process data wrong: %s != %s", destWriter.Bytes(), b32InputDataProcessed)
		}
	}
}

func TestPluginBase32ProcessBase32(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	encoding := base32.StdEncoding
	processBase32(encoding, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b32InputDataProcessed); c != 0 {
			t.Errorf("Base32Process data wrong: %s != %s", destWriter.Bytes(), b32InputDataProcessed)
		}
	}
}

func TestPluginBase32UnprocessDeenTaskFunc(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b32InputDataProcessed)
	plugin := NewPluginBase32()
	plugin.UnprocessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b32InputData)); c != 0 {
			t.Errorf("Base32ProcessUnprocessDeenTaskFunc data wrong: %s != %s", destWriter.Bytes(), b32InputData)
		}
	}
}

func TestPluginBase32UnprocessDeenTaskFuncInvalid(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	plugin := NewPluginBase32()
	plugin.UnprocessDeenTaskFunc(task)
	var err error
	select {
	case processErr := <-task.ErrChan:
		err = processErr
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b32InputDataProcessedHex); c == 0 {
			t.Errorf("Base32Process should have failed: %s == %s", destWriter.Bytes(), b32InputDataProcessed)
		}
	}
	if err == nil {
		t.Error("Invalid data should have triggered an error")
	}
}

func TestPluginBase32ProcessBase32HexEncoding(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	encoding := base32.HexEncoding
	processBase32(encoding, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b32InputDataProcessedHex); c != 0 {
			t.Errorf("Base32Process data wrong: %s != %s", destWriter.Bytes(), b32InputDataProcessed)
		}
	}
}

func TestPluginBase32AddDefaultCliFunc(t *testing.T) {
	plugin := NewPluginBase32()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-hex", "-no-pad"})
	if flags == nil {
		t.Error("Failed to create FlagSet")
	}
	if !helpers.IsBoolFlag(flags, "hex") {
		t.Error("hex should be true")
	}
	if !helpers.IsBoolFlag(flags, "no-pad") {
		t.Error("no-pad should be true")
	}
}

func TestPluginBase32ProcessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	plugin := NewPluginBase32()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-hex"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), b32InputDataProcessedHex); c != 0 {
			t.Errorf("TestPluginBase32ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), b32InputData)
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	plugin = NewPluginBase32()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-no-pad"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), bytes.TrimSuffix(b32InputDataProcessed, []byte("="))); c != 0 {
			t.Errorf("TestPluginBase32ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), b32InputData)
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(b32InputData)
	plugin = NewPluginBase32()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-hex", "-no-pad"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), bytes.TrimSuffix(b32InputDataProcessedHex, []byte("="))); c != 0 {
			t.Errorf("TestPluginBase32ProcessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), bytes.TrimSuffix(b32InputDataProcessedHex, []byte("=")))
		}
	}
}

func TestPluginBase32UnprocessDeenTaskWithFlags(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b32InputDataProcessed)
	plugin := NewPluginBase32()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b32InputData)); c != 0 {
			t.Errorf("TestPluginBase32UnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), b32InputData)
		}
	}

	destWriter = new(bytes.Buffer)
	task = types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(b32InputDataProcessedHex)
	plugin = NewPluginBase32()
	flags = helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-hex"})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(b32InputData)); c != 0 {
			t.Errorf("TestPluginBase32UnprocessDeenTaskWithFlags data wrong: %s != %s", destWriter.Bytes(), b32InputData)
		}
	}
}

func TestPluginBase32Usage(t *testing.T) {
	p := NewPluginBase32()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
