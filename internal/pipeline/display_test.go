package pipeline

import (
	"strings"
	"testing"
)

func TestStringsDisplay(t *testing.T) {
	got, capped := StringsDisplay([]byte{0x00, 'a', 'b', 'c', 0x00, 'd', 'e', 'e', 'n', 0xff, 't', 'e', 's', 't'})
	if capped {
		t.Fatal("small input should not be capped")
	}
	if got != "deen\ntest" {
		t.Fatalf("StringsDisplay() = %q, want %q", got, "deen\ntest")
	}
}

func TestStringsDisplayNoMatches(t *testing.T) {
	got, _ := StringsDisplay([]byte{0x00, 0x01, 'a', 'b', 'c'})
	if !strings.Contains(got, "no printable strings") {
		t.Fatalf("StringsDisplay() = %q, want no-match message", got)
	}
}

func TestFullDisplaysDoNotTruncate(t *testing.T) {
	data := []byte(strings.Repeat("deen", TextPreviewLimit/4+1))
	if got := TextDisplayFull(data); strings.Contains(got, "preview truncated") {
		t.Fatalf("TextDisplayFull() unexpectedly truncated")
	}
	if got := HexDisplayFull(data); strings.Contains(got, "preview truncated") {
		t.Fatalf("HexDisplayFull() unexpectedly truncated")
	}
	if got := StringsDisplayFull(data); strings.Contains(got, "preview truncated") {
		t.Fatalf("StringsDisplayFull() unexpectedly truncated")
	}
}
