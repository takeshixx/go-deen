package hashs

import (
	"bytes"
	"testing"
)

var shaTestData = []byte("deenshatest")

func TestNewPluginSHA1(t *testing.T) {
	p := NewPluginSHA1()
	r := bytes.NewReader(shaTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("SHA224Process failed: %s", e)
	}
	if c := bytes.Compare(d, []byte("c324da7d32853ffaeb6577f624753c7f0f2842c0")); c != 0 {
		t.Errorf("Invalid SHA1 data: %s", d)
	}
}
