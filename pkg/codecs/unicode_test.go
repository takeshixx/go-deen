package codecs

import (
	"bytes"
	"testing"
)

func TestPluginUnicodeRoundTrip(t *testing.T) {
	input := []byte("deen unicode test 123")
	for _, command := range []string{"utf8", "utf16", "utf16le", "utf16be", "utf32", "utf32le", "utf32be", "euckr"} {
		p := NewPluginUnicode()
		p.Command = command
		encoded := runCodec(t, p.Process, p.RegisterFlags, input)
		decoded := runCodec(t, p.Unprocess, p.RegisterFlags, encoded)
		if !bytes.Equal(decoded, input) {
			t.Errorf("%s round-trip mismatch: got %q, want %q", command, decoded, input)
		}
	}
}

func TestPluginUnicodeExplicitEndianOutput(t *testing.T) {
	input := []byte("A\n")
	tests := []struct {
		name    string
		command string
		args    []string
		want    []byte
	}{
		{
			name:    "utf16le alias",
			command: "utf16le",
			want:    []byte{0x41, 0x00, 0x0a, 0x00},
		},
		{
			name:    "utf16be alias",
			command: "utf16be",
			want:    []byte{0x00, 0x41, 0x00, 0x0a},
		},
		{
			name:    "utf16 flag big",
			command: "utf16",
			args:    []string{"--big"},
			want:    []byte{0x00, 0x41, 0x00, 0x0a},
		},
		{
			name:    "utf32le alias",
			command: "utf32le",
			want:    []byte{0x41, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00},
		},
		{
			name:    "utf32be alias",
			command: "utf32be",
			want:    []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x0a},
		},
		{
			name:    "utf32 flag big",
			command: "utf32",
			args:    []string{"--big"},
			want:    []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x0a},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicode()
			p.Command = tt.command
			assertCodec(t, p, p.Process, input, tt.want, tt.args...)
		})
	}
}

func TestPluginUnicodeEncodingOption(t *testing.T) {
	p := NewPluginUnicode()
	p.Command = "unicode"
	encoded := runCodec(t, p.Process, p.RegisterFlags, []byte("A"), "--encoding", "utf16be")
	if !bytes.Equal(encoded, []byte{0x00, 0x41}) {
		t.Fatalf("encoded = % x, want 00 41", encoded)
	}
	decoded := runCodec(t, p.Unprocess, p.RegisterFlags, encoded, "--encoding", "utf16be")
	if !bytes.Equal(decoded, []byte("A")) {
		t.Fatalf("decoded = %q, want A", decoded)
	}
}

func TestPluginUnicodeAliasOverridesEncodingOption(t *testing.T) {
	p := NewPluginUnicode()
	p.Command = "utf16le"
	encoded := runCodec(t, p.Process, p.RegisterFlags, []byte("A"), "--encoding", "utf16be")
	if !bytes.Equal(encoded, []byte{0x41, 0x00}) {
		t.Fatalf("encoded = % x, want 41 00", encoded)
	}
}

func TestPluginUnicodeBOMOutput(t *testing.T) {
	input := []byte("A")
	tests := []struct {
		name    string
		command string
		want    []byte
	}{
		{
			name:    "utf16le",
			command: "utf16le",
			want:    []byte{0xff, 0xfe, 0x41, 0x00},
		},
		{
			name:    "utf16be",
			command: "utf16be",
			want:    []byte{0xfe, 0xff, 0x00, 0x41},
		},
		{
			name:    "utf32le",
			command: "utf32le",
			want:    []byte{0xff, 0xfe, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00},
		},
		{
			name:    "utf32be",
			command: "utf32be",
			want:    []byte{0x00, 0x00, 0xfe, 0xff, 0x00, 0x00, 0x00, 0x41},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicode()
			p.Command = tt.command
			assertCodec(t, p, p.Process, input, tt.want, "--bom", "use")
		})
	}
}

func TestPluginUnicodeDecodesExplicitEndian(t *testing.T) {
	tests := []struct {
		name    string
		command string
		input   []byte
	}{
		{
			name:    "utf16le",
			command: "utf16le",
			input:   []byte{0x41, 0x00, 0x0a, 0x00},
		},
		{
			name:    "utf16be",
			command: "utf16be",
			input:   []byte{0x00, 0x41, 0x00, 0x0a},
		},
		{
			name:    "utf32le",
			command: "utf32le",
			input:   []byte{0x41, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00},
		},
		{
			name:    "utf32be",
			command: "utf32be",
			input:   []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x0a},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicode()
			p.Command = tt.command
			assertCodec(t, p, p.Unprocess, tt.input, []byte("A\n"))
		})
	}
}

func TestPluginUnicodeLegacySingleByteOutput(t *testing.T) {
	tests := []struct {
		name    string
		command string
		input   []byte
		want    []byte
	}{
		{
			name:    "latin1",
			command: "latin1",
			input:   []byte("Caf\u00e9"),
			want:    []byte{0x43, 0x61, 0x66, 0xe9},
		},
		{
			name:    "windows1252 smart quotes",
			command: "windows1252",
			input:   []byte("\u201cdeen\u201d"),
			want:    []byte{0x93, 0x64, 0x65, 0x65, 0x6e, 0x94},
		},
		{
			name:    "koi8-r",
			command: "koi8r",
			input:   []byte("\u042f"),
			want:    []byte{0xf1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicode()
			p.Command = tt.command
			assertCodec(t, p, p.Process, tt.input, tt.want)
			assertCodec(t, p, p.Unprocess, tt.want, tt.input)
		})
	}
}

func TestPluginUnicodeLegacyMultibyteRoundTrip(t *testing.T) {
	tests := []struct {
		command string
		input   []byte
	}{
		{"shiftjis", []byte("\u65e5\u672c\u8a9e")},
		{"eucjp", []byte("\u65e5\u672c\u8a9e")},
		{"gbk", []byte("\u4e2d\u6587")},
		{"gb18030", []byte("\U0001f600\u4e2d\u6587")},
		{"big5", []byte("\u7e41\u9ad4\u4e2d\u6587")},
		{"euckr", []byte("\ud55c\uad6d\uc5b4")},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			p := NewPluginUnicode()
			p.Command = tt.command
			encoded := runCodec(t, p.Process, p.RegisterFlags, tt.input)
			decoded := runCodec(t, p.Unprocess, p.RegisterFlags, encoded)
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("%s round-trip mismatch: got %q, want %q", tt.command, decoded, tt.input)
			}
		})
	}
}
