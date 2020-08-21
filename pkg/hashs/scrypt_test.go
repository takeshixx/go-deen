package hashs

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

var scryptTestData = "verysecurepassword"
var scryptTestSaltHex = "7465737473616c74"
var scryptTestSalt = []byte{0xff, 0x12, 0x23, 0x45, 0xee, 0xff, 0x77, 0xff}

func TestPluginScryptProcess(t *testing.T) {
	p := NewPluginScrypt()
	r := strings.NewReader(scryptTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("TestPluginScryptProcess failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("0IlaUn16JVhuciJDk6ow51TFVcgG7T14dYlJsNsWFRQ=")); c != 0 {
		t.Errorf("TestPluginScryptProcess returned invalid data: %s", d)
	}
}

func TestPluginScryptProcessWithFlags(t *testing.T) {
	p := NewPluginScrypt()
	r := strings.NewReader(scryptTestData)
	testFlags := p.AddCliOptionsFunc(&p, []string{"-salt", scryptTestSaltHex})
	d, e := p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("TestPluginScryptProcessWithFlags failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("m39r9JZGQxVP76+lu4zGiFCbspLci15u9OTI9zlC1lU=")); c != 0 {
		t.Errorf("TestPluginScryptProcess returned invalid data: %s", d)
	}

	p = NewPluginScrypt()
	r = strings.NewReader(scryptTestData)
	testFlags = p.AddCliOptionsFunc(&p, []string{"-cost", fmt.Sprintf("%d", 1<<12), "-len", "96", "-salt", scryptTestSaltHex})
	d, e = p.ProcessStreamWithCliFlagsFunc(testFlags, r)
	if e != nil {
		t.Errorf("TestPluginScryptProcessWithFlags failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("CLf5VMJVqyLztPE4cK1fpoRbOwQDUSgYm4VWfxgdMvpH6dBbaJ2rD0+hRhC6vcbEaL0/XHQSJTFYifshfoIRh+B2RRhZKqeTpXqP+4jxhiuMVa1lgInQMlAflOQfSCaq")); c != 0 {
		t.Errorf("TestPluginScryptProcess returned invalid data: %s", d)
	}
}

func TestPluginScryptUsage(t *testing.T) {
	p := NewPluginScrypt()
	testFlags := p.AddCliOptionsFunc(&p, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	testFlags.Usage()
}
