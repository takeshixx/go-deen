package pipeline

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"net/url"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/vmihailenco/msgpack/v5"
)

func TestSuggestionsDetectCommonTransforms(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		plugin    string
		unprocess bool
	}{
		{"base64", []byte("dGVzdA=="), "base64", true},
		{"hex", []byte("74657374"), "hex", true},
		{"url", []byte("hello%20world"), "url", true},
		{"html", []byte("Tom &amp; Jerry"), "html", true},
		{"json", []byte(`{"ok":true}`), "json", false},
		{"xml", []byte(`<root><ok>true</ok></root>`), "xml", false},
		{"gzip", []byte{0x1f, 0x8b, 0x08, 0x00}, "gzip", true},
		{"zlib", []byte{0x78, 0x9c, 0x00}, "zlib", true},
		{"protobuf", []byte{0x08, 0x96, 0x01}, "protobuf", false},
		{"uuid", []byte("550e8400-e29b-41d4-a716-446655440000"), "uuid", false},
		{"asn1", []byte{0x30, 0x03, 0x02, 0x01, 0x2a}, "asn1", false},
		{"dns", []byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}, "dns", true},
		{"magic", []byte("%PDF-1.7\n"), "magic", false},
		{"bininspect elf", []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01}, "bininspect", false},
		{"bininspect pe", []byte("MZ\x90\x00"), "bininspect", false},
		{"bininspect macho", []byte{0xcf, 0xfa, 0xed, 0xfe}, "bininspect", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !hasSuggestion(Suggestions(tt.input), tt.plugin, tt.unprocess) {
				t.Fatalf("missing suggestion %s/%v in %#v", tt.plugin, tt.unprocess, Suggestions(tt.input))
			}
		})
	}
}

func TestSuggestionsDetectBinaryStructuredFormats(t *testing.T) {
	msg, err := msgpack.Marshal(map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasSuggestion(Suggestions(msg), "msgpack", true) {
		t.Fatalf("missing msgpack suggestion: %#v", Suggestions(msg))
	}
	cb, err := cbor.Marshal(map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasSuggestion(Suggestions(cb), "cbor", true) {
		t.Fatalf("missing cbor suggestion: %#v", Suggestions(cb))
	}
}

func TestSuggestionsIgnoreBinaryStructuredScalars(t *testing.T) {
	msg, err := msgpack.Marshal([]byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatal(err)
	}
	if hasSuggestion(Suggestions(msg), "msgpack", true) {
		t.Fatalf("unexpected msgpack suggestion for scalar binary: %#v", Suggestions(msg))
	}
	cb, err := cbor.Marshal([]byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatal(err)
	}
	if hasSuggestion(Suggestions(cb), "cbor", true) {
		t.Fatalf("unexpected cbor suggestion for scalar binary: %#v", Suggestions(cb))
	}
}

func TestSuggestionsDetectTextEncodings(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		encoding string
	}{
		{
			name:     "utf16le",
			input:    []byte{0x41, 0x00, 0x42, 0x00},
			encoding: "utf16le",
		},
		{
			name:     "utf16be",
			input:    []byte{0x00, 0x41, 0x00, 0x42},
			encoding: "utf16be",
		},
		{
			name:     "utf32le",
			input:    []byte{0x41, 0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00},
			encoding: "utf32le",
		},
		{
			name:     "utf32be",
			input:    []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x42},
			encoding: "utf32be",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := Suggestions(tt.input)
			if !hasSuggestionOption(suggestions, "unicode", true, "encoding", tt.encoding) {
				t.Fatalf("missing unicode decode suggestion with encoding=%s in %#v", tt.encoding, suggestions)
			}
			if !hasSuggestion(suggestions, "unicode-inspect", false) {
				t.Fatalf("missing unicode-inspect suggestion in %#v", suggestions)
			}
		})
	}
}

func TestAddStepWithOptions(t *testing.T) {
	p := New()
	p.SetSource([]byte{0x00, 0x41})
	p.AddStepWithOptions("unicode", true, map[string]string{"encoding": "utf16be"})
	if got := string(p.Result()); got != "A" {
		t.Fatalf("result = %q, want A", got)
	}
}

