package pipeline

import "unicode"

// SyntaxKind identifies a token style for preview syntax highlighting.
type SyntaxKind string

const (
	SyntaxKey         SyntaxKind = "key"
	SyntaxString      SyntaxKind = "string"
	SyntaxNumber      SyntaxKind = "number"
	SyntaxBool        SyntaxKind = "bool"
	SyntaxNull        SyntaxKind = "null"
	SyntaxPunctuation SyntaxKind = "punctuation"
)

// SyntaxSpan describes a highlighted byte range in preview text.
type SyntaxSpan struct {
	Start int
	End   int
	Kind  SyntaxKind
}

// HighlightedPreview returns a structured preview plus syntax ranges for
// JSON-like fragments embedded in that preview.
func HighlightedPreview(data []byte) (string, []SyntaxSpan, bool) {
	preview, ok := StructuredPreview(data)
	if !ok {
		return "", nil, false
	}
	return preview, JSONSyntaxSpans(preview), true
}

// JSONSyntaxSpans returns simple JSON token spans for syntax highlighting.
func JSONSyntaxSpans(text string) []SyntaxSpan {
	var spans []SyntaxSpan
	for i := 0; i < len(text); {
		c := text[i]
		switch {
		case c == '"':
			end := scanJSONString(text, i)
			kind := SyntaxString
			if isJSONObjectKey(text, end) {
				kind = SyntaxKey
			}
			spans = append(spans, SyntaxSpan{Start: i, End: end, Kind: kind})
			i = end
		case c == '-' || c >= '0' && c <= '9':
			if end, ok := scanJSONNumber(text, i); ok {
				spans = append(spans, SyntaxSpan{Start: i, End: end, Kind: SyntaxNumber})
				i = end
			} else {
				i++
			}
		case hasJSONWord(text, i, "true"), hasJSONWord(text, i, "false"):
			end := i + 4
			if text[i] == 'f' {
				end = i + 5
			}
			spans = append(spans, SyntaxSpan{Start: i, End: end, Kind: SyntaxBool})
			i = end
		case hasJSONWord(text, i, "null"):
			end := i + 4
			spans = append(spans, SyntaxSpan{Start: i, End: end, Kind: SyntaxNull})
			i = end
		case isJSONPunctuation(c):
			spans = append(spans, SyntaxSpan{Start: i, End: i + 1, Kind: SyntaxPunctuation})
			i++
		default:
			i++
		}
	}
	return spans
}

func scanJSONString(text string, start int) int {
	escaped := false
	for i := start + 1; i < len(text); i++ {
		switch {
		case escaped:
			escaped = false
		case text[i] == '\\':
			escaped = true
		case text[i] == '"':
			return i + 1
		}
	}
	return len(text)
}

func isJSONObjectKey(text string, end int) bool {
	for i := end; i < len(text); i++ {
		if isASCIIWhitespace(text[i]) {
			continue
		}
		return text[i] == ':'
	}
	return false
}

func scanJSONNumber(text string, start int) (int, bool) {
	i := start
	if text[i] == '-' {
		i++
		if i >= len(text) {
			return start, false
		}
	}
	if text[i] == '0' {
		i++
	} else if text[i] >= '1' && text[i] <= '9' {
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	} else {
		return start, false
	}
	if i < len(text) && text[i] == '.' {
		i++
		if i >= len(text) || text[i] < '0' || text[i] > '9' {
			return start, false
		}
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	}
	if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
		i++
		if i < len(text) && (text[i] == '+' || text[i] == '-') {
			i++
		}
		if i >= len(text) || text[i] < '0' || text[i] > '9' {
			return start, false
		}
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	}
	return i, true
}

func hasJSONWord(text string, start int, word string) bool {
	if start+len(word) > len(text) || text[start:start+len(word)] != word {
		return false
	}
	beforeOK := start == 0 || !isJSONIdentifierRune(rune(text[start-1]))
	after := start + len(word)
	afterOK := after == len(text) || !isJSONIdentifierRune(rune(text[after]))
	return beforeOK && afterOK
}

func isJSONIdentifierRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isJSONPunctuation(c byte) bool {
	switch c {
	case '{', '}', '[', ']', ':', ',':
		return true
	default:
		return false
	}
}

func isASCIIWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t'
}
