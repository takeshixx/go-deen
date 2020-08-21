package compressions

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var flateTestData = []byte("deenblaketest\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98")
var deflatedData = "789c4b494dcd4bca49cc4e2d492d2ed9bbc9768fc2a39e1900646309ea"

func TestPluginDeflateProcessStreamFunc(t *testing.T) {
	p := NewPluginFlate()
	r := bytes.NewReader(flateTestData)
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
