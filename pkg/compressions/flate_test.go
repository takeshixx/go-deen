package compressions

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var deflateTestData = []byte("deenblaketest")
var deflatedData = "4b494dcd4bca49cc4e2d492d2e0100"

func TestPluginDeflateProcessStreamFunc(t *testing.T) {
	p := NewPluginDeflate()
	r := bytes.NewReader(deflateTestData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("DeflateProcessStreamFunc failed: %s", e)
	}
	decoded, e := hex.DecodeString(deflatedData)
	if e != nil {
		t.Errorf("Decoding test data failed: %s", e)
	}
	t.Log(d)
	t.Log(decoded)
	if c := bytes.Compare(d, decoded); c != 0 {
		t.Errorf("DeflateProcessSteamFunc returned invalid data: %s", d)
	}
}
