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
	if body, offset, ok := previewBody(preview, "XML"); ok {
		return preview, offsetSpans(XMLSyntaxSpans(body), offset), true
	}
	if body, offset, ok := previewBody(preview, "YAML"); ok {
		return preview, offsetSpans(LineSyntaxSpans(body, true), offset), true
	}
	if body, offset, ok := previewBody(preview, "TOML"); ok {
		return preview, offsetSpans(LineSyntaxSpans(body, true), offset), true
	}
	if body, offset, ok := previewBody(preview, "CSV"); ok {
		return preview, offsetSpans(CSVSyntaxSpans(body), offset), true
	}
	return preview, JSONSyntaxSpans(preview), true
}

func offsetSpans(spans []SyntaxSpan, offset int) []SyntaxSpan {
	for i := range spans {
		spans[i].Start += offset
		spans[i].End += offset
	}
	return spans
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

// XMLSyntaxSpans returns simple XML token spans for preview highlighting.
func XMLSyntaxSpans(text string) []SyntaxSpan {
	var spans []SyntaxSpan
	for i := 0; i < len(text); {
		if text[i] != '<' {
			i++
			continue
		}
		end := stringsIndexByte(text, '>', i+1)
		if end < 0 {
			break
		}
		spans = append(spans, SyntaxSpan{Start: i, End: i + 1, Kind: SyntaxPunctuation})
		spans = append(spans, SyntaxSpan{Start: end, End: end + 1, Kind: SyntaxPunctuation})
		j := i + 1
		if j < end && text[j] == '/' {
			spans = append(spans, SyntaxSpan{Start: j, End: j + 1, Kind: SyntaxPunctuation})
			j++
		}
		for j < end && isASCIIWhitespace(text[j]) {
			j++
		}
		nameStart := j
		for j < end && isXMLNameByte(text[j]) {
			j++
		}
		if j > nameStart {
			spans = append(spans, SyntaxSpan{Start: nameStart, End: j, Kind: SyntaxKey})
		}
		for j < end {
			switch {
			case isASCIIWhitespace(text[j]) || text[j] == '/':
				j++
			case isXMLNameByte(text[j]):
				attrStart := j
				for j < end && isXMLNameByte(text[j]) {
					j++
				}
				spans = append(spans, SyntaxSpan{Start: attrStart, End: j, Kind: SyntaxKey})
			case text[j] == '=':
				spans = append(spans, SyntaxSpan{Start: j, End: j + 1, Kind: SyntaxPunctuation})
				j++
			case text[j] == '"' || text[j] == '\'':
				quote := text[j]
				valueStart := j
				j++
				for j < end && text[j] != quote {
					j++
				}
				if j < end {
					j++
				}
				spans = append(spans, SyntaxSpan{Start: valueStart, End: j, Kind: SyntaxString})
			default:
				j++
			}
		}
		i = end + 1
	}
	return spans
}

// LineSyntaxSpans highlights key/value line formats such as YAML and TOML.
func LineSyntaxSpans(text string, sectionHeaders bool) []SyntaxSpan {
	var spans []SyntaxSpan
	lineStart := 0
	for lineStart <= len(text) {
		lineEnd := lineStart
		for lineEnd < len(text) && text[lineEnd] != '\n' {
			lineEnd++
		}
		spans = append(spans, lineValueSpans(text, lineStart, lineEnd, sectionHeaders)...)
		if lineEnd == len(text) {
			break
		}
		lineStart = lineEnd + 1
	}
	return spans
}

func lineValueSpans(text string, start, end int, sectionHeaders bool) []SyntaxSpan {
	var spans []SyntaxSpan
	i := start
	for i < end && isASCIIWhitespace(text[i]) {
		i++
	}
	if i >= end || text[i] == '#' {
		return spans
	}
	if sectionHeaders && text[i] == '[' {
		spans = append(spans, SyntaxSpan{Start: i, End: end, Kind: SyntaxKey})
		return spans
	}
	sep := -1
	for j := i; j < end; j++ {
		if text[j] == ':' || text[j] == '=' {
			sep = j
			break
		}
	}
	if sep < 0 {
		return spans
	}
	keyEnd := sep
	for keyEnd > i && isASCIIWhitespace(text[keyEnd-1]) {
		keyEnd--
	}
	if keyEnd > i {
		spans = append(spans, SyntaxSpan{Start: i, End: keyEnd, Kind: SyntaxKey})
	}
	spans = append(spans, SyntaxSpan{Start: sep, End: sep + 1, Kind: SyntaxPunctuation})
	valueStart := sep + 1
	for valueStart < end && isASCIIWhitespace(text[valueStart]) {
		valueStart++
	}
	if valueStart < end {
		spans = append(spans, valueSpan(text, valueStart, end))
	}
	return spans
}

func valueSpan(text string, start, end int) SyntaxSpan {
	kind := SyntaxString
	if text[start] == '"' || text[start] == '\'' {
		return SyntaxSpan{Start: start, End: end, Kind: SyntaxString}
	}
	wordEnd := start
	for wordEnd < end && !isASCIIWhitespace(text[wordEnd]) && text[wordEnd] != '#' {
		wordEnd++
	}
	word := text[start:wordEnd]
	switch {
	case word == "true" || word == "false":
		kind = SyntaxBool
	case word == "null" || word == "~":
		kind = SyntaxNull
	case isNumberWord(word):
		kind = SyntaxNumber
	}
	return SyntaxSpan{Start: start, End: wordEnd, Kind: kind}
}

// CSVSyntaxSpans highlights the first row as keys and separators as punctuation.
func CSVSyntaxSpans(text string) []SyntaxSpan {
	var spans []SyntaxSpan
	firstLineEnd := stringsIndexByte(text, '\n', 0)
	if firstLineEnd < 0 {
		firstLineEnd = len(text)
	}
	fieldStart := 0
	for i := 0; i <= firstLineEnd; i++ {
		if i == firstLineEnd || text[i] == '\t' {
			if i > fieldStart {
				spans = append(spans, SyntaxSpan{Start: fieldStart, End: i, Kind: SyntaxKey})
			}
			if i < firstLineEnd {
				spans = append(spans, SyntaxSpan{Start: i, End: i + 1, Kind: SyntaxPunctuation})
			}
			fieldStart = i + 1
		}
	}
	for i := firstLineEnd + 1; i < len(text); i++ {
		if text[i] == '\t' {
			spans = append(spans, SyntaxSpan{Start: i, End: i + 1, Kind: SyntaxPunctuation})
		}
	}
	return spans
}

func stringsIndexByte(text string, b byte, start int) int {
	for i := start; i < len(text); i++ {
		if text[i] == b {
			return i
		}
	}
	return -1
}

func isXMLNameByte(c byte) bool {
	return c == ':' || c == '_' || c == '-' || c == '.' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

func isNumberWord(word string) bool {
	if word == "" {
		return false
	}
	end, ok := scanJSONNumber(word, 0)
	return ok && end == len(word)
}
