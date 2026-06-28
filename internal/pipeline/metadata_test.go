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
	if m.Encoding != "ASCII / UTF-8" {
		t.Fatalf("encoding = %q, want ASCII / UTF-8", m.Encoding)
	}
	if m.Printable != 100 {
		t.Fatalf("printable = %d, want 100", m.Printable)
	}
	summary := m.Summary()
	for _, want := range []string{"11 bytes", "2 lines", "ASCII / UTF-8", "2.20x input"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q: %s", want, summary)
		}
	}
	fields := m.Fields()
	if fields[0].Label != "size" || fields[0].Value != "11 bytes" {
		t.Fatalf("unexpected first field: %#v", fields[0])
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

func TestDataMetadataSummaryIncludesLargeUnits(t *testing.T) {
	m := DataMetadata(make([]byte, LargeDataThreshold+1), 0)
	summary := m.Summary()
	for _, want := range []string{"1048577 bytes", "1.00 MB"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q: %s", want, summary)
		}
	}
	if strings.Contains(summary, "GB") {
		t.Fatalf("summary should not include GB below 1 GiB: %s", summary)
	}
}

func TestDataMetadataSummaryUsesGBForGigabyte(t *testing.T) {
	m := Metadata{Bytes: 1 << 30}
	summary := m.Summary()
	if !strings.Contains(summary, "1.00 GB") {
		t.Fatalf("summary should include GB: %s", summary)
	}
	if strings.Contains(summary, "MB") {
		t.Fatalf("summary should not include MB at 1 GiB: %s", summary)
	}
}

func TestDataMetadataEncodingHints(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		encoding string
		bom      string
	}{
		{
			name:     "utf8 bom",
			data:     []byte{0xef, 0xbb, 0xbf, 'h', 'i'},
			encoding: "UTF-8",
			bom:      "UTF-8",
		},
		{
			name:     "utf16le bom",
			data:     []byte{0xff, 0xfe, 0x41, 0x00},
			encoding: "UTF-16LE",
			bom:      "UTF-16LE",
		},
		{
			name:     "utf32be pattern",
			data:     []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x42},
			encoding: "likely UTF-32BE",
			bom:      "none",
		},
		{
			name:     "utf16le pattern",
			data:     []byte{0x41, 0x00, 0x42, 0x00},
			encoding: "likely UTF-16LE",
			bom:      "none",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := DataMetadata(tt.data, 0)
			if m.Encoding != tt.encoding {
				t.Fatalf("encoding = %q, want %q", m.Encoding, tt.encoding)
			}
			if m.BOM != tt.bom {
				t.Fatalf("bom = %q, want %q", m.BOM, tt.bom)
			}
			if !strings.Contains(m.Summary(), tt.encoding) {
				t.Fatalf("summary missing encoding %q: %s", tt.encoding, m.Summary())
			}
		})
	}
}

func TestDataMetadataEmpty(t *testing.T) {
	m := DataMetadata(nil, 0)
	if m.Bytes != 0 || m.Lines != 0 || m.Printable != 0 || m.Entropy != 0 {
		t.Fatalf("unexpected empty metadata: %#v", m)
	}
}
