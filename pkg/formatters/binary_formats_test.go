package formatters

import (
	"bytes"
	"strings"
	"testing"
)

func TestMessagePackRoundTrip(t *testing.T) {
	p := NewPluginMessagePack()
	packed := runFormat(t, p.Process, p.RegisterFlags, []byte(`{"name":"deen","ok":true}`))
	json := runFormat(t, p.Unprocess, p.RegisterFlags, packed)
	if !strings.Contains(string(json), `"name": "deen"`) || !strings.Contains(string(json), `"ok": true`) {
		t.Fatalf("MessagePack JSON unexpected:\n%s", string(json))
	}
}

func TestCBORRoundTrip(t *testing.T) {
	p := NewPluginCBOR()
	packed := runFormat(t, p.Process, p.RegisterFlags, []byte(`{"name":"deen","ok":true}`))
	json := runFormat(t, p.Unprocess, p.RegisterFlags, packed)
	if !strings.Contains(string(json), `"name": "deen"`) || !strings.Contains(string(json), `"ok": true`) {
		t.Fatalf("CBOR JSON unexpected:\n%s", string(json))
	}
}

func TestQRRoundTrip(t *testing.T) {
	p := NewPluginQR()
	png := runFormat(t, p.Process, p.RegisterFlags, []byte("deen"), "-size", "128")
	if !bytes.HasPrefix(png, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("QR output is not PNG")
	}
	decoded := runFormat(t, p.Unprocess, p.RegisterFlags, png)
	if string(decoded) != "deen" {
		t.Fatalf("QR decoded = %q, want deen", string(decoded))
	}
}
