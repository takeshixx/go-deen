package compressions

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var binData = []byte("4e1260f08e8e1b7d0a829cf82fb23cf95cc1a919cf2187d0276f4c00807faec1")
var bzipCompressedData = "425a68393141592653591a1c4b54000005fffec0101028208080040005000440008002d0810004002110002000c0004060200031400000000c34610c9a6268f136a9b94e3edcd171788eab42a3ad20964f821aa81f8bb9229c28480d0e25aa00"

func TestPluginBzip2Process(t *testing.T) {
	p := NewPluginBzip2()
	r := bytes.NewReader(binData)
	d, e := p.ProcessStreamFunc(r)
	if e != nil {
		t.Errorf("Bzip2ProcessStreamFunc failed: %s", e)
	}
	decoded, e := hex.DecodeString(bzipCompressedData)
	if e != nil {
		t.Errorf("Decoding test data failed: %s", e)
	}
	t.Log(d)
	t.Log(decoded)
	if c := bytes.Compare(d, decoded); c != 0 {
		t.Errorf("Bzip2ProcessSteamFunc returned invalid data: %s", d)
	}
}

func TestPluginBzip2Unprocess(t *testing.T) {

}
