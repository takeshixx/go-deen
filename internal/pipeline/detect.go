package pipeline

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Suggestion describes a transform that is likely useful for the given data.
type Suggestion struct {
	Plugin    string
	Unprocess bool
	Label     string
	Reason    string
}

// Suggestions returns likely next transforms for a byte slice. It is heuristic:
// suggestions should be helpful shortcuts, not declarations of file type.
func Suggestions(data []byte) []Suggestion {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil
	}

	var out []Suggestion
	add := func(plugin string, unprocess bool, label, reason string) {
		out = append(out, Suggestion{Plugin: plugin, Unprocess: unprocess, Label: label, Reason: reason})
	}

	text := string(trimmed)
	if looksLikeJWT(text) {
		add("jwt", true, "Decode JWT", "input has three base64url JWT sections")
	}
	if looksLikeBase64(text) {
		add("base64", true, "Decode Base64", "input matches a Base64 alphabet and decodes cleanly")
	}
	if looksLikeHex(text) {
		add("hex", true, "Decode hex", "input contains an even-length hexadecimal byte string")
	}
	if strings.Contains(text, "%") && looksLikeURLEncoded(text) {
		add("url", true, "URL decode", "input contains percent-encoded bytes")
	}
	if strings.Contains(text, "&") && strings.Contains(text, ";") {
		add("html", true, "HTML decode", "input contains entity-like ampersand escapes")
	}
	if bytes.Contains(trimmed, []byte("-----BEGIN CERTIFICATE-----")) {
		add("certPrinter", false, "Inspect certificate", "input contains a PEM certificate block")
	}
	if json.Valid(trimmed) {
		add("json", false, "Format JSON", "input parses as JSON")
	}
	if looksLikeXML(trimmed) {
		add("xml", false, "Format XML", "input parses as XML")
	}
	if bytes.HasPrefix(trimmed, []byte{0x1f, 0x8b}) {
		add("gzip", true, "Decompress gzip", "input has a gzip magic header")
	}
	if looksLikeZlib(trimmed) {
		add("zlib", true, "Decompress zlib", "input has a likely zlib header")
	}
	if looksLikeProtobuf(trimmed) {
		add("protobuf", false, "Decode protobuf", "input looks like binary protobuf wire data")
	}
	return out
}

func looksLikeJWT(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || !json.Valid(header) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	return err == nil && json.Valid(payload)
}

func looksLikeBase64(s string) bool {
	compact := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
	if len(compact) < 8 {
		return false
	}
	candidates := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, enc := range candidates {
		decoded, err := enc.DecodeString(compact)
		if err == nil && len(decoded) > 0 && base64Alphabet(compact) {
			return true
		}
	}
	return false
}

func base64Alphabet(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		if r == '+' || r == '/' || r == '-' || r == '_' || r == '=' {
			continue
		}
		return false
	}
	return true
}

func looksLikeHex(s string) bool {
	compact := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
	if len(compact) < 4 || len(compact)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(compact)
	return err == nil
}

func looksLikeURLEncoded(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if s[i] == '%' && isHexByte(s[i+1]) && isHexByte(s[i+2]) {
			return true
		}
	}
	return false
}

func isHexByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func looksLikeXML(data []byte) bool {
	if !bytes.HasPrefix(data, []byte("<")) {
		return false
	}
	var v any
	return xml.Unmarshal(data, &v) == nil
}

func looksLikeZlib(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	cmf, flg := data[0], data[1]
	return cmf&0x0f == 8 && (uint16(cmf)<<8|uint16(flg))%31 == 0
}

func looksLikeProtobuf(data []byte) bool {
	if utf8.Valid(data) && mostlyPrintable(data) {
		return false
	}
	return scanProto(data, 0) == nil
}

func scanProto(data []byte, depth int) error {
	if len(data) == 0 || depth > 3 {
		return errProto
	}
	fields := 0
	for len(data) > 0 {
		key, n := binary.Uvarint(data)
		if n <= 0 {
			return errProto
		}
		data = data[n:]
		field := key >> 3
		wire := key & 0x7
		if field == 0 {
			return errProto
		}
		fields++
		switch wire {
		case 0:
			_, n := binary.Uvarint(data)
			if n <= 0 {
				return errProto
			}
			data = data[n:]
		case 1:
			if len(data) < 8 {
				return errProto
			}
			data = data[8:]
		case 2:
			l, n := binary.Uvarint(data)
			if n <= 0 || uint64(len(data[n:])) < l {
				return errProto
			}
			data = data[n+int(l):]
		case 5:
			if len(data) < 4 {
				return errProto
			}
			data = data[4:]
		default:
			return errProto
		}
	}
	if fields == 0 {
		return errProto
	}
	return nil
}

var errProto = protoScanError{}

type protoScanError struct{}

func (protoScanError) Error() string { return "not protobuf" }

func mostlyPrintable(data []byte) bool {
	var printable, total int
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			return false
		}
		total++
		if r == '\n' || r == '\r' || r == '\t' || (r >= 0x20 && r != 0x7f) {
			printable++
		}
		data = data[size:]
	}
	return total > 0 && printable*100/total >= 85
}
