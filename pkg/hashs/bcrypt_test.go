package hashs

import (
	"bytes"
	"os"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"golang.org/x/crypto/bcrypt"
)

var bcryptTestData = []byte("deenpassword")

func TestPluginBcryptProcess(t *testing.T) {
	p := NewPluginBcrypt()
	r := bytes.NewReader(bcryptTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestPluginBcryptProcess failed: %s", e)
	}
	e = bcrypt.CompareHashAndPassword(d, bcryptTestData)
	if e != nil {
		t.Error("TestPluginBcryptProcess returned wrong hash")
	}
}

func TestPluginBcryptProcessWithFlags(t *testing.T) {
	p := NewPluginBcrypt()
	r := bytes.NewReader(bcryptTestData)
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	d, e := p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("TestPluginBcryptProcessWithFlags failed: %s", e)
	}
	e = bcrypt.CompareHashAndPassword(d, bcryptTestData)
	if e != nil {
		t.Error("TestPluginBcryptProcessWithFlags returned wrong hash")
	}

	p = NewPluginBcrypt()
	r = bytes.NewReader(bcryptTestData)
	flags = helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{"-cost", "7"})
	d, e = p.ProcessStreamWithCliFlagsFunc(flags, r)
	if e != nil {
		t.Errorf("TestPluginBcryptProcessWithFlags failed: %s", e)
	}
	e = bcrypt.CompareHashAndPassword(d, bcryptTestData)
	if e != nil {
		t.Error("TestPluginBcryptProcessWithFlags returned wrong hash")
	}
}

func TestPluginBcryptUsage(t *testing.T) {
	p := NewPluginBcrypt()
	flags := helpers.DefaultFlagSet()
	testFlags := p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	testFlags.Usage()
}
