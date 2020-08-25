package hashs

import (
	"bytes"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var mdTestData = []byte("deenmdtest")

func TestNewPluginMD4(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(mdTestData)
	plugin := NewPluginMD4()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte("a635f9247276ff156bbbb3752db8a2b1")); c != 0 {
			t.Errorf("TestPluginBase85ProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), []byte("a635f9247276ff156bbbb3752db8a2b1"))
		}
	}
}

func TestNewPluginMD5(t *testing.T) {
	p := NewPluginMD5()
	r := bytes.NewReader(mdTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("MD5Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("204778cb29e3c5ffa8037312a6ea2a56")); c != 0 {
		t.Errorf("MD5Process invalid data: %s", d)
	}
}

func TestNewPluginRIPEMD160(t *testing.T) {
	p := NewPluginRIPEMD160()
	r := bytes.NewReader(mdTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("RIPEMD160Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("8bcf2832074848df4727aeb275e4905f2213814e")); c != 0 {
		t.Errorf("RIPEMD160Process invalid data: %s", d)
	}
}

func TestMDUsage(t *testing.T) {
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w

	p := NewPluginMD4()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()

	p = NewPluginMD5()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()

	p = NewPluginRIPEMD160()
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	flags.Usage()
}
