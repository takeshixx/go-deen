package pipeline

import "testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !hasSuggestion(Suggestions(tt.input), tt.plugin, tt.unprocess) {
				t.Fatalf("missing suggestion %s/%v in %#v", tt.plugin, tt.unprocess, Suggestions(tt.input))
			}
		})
	}
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
