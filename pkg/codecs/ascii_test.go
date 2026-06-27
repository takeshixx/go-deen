package codecs

import "testing"

func TestPluginASCII(t *testing.T) {
	tests := []struct {
		name string
		args []string
		in   []byte
		want []byte
	}{
		{
			name: "strict ascii",
			in:   []byte("deen"),
			want: []byte("deen"),
		},
		{
			name: "replace",
			args: []string{"--mode", "replace"},
			in:   []byte("Café 😀"),
			want: []byte("Caf? ?"),
		},
		{
			name: "strip",
			args: []string{"--mode", "strip"},
			in:   []byte("Café 😀"),
			want: []byte("Caf "),
		},
		{
			name: "escape",
			args: []string{"--mode", "escape"},
			in:   []byte("Café 😀"),
			want: []byte("Caf\\u00E9 \\U0001F600"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPluginASCII()
			assertCodec(t, p, p.Process, tt.in, tt.want, tt.args...)
		})
	}
}

func TestPluginASCIIStrictRejectsNonASCII(t *testing.T) {
	p := NewPluginASCII()
	if _, err := tryCodec(p.Process, p.RegisterFlags, []byte("é")); err == nil {
		t.Fatal("expected strict ASCII conversion to reject non-ASCII text")
	}
}

func TestPluginASCIIRejectsInvalidUTF8(t *testing.T) {
	p := NewPluginASCII()
	if _, err := tryCodec(p.Process, p.RegisterFlags, []byte{0xff}); err == nil {
		t.Fatal("expected ASCII conversion to reject invalid UTF-8")
	}
}
