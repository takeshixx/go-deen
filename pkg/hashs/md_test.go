package hashs

import (
	"bytes"
	"testing"
)

var mdTestData = []byte("deenmdtest")

func TestNewPluginMD4(t *testing.T) {
	p := NewPluginMD4()
	r := bytes.NewReader(mdTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("MD4Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("a635f9247276ff156bbbb3752db8a2b1")); c != 0 {
		t.Errorf("MD4Process invalid data: %s", d)
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
