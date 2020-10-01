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
