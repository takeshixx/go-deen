package codecs

import (
	"strings"
	"testing"
)

func TestPluginUnicodeInspectUTF8(t *testing.T) {
	p := NewPluginUnicodeInspect()
	got := string(runCodec(t, p.Process, p.RegisterFlags, []byte("hello, 世界")))
	for _, want := range []string{
		"likely: UTF-8",
		"bom: none",
		"utf-8 valid: true",
		"invalid utf-8 bytes: 0",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("unicode-inspect output missing %q:\n%s", want, got)
		}
	}
}

func TestPluginUnicodeInspectBOMs(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "utf8 bom",
			input: []byte{0xef, 0xbb, 0xbf, 'A'},
			want:  "likely: UTF-8 with BOM",
		},
		{
			name:  "utf16le bom",
			input: []byte{0xff, 0xfe, 0x41, 0x00},
			want:  "likely: UTF-16LE with BOM",
		},
		{
			name:  "utf16be bom",
			input: []byte{0xfe, 0xff, 0x00, 0x41},
			want:  "likely: UTF-16BE with BOM",
		},
		{
			name:  "utf32le bom",
			input: []byte{0xff, 0xfe, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00},
			want:  "likely: UTF-32LE with BOM",
		},
		{
			name:  "utf32be bom",
			input: []byte{0x00, 0x00, 0xfe, 0xff, 0x00, 0x00, 0x00, 0x41},
			want:  "likely: UTF-32BE with BOM",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicodeInspect()
			got := string(runCodec(t, p.Process, p.RegisterFlags, tt.input))
			if !strings.Contains(got, tt.want) {
				t.Errorf("unicode-inspect output missing %q:\n%s", tt.want, got)
			}
		})
	}
}

func TestPluginUnicodeInspectLikelyEndianWithoutBOM(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "utf16le",
			input: []byte{0x41, 0x00, 0x42, 0x00},
			want:  "likely: likely UTF-16LE",
		},
		{
			name:  "utf16be",
			input: []byte{0x00, 0x41, 0x00, 0x42},
			want:  "likely: likely UTF-16BE",
		},
		{
			name:  "utf32le",
			input: []byte{0x41, 0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00},
			want:  "likely: likely UTF-32LE",
		},
		{
			name:  "utf32be",
			input: []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x42},
			want:  "likely: likely UTF-32BE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicodeInspect()
			got := string(runCodec(t, p.Process, p.RegisterFlags, tt.input))
			if !strings.Contains(got, tt.want) {
				t.Errorf("unicode-inspect output missing %q:\n%s", tt.want, got)
			}
		})
	}
}
