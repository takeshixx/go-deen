package codecs

import (
	"os"
	"reflect"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func TestNewPluginUnicode(t *testing.T) {
	p := NewPluginUnicode()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginUnicode: %s", reflect.TypeOf(p))
	}
}

func TestPluginUnicodeUsage(t *testing.T) {
	p := NewPluginUnicode()
	p.Command = "utf8"
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()

	p = NewPluginUnicode()
	p.Command = "utf16"
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err = os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()

	p = NewPluginUnicode()
	p.Command = "utf32"
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err = os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()

	p = NewPluginUnicode()
	p.Command = "euckr"
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err = os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}

var unicodeTestData = []byte{0xfc, 0xc8}
var unicodeTestDataUTF16 = []byte("주")

/* func TestPluginUnicodeUTF16Process(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = bytes.NewReader(unicodeTestData)
	plugin := NewPluginUnicode()
	plugin.ProcessDeenTaskFunc(task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), hexInputDataProcessed); c != 0 {
			t.Errorf("TestPluginUnicodeUTF16Process data wrong: %s != %s", hex.EncodeToString(destWriter.Bytes()), unicodeTestDataUTF16)
		}
	}
} */