func TestSuggestionsDetectAutomatedDecodeChain(t *testing.T) {
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write([]byte(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	input := []byte(url.QueryEscape(base64.StdEncoding.EncodeToString(gz.Bytes())))

	suggestions := Suggestions(input)
	chain := findSuggestionChain(suggestions, []SuggestionStep{
		{Plugin: "url", Unprocess: true},
		{Plugin: "base64", Unprocess: true},
		{Plugin: "gzip", Unprocess: true},
		{Plugin: "json", Unprocess: false},
	})
	if chain == nil {
		t.Fatalf("missing automated decode chain in %#v", suggestions)
	}
	if chain.Confidence == 0 || chain.Preview == "" {
		t.Fatalf("chain missing confidence or preview: %#v", chain)
	}

	p := New()
	p.SetSource(input)
	p.AddSuggestion(*chain)
	if got := string(p.Result()); !strings.Contains(got, `"ok": true`) {
		t.Fatalf("AddSuggestion result = %q", got)
	}
	if len(p.Steps()) != 4 {
		t.Fatalf("AddSuggestion added %d steps, want 4", len(p.Steps()))
	}
}

func TestSuggestionsJSONPreviewHasNoANSIColor(t *testing.T) {
	suggestions := Suggestions([]byte(`{"ok":true}`))
	if !hasSuggestionOption(suggestions, "json", false, "no-color", "true") {
		t.Fatalf("missing no-color JSON suggestion in %#v", suggestions)
	}

	input := []byte(url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(`{"ok":true}`))))
	chain := findSuggestionChain(Suggestions(input), []SuggestionStep{
		{Plugin: "url", Unprocess: true},
		{Plugin: "base64", Unprocess: true},
		{Plugin: "json", Unprocess: false},
	})
	if chain == nil {
		t.Fatal("missing URL/Base64/JSON chain")
	}
	if strings.Contains(chain.Preview, "\x1b[") {
		t.Fatalf("JSON preview contains ANSI color escape: %q", chain.Preview)
	}
}

func TestSuggestionsDetectNestedBinaryInspectionChains(t *testing.T) {
	der := []byte{0x30, 0x03, 0x02, 0x01, 0x2a}
	tests := []struct {
		name  string
		input []byte
		want  []SuggestionStep
	}{
		{
			name:  "hex encoded asn1",
			input: []byte(hex.EncodeToString(der)),
			want: []SuggestionStep{
				{Plugin: "hex", Unprocess: true},
				{Plugin: "asn1"},
			},
		},
		{
			name:  "pem wrapped asn1",
			input: pem.EncodeToMemory(&pem.Block{Type: "MESSAGE", Bytes: der}),
			want: []SuggestionStep{
				{Plugin: "pem", Unprocess: true},
				{Plugin: "asn1"},
			},
		},
		{
			name:  "base64 encoded protobuf",
			input: []byte(base64.StdEncoding.EncodeToString([]byte{0x08, 0x96, 0x01, 0x12, 0x03, 'f', 'o', 'o'})),
			want: []SuggestionStep{
				{Plugin: "base64", Unprocess: true},
				{Plugin: "protobuf"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := Suggestions(tt.input)
			if chain := findSuggestionChain(suggestions, tt.want); chain == nil {
				t.Fatalf("missing nested detect chain in %#v", suggestions)
			}
		})
	}
}

func TestStructuredPreview(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   []byte
		want string
	}{
		{"json", []byte(`{"ok":true}`), "JSON"},
		{"xml", []byte(`<root><ok>true</ok></root>`), "XML"},
		{"yaml", []byte("ok: true\nname: deen\n"), "YAML"},
		{"toml", []byte("ok = true\nname = \"deen\"\n"), "TOML"},
		{"csv", []byte("name,ok\ndeen,true\n"), "CSV"},
		{"jwt", []byte("eyJhbGciOiJub25lIn0.eyJzdWIiOiIxMjMifQ."), "JWT"},
		{"uuid", []byte("550e8400-e29b-41d4-a716-446655440000"), "UUID"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := StructuredPreview(tt.in)
			if !ok || !strings.Contains(got, tt.want) {
				t.Fatalf("preview = %q/%v, want %q", got, ok, tt.want)
			}
		})
	}
}

func TestStructuredPreviewSkipsDecoderHints(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{"asn1", []byte{0x30, 0x03, 0x02, 0x01, 0x2a}},
		{"dns", []byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}},
		{"protobuf", []byte{0x08, 0x96, 0x01}},
	}
	msg, err := msgpack.Marshal(map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	tests = append(tests, struct {
		name string
		in   []byte
	}{"msgpack", msg})
	cb, err := cbor.Marshal(map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	tests = append(tests, struct {
		name string
		in   []byte
	}{"cbor", cb})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ok := StructuredPreview(tt.in); ok {
				t.Fatalf("StructuredPreview() = %q/%v, want no decoder hint preview", got, ok)
			}
		})
	}
}

