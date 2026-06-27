package misc

import (
	"bytes"
	"flag"
	"testing"
)

func TestDNSNameRoundTrip(t *testing.T) {
	p := NewPluginDNS()
	wire := runMisc(t, p.RegisterFlags, p.Process, []byte("www.example.com."))
	want := []byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}
	if !bytes.Equal(wire, want) {
		t.Fatalf("DNS wire = %x, want %x", wire, want)
	}
	var out bytes.Buffer
	if err := p.Unprocess(bytes.NewReader(wire), &out, flag.NewFlagSet("dns", flag.ContinueOnError)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "www.example.com." {
		t.Fatalf("decoded DNS name = %q", out.String())
	}
}

func TestDNSRejectsBadName(t *testing.T) {
	p := NewPluginDNS()
	if err := p.Process(bytes.NewReader([]byte("bad..name")), &bytes.Buffer{}, flag.NewFlagSet("dns", flag.ContinueOnError)); err == nil {
		t.Fatal("expected bad DNS name to fail")
	}
}
