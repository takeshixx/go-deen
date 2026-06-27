package codecs

import (
	"bytes"
	"testing"
)

func TestPluginUnicodeNormalizeForms(t *testing.T) {
	tests := []struct {
		name string
		args []string
		in   []byte
		want []byte
	}{
		{
			name: "nfc composes accent",
			args: []string{"--form", "nfc"},
			in:   []byte("Cafe\u0301"),
			want: []byte("Caf\u00e9"),
		},
		{
			name: "nfd decomposes accent",
			args: []string{"--form", "nfd"},
			in:   []byte("Caf\u00e9"),
			want: []byte("Cafe\u0301"),
		},
		{
			name: "nfkc normalizes compatibility character",
			args: []string{"--form", "nfkc"},
			in:   []byte("\u212b"),
			want: []byte("\u00c5"),
		},
		{
			name: "nfkd decomposes compatibility character",
			args: []string{"--form", "nfkd"},
			in:   []byte("\u212b"),
			want: []byte("A\u030a"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginUnicodeNormalize()
			got := runCodec(t, p.Process, p.RegisterFlags, tt.in, tt.args...)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("normalized bytes = % x, want % x", got, tt.want)
			}
		})
	}
}

func TestPluginUnicodeNormalizeAliasSelectsForm(t *testing.T) {
	p := NewPluginUnicodeNormalize()
	p.Command = "nfd"
	got := runCodec(t, p.Process, p.RegisterFlags, []byte("Caf\u00e9"))
	if !bytes.Equal(got, []byte("Cafe\u0301")) {
		t.Errorf("nfd alias normalized to %q", got)
	}
}

func TestPluginUnicodeNormalizeRejectsUnknownForm(t *testing.T) {
	p := NewPluginUnicodeNormalize()
	_, err := tryCodec(p.Process, p.RegisterFlags, []byte("test"), "--form", "weird")
	if err == nil {
		t.Fatal("expected unsupported normalization form error")
	}
}