func TestHighlightedPreviewStructuredFormats(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want []SyntaxKind
	}{
		{
			name: "xml",
			in:   []byte(`<root id="1"><ok>true</ok></root>`),
			want: []SyntaxKind{SyntaxKey, SyntaxString, SyntaxPunctuation},
		},
		{
			name: "yaml",
			in:   []byte("ok: true\nn: 12\nname: deen\n"),
			want: []SyntaxKind{SyntaxKey, SyntaxBool, SyntaxNumber, SyntaxString, SyntaxPunctuation},
		},
		{
			name: "toml",
			in:   []byte("ok = true\nn = 12\nname = \"deen\"\n"),
			want: []SyntaxKind{SyntaxKey, SyntaxBool, SyntaxNumber, SyntaxString, SyntaxPunctuation},
		},
		{
			name: "csv",
			in:   []byte("name,ok\ndeen,true\n"),
			want: []SyntaxKind{SyntaxKey, SyntaxPunctuation},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview, spans, ok := HighlightedPreview(tt.in)
			if !ok {
				t.Fatalf("HighlightedPreview(%s) ok = false", tt.name)
			}
			assertSortedNonOverlappingSpans(t, spans)
			for _, want := range tt.want {
				if !hasSyntaxKind(spans, want) {
					t.Fatalf("preview %q missing span kind %s in %#v", preview, want, spans)
				}
			}
		})
	}
}

func assertSortedNonOverlappingSpans(t *testing.T, spans []SyntaxSpan) {
	t.Helper()
	lastEnd := 0
	for _, span := range spans {
		if span.Start < lastEnd {
			t.Fatalf("spans are not sorted and non-overlapping: %#v", spans)
		}
		if span.Start < 0 || span.End <= span.Start {
			t.Fatalf("invalid span range: %#v", span)
		}
		lastEnd = span.End
	}
}

func TestJSONSyntaxSpans(t *testing.T) {
	spans := JSONSyntaxSpans(`{"ok":true,"n":12,"s":"x","nil":null}`)
	for _, want := range []SyntaxKind{SyntaxKey, SyntaxBool, SyntaxNumber, SyntaxString, SyntaxNull, SyntaxPunctuation} {
		if !hasSyntaxKind(spans, want) {
			t.Fatalf("missing span kind %s in %#v", want, spans)
		}
	}
}

func hasSyntaxKind(spans []SyntaxSpan, want SyntaxKind) bool {
	for _, span := range spans {
		if span.Kind == want {
			return true
		}
	}
	return false
}

func TestSuggestionsDetectJWT(t *testing.T) {
	token := "eyJhbGciOiJub25lIn0.eyJzdWIiOiIxMjMifQ."
	if !hasSuggestion(Suggestions([]byte(token)), "jwt", true) {
		t.Fatalf("missing jwt suggestion: %#v", Suggestions([]byte(token)))
	}
}

func TestSuggestionsEmptyInput(t *testing.T) {
	if got := Suggestions([]byte(" \n\t")); len(got) != 0 {
		t.Fatalf("expected no suggestions, got %#v", got)
	}
}

func hasSuggestion(suggestions []Suggestion, plugin string, unprocess bool) bool {
	for _, s := range suggestions {
		if s.Plugin == plugin && s.Unprocess == unprocess {
			return true
		}
	}
	return false
}

func hasSuggestionOption(suggestions []Suggestion, plugin string, unprocess bool, name, value string) bool {
	for _, s := range suggestions {
		if s.Plugin == plugin && s.Unprocess == unprocess && s.Options[name] == value {
			return true
		}
	}
	return false
}

func findSuggestionChain(suggestions []Suggestion, want []SuggestionStep) *Suggestion {
	for i := range suggestions {
		if sameSuggestionChain(suggestions[i].Steps, want) {
			return &suggestions[i]
		}
	}
	return nil
}

func sameSuggestionChain(got, want []SuggestionStep) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Plugin != want[i].Plugin || got[i].Unprocess != want[i].Unprocess {
			return false
		}
	}
	return true
}
