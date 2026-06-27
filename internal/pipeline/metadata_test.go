package pipeline

import (
	"strings"
	"testing"
)

func TestDataMetadataText(t *testing.T) {
	m := DataMetadata([]byte("hello\nworld"), 5)
	if m.Bytes != 11 {
		t.Fatalf("bytes = %d, want 11", m.Bytes)
	}
	if m.Lines != 2 {
		t.Fatalf("lines = %d, want 2", m.Lines)
	}
	if !m.UTF8 {
		t.Fatal("expected valid UTF-8")
	}
	if m.Printable != 100 {
		t.Fatalf("printable = %d, want 100", m.Printable)
	}
	summary := m.Summary()
	for _, want := range []string{"11 bytes", "2 lines", "utf-8", "2.20x input"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q: %s", want, summary)
		}
	}
}

func TestDataMetadataBinary(t *testing.T) {
	m := DataMetadata([]byte{0xff, 0x00, 0x01}, 0)
	if m.UTF8 {
		t.Fatal("expected invalid UTF-8")
	}
	if m.Printable != 0 {
		t.Fatalf("printable = %d, want 0", m.Printable)
	}
	if !strings.Contains(m.Summary(), "binary") {
		t.Fatalf("summary should describe binary data: %s", m.Summary())
	}
}

func TestDataMetadataEmpty(t *testing.T) {
	m := DataMetadata(nil, 0)
	if m.Bytes != 0 || m.Lines != 0 || m.Printable != 0 || m.Entropy != 0 {
		t.Fatalf("unexpected empty metadata: %#v", m)
	}
}
