package pipeline

import (
	"bytes"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/fxamacker/cbor/v2"
	"github.com/liyue201/goqr"
	"github.com/vmihailenco/msgpack/v5"
)

// Suggestion describes a transform that is likely useful for the given data.
type Suggestion struct {
	Plugin     string
	Unprocess  bool
	Options    map[string]string
	Label      string
	Reason     string
	Steps      []SuggestionStep
	Confidence int
	Preview    string
}

// SuggestionStep is one step in an automated suggestion chain.
type SuggestionStep struct {
	Plugin    string
	Unprocess bool
	Options   map[string]string
}

// Suggestions returns likely next transforms for a byte slice. It is heuristic:
// suggestions should be helpful shortcuts, not declarations of file type.
func Suggestions(data []byte) []Suggestion {
	suggestions := oneStepSuggestions(data)
	return append(suggestions, automatedSuggestions(data)...)
}

func oneStepSuggestions(data []byte) []Suggestion {
	if len(data) > LargeDataThreshold {
		data = data[:LargeDataThreshold]
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil
	}

	var out []Suggestion
	add := func(plugin string, unprocess bool, label, reason string) {
		out = append(out, suggestionForStep(SuggestionStep{Plugin: plugin, Unprocess: unprocess}, label, reason, 75))
	}
	addOptions := func(plugin string, unprocess bool, options map[string]string, label, reason string) {
		out = append(out, suggestionForStep(SuggestionStep{Plugin: plugin, Unprocess: unprocess, Options: options}, label, reason, 75))
	}

	meta := DataMetadata(trimmed, 0)
	switch meta.Encoding {
	case "UTF-16LE", "likely UTF-16LE":
		addOptions("unicode", true, map[string]string{"encoding": "utf16le"}, "Decode UTF-16LE to UTF-8", "input has UTF-16 little-endian byte patterns")
	case "UTF-16BE", "likely UTF-16BE":
		addOptions("unicode", true, map[string]string{"encoding": "utf16be"}, "Decode UTF-16BE to UTF-8", "input has UTF-16 big-endian byte patterns")
	case "UTF-32LE", "likely UTF-32LE":
		addOptions("unicode", true, map[string]string{"encoding": "utf32le"}, "Decode UTF-32LE to UTF-8", "input has UTF-32 little-endian byte patterns")
	case "UTF-32BE", "likely UTF-32BE":
		addOptions("unicode", true, map[string]string{"encoding": "utf32be"}, "Decode UTF-32BE to UTF-8", "input has UTF-32 big-endian byte patterns")
	}
	if meta.Encoding != "ASCII / UTF-8" && meta.Encoding != "UTF-8" && meta.Encoding != "empty" {
		add("unicode-inspect", false, "Inspect text encoding", "input has text-encoding clues worth inspecting")
	}

	text := string(trimmed)
	if looksLikeJWT(text) {
		add("jwt", true, "Decode JWT", "input has three base64url JWT sections")
	}
	if looksLikeUUID(text) {
		add("uuid", false, "Inspect UUID", "input is a UUID")
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
	if bytes.Contains(trimmed, []byte("-----BEGIN ")) {
		add("pem", true, "Unwrap PEM", "input contains a PEM block")
	}
	if bytes.Contains(trimmed, []byte("-----BEGIN CERTIFICATE-----")) {
		add("certPrinter", false, "Inspect certificate", "input contains a PEM certificate block")
	}
	if json.Valid(trimmed) {
		addOptions("json", false, map[string]string{"no-color": "true"}, "Format JSON", "input parses as JSON")
	}
	if looksLikeXML(trimmed) {
		add("xml", false, "Format XML", "input parses as XML")
	}
	if looksLikeASN1(trimmed) {
		add("asn1", false, "Inspect ASN.1 DER", "input parses as ASN.1 DER")
	}
	if looksLikeDNSName(trimmed) {
		add("dns", true, "Decode DNS name", "input looks like a DNS wire-format name")
	}
	if looksLikeMessagePack(trimmed) {
		add("msgpack", true, "Decode MessagePack", "input parses as MessagePack")
	}
	if looksLikeCBOR(trimmed) {
		add("cbor", true, "Decode CBOR", "input parses as CBOR")
	}
	if magicType(trimmed) != "" {
		add("magic", false, "Detect file type", "input has a recognizable file signature")
	}
	if looksLikeExecutable(trimmed) {
		add("bininspect", false, "Inspect binary structure", "input has an executable file signature")
	}
	if looksLikeQRImage(trimmed) {
		add("qr", true, "Decode QR image", "input image contains a QR code")
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

func suggestionForStep(step SuggestionStep, label, reason string, confidence int) Suggestion {
	return Suggestion{
		Plugin:     step.Plugin,
		Unprocess:  step.Unprocess,
		Options:    cloneSuggestionOptions(step.Options),
		Label:      label,
		Reason:     reason,
		Steps:      []SuggestionStep{cloneSuggestionStep(step)},
		Confidence: confidence,
	}
}

func automatedSuggestions(data []byte) []Suggestion {
	if len(data) > LargeDataThreshold {
		data = data[:LargeDataThreshold]
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil
	}

	type state struct {
		data  []byte
		steps []SuggestionStep
	}
	queue := []state{{data: data}}
	seen := map[string]bool{stateKey(data): true}
	var out []Suggestion

	for depth := 0; depth < 4 && len(queue) > 0; depth++ {
		current := queue
		queue = nil
		for _, st := range current {
			for _, s := range oneStepSuggestions(st.data) {
				step := SuggestionStep{Plugin: s.Plugin, Unprocess: s.Unprocess, Options: s.Options}
				if canFinishAutomatedChain(s) && len(st.steps) > 0 {
					steps := append(cloneSuggestionSteps(st.steps), cloneSuggestionStep(step))
					out = append(out, chainSuggestion(steps, s, applyPreview(st.data, step)))
				}
				if !canExpandAutomatedChain(s) || len(st.steps) >= 3 {
					continue
				}
				next, ok := applySuggestionStep(st.data, step)
				if !ok || bytes.Equal(bytes.TrimSpace(next), bytes.TrimSpace(st.data)) {
					continue
				}
				key := stateKey(next)
				if seen[key] {
					continue
				}
				seen[key] = true
				queue = append(queue, state{
					data:  next,
					steps: append(cloneSuggestionSteps(st.steps), cloneSuggestionStep(step)),
				})
			}
		}
	}
	return dedupeChainSuggestions(out, 4)
}

func chainSuggestion(steps []SuggestionStep, terminal Suggestion, preview string) Suggestion {
	label := "Apply chain: " + chainLabel(steps)
	reason := fmt.Sprintf("%s after %d automatic decode step(s)", terminal.Reason, len(steps)-1)
	confidence := 70 + len(steps)*5
	if confidence > 95 {
		confidence = 95
	}
	return Suggestion{
		Plugin:     steps[0].Plugin,
		Unprocess:  steps[0].Unprocess,
		Options:    cloneSuggestionOptions(steps[0].Options),
		Label:      label,
		Reason:     reason,
		Steps:      steps,
		Confidence: confidence,
		Preview:    preview,
	}
}

func chainLabel(steps []SuggestionStep) string {
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		parts = append(parts, stepLabel(step))
	}
	return strings.Join(parts, " -> ")
}

func stepLabel(step SuggestionStep) string {
	switch step.Plugin {
	case "base64":
		return "Base64 decode"
	case "hex":
		return "hex decode"
	case "url":
		return "URL decode"
	case "html":
		return "HTML decode"
	case "gzip":
		return "gzip decompress"
	case "zlib":
		return "zlib decompress"
	case "unicode":
		if step.Options["encoding"] != "" {
			return "decode " + strings.ToUpper(step.Options["encoding"])
		}
		return "Unicode decode"
	case "pem":
		return "PEM unwrap"
	case "json":
		return "format JSON"
	case "xml":
		return "format XML"
	case "asn1":
		return "inspect ASN.1"
	case "protobuf":
		return "decode protobuf"
	case "msgpack":
		return "decode MessagePack"
	case "cbor":
		return "decode CBOR"
	case "dns":
		return "decode DNS"
	case "magic":
		return "detect file type"
	case "bininspect":
		return "inspect binary"
	case "certPrinter":
		return "inspect certificate"
	default:
		if step.Unprocess {
			return step.Plugin + " decode"
		}
		return step.Plugin
	}
}

func canExpandAutomatedChain(s Suggestion) bool {
	switch s.Plugin {
	case "base64", "hex", "url", "html", "gzip", "zlib", "unicode", "pem":
		return s.Unprocess
	default:
		return false
	}
}

func canFinishAutomatedChain(s Suggestion) bool {
	switch s.Plugin {
	case "json", "xml", "asn1", "protobuf", "msgpack", "cbor", "dns", "magic", "bininspect", "certPrinter", "uuid", "jwt":
		return true
	default:
		return false
	}
}

func applySuggestionStep(data []byte, step SuggestionStep) ([]byte, bool) {
	out, err := runStep(&Step{
		Plugin:    step.Plugin,
		Unprocess: step.Unprocess,
		Options:   cloneSuggestionOptions(step.Options),
	}, data)
	if err != nil || len(out) > LargeDataThreshold {
		return nil, false
	}
	return bytes.TrimSpace(out), true
}

func applyPreview(data []byte, step SuggestionStep) string {
	out, ok := applySuggestionStep(data, step)
	if !ok {
		return ""
	}
	if utf8.Valid(out) && mostlyPrintable(out) {
		if len(out) > 240 {
			return string(out[:safeTextCut(out, 240)]) + "..."
		}
		return string(out)
	}
	if len(out) > 64 {
		out = out[:64]
	}
	return hex.Dump(out)
}

func stateKey(data []byte) string {
	if len(data) > 256 {
		data = data[:256]
	}
	return fmt.Sprintf("%d:%x", len(data), data)
}

func dedupeChainSuggestions(in []Suggestion, limit int) []Suggestion {
	var out []Suggestion
	seen := map[string]bool{}
	for _, s := range in {
		key := chainLabel(s.Steps)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
		if len(out) >= limit {
			return out
		}
	}
	return out
}

func cloneSuggestionStep(step SuggestionStep) SuggestionStep {
	return SuggestionStep{
		Plugin:    step.Plugin,
		Unprocess: step.Unprocess,
		Options:   cloneSuggestionOptions(step.Options),
	}
}

func cloneSuggestionSteps(steps []SuggestionStep) []SuggestionStep {
	out := make([]SuggestionStep, 0, len(steps))
	for _, step := range steps {
		out = append(out, cloneSuggestionStep(step))
	}
	return out
}

func cloneSuggestionOptions(options map[string]string) map[string]string {
	if len(options) == 0 {
		return nil
	}
	out := make(map[string]string, len(options))
	for k, v := range options {
		out[k] = v
	}
	return out
}

var uuidRE = regexp.MustCompile(`(?i)^(urn:uuid:)?[0-9a-f]{8}-?[0-9a-f]{4}-?[1-8][0-9a-f]{3}-?[89ab][0-9a-f]{3}-?[0-9a-f]{12}$`)

func looksLikeUUID(s string) bool {
	return uuidRE.MatchString(strings.TrimSpace(s))
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

func looksLikeASN1(data []byte) bool {
	var raw asn1.RawValue
	rest, err := asn1.Unmarshal(data, &raw)
	return err == nil && len(rest) == 0 && len(data) > 2
}

func looksLikeDNSName(data []byte) bool {
	if len(data) < 2 || len(data) > 255 {
		return false
	}
	pos := 0
	labels := 0
	for {
		if pos >= len(data) {
			return false
		}
		l := int(data[pos])
		pos++
		if l == 0 {
			return pos == len(data) && labels > 0
		}
		if l > 63 || pos+l > len(data) {
			return false
		}
		for _, b := range data[pos : pos+l] {
			if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-') {
				return false
			}
		}
		pos += l
		labels++
	}
}

func looksLikeMessagePack(data []byte) bool {
	if len(data) < 2 || utf8.Valid(data) && mostlyPrintable(data) {
		return false
	}
	if !plausibleMessagePackContainer(data) {
		return false
	}
	var v interface{}
	if err := msgpack.Unmarshal(data, &v); err != nil {
		return false
	}
	return structuredBinaryValue(v)
}

func plausibleMessagePackContainer(data []byte) bool {
	first := data[0]
	switch {
	case first >= 0x80 && first <= 0x8f:
		return int(first&0x0f) <= len(data)-1
	case first >= 0x90 && first <= 0x9f:
		return int(first&0x0f) <= len(data)-1
	case first == 0xdc || first == 0xde:
		if len(data) < 3 {
			return false
		}
		return int(binary.BigEndian.Uint16(data[1:3])) <= len(data)-3
	case first == 0xdd || first == 0xdf:
		if len(data) < 5 {
			return false
		}
		size := binary.BigEndian.Uint32(data[1:5])
		return size <= uint32(len(data)-5)
	default:
		return false
	}
}

func looksLikeCBOR(data []byte) bool {
	if len(data) < 2 || utf8.Valid(data) && mostlyPrintable(data) {
		return false
	}
	var v interface{}
	return cbor.Unmarshal(data, &v) == nil && structuredBinaryValue(v)
}

func structuredBinaryValue(v interface{}) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		return rv.Len() > 0
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		return rv.Len() > 0
	default:
		return false
	}
}

