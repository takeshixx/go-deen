package hashs

import (
	"bytes"
	"testing"
)

func TestNewPluginSHA224(t *testing.T) {
	p := NewPluginSHA224()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("SHA224Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("b1dc6360c8531f1b465a484dbea1c5cec454ba3ca29c6eb4cd5ae406")); c != 0 {
		t.Errorf("Invalid SHA224 data: %s", d)
	}
}

func TestNewPluginSHA256(t *testing.T) {
	p := NewPluginSHA256()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("SHA256Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("4ada38e80388198b04707df9c7bc6f2d2c3614fc26e7bbf53494008204d80519")); c != 0 {
		t.Errorf("Invalid SHA256 data: %s", d)
	}
}

func TestNewPluginSHA384(t *testing.T) {
	p := NewPluginSHA384()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("SHA384Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("e78f30f30f042989efccc643fb310aef66f7602587d92be0657fcb080ab6bf9dea1df62389e70882812dc446587ea7b8")); c != 0 {
		t.Errorf("Invalid SHA384 data: %s", d)
	}
}

func TestNewPluginSHA512(t *testing.T) {
	p := NewPluginSHA512()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("SHA512Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("695359c0ba4b7cb76c5287e14c5f2d5284bfa0b5df81dbb2abfab080221019ed9de0a3f3d4307772cf8bc40c16930d4f1b2a0bd0d81e8a9bed2290f588d2d90b")); c != 0 {
		t.Errorf("Invalid SHA512 data: %s", d)
	}
}