func magicType(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}):
		return "PNG image"
	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		return "JPEG image"
	case bytes.HasPrefix(data, []byte("GIF8")):
		return "GIF image"
	case bytes.HasPrefix(data, []byte("%PDF-")):
		return "PDF document"
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return "ZIP archive"
	case bytes.HasPrefix(data, []byte{0x1f, 0x8b}):
		return "gzip data"
	case bytes.HasPrefix(data, []byte{0x7f, 'E', 'L', 'F'}):
		return "ELF executable"
	case bytes.HasPrefix(data, []byte("MZ")):
		return "PE executable"
	case looksLikeMachO(data):
		return "Mach-O executable"
	default:
		mime := http.DetectContentType(data)
		if strings.HasPrefix(mime, "image/") || strings.HasPrefix(mime, "application/pdf") {
			return mime
		}
		return ""
	}
}

func looksLikeExecutable(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x7f, 'E', 'L', 'F'}) ||
		bytes.HasPrefix(data, []byte("MZ")) ||
		looksLikeMachO(data)
}

func looksLikeMachO(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	switch binary.BigEndian.Uint32(data[:4]) {
	case 0xfeedface, 0xcefaedfe, 0xfeedfacf, 0xcffaedfe, 0xcafebabe, 0xbebafeca:
		return true
	default:
		return false
	}
}

func looksLikeQRImage(data []byte) bool {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}
	codes, err := goqr.Recognize(img)
	return err == nil && len(codes) > 0
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
