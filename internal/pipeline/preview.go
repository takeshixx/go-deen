package pipeline

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// StructuredPreview returns a readable preview for common structured payloads.
func StructuredPreview(data []byte) (string, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", false
	}
	if json.Valid(trimmed) {
		var v interface{}
		if err := json.Unmarshal(trimmed, &v); err == nil {
			var out bytes.Buffer
			enc := json.NewEncoder(&out)
			enc.SetIndent("", "    ")
			if enc.Encode(v) == nil {
				return "JSON\n\n" + out.String(), true
			}
		}
	}
	if looksLikeJWT(string(trimmed)) {
		parts := strings.Split(string(trimmed), ".")
		header, _ := base64.RawURLEncoding.DecodeString(parts[0])
		payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
		return fmt.Sprintf("JWT\n\nheader:\n%s\n\npayload:\n%s", prettyJSON(header), prettyJSON(payload)), true
	}
	if looksLikeUUID(string(trimmed)) {
		return "UUID\n\n" + strings.TrimSpace(string(trimmed)), true
	}
	if kind := magicType(trimmed); kind != "" {
		return fmt.Sprintf("File type\n\ntype: %s\nmime: %s", kind, http.DetectContentType(trimmed)), true
	}
	if looksLikeDNSName(trimmed) {
		return "DNS wire-format name\n\nUse .dns to decode this name.", true
	}
	if looksLikeASN1(trimmed) {
		return "ASN.1 DER\n\nUse asn1 to inspect the tag/length/value tree.", true
	}
	if looksLikeMessagePack(trimmed) {
		return "MessagePack\n\nUse .msgpack to decode to JSON.", true
	}
	if looksLikeCBOR(trimmed) {
		return "CBOR\n\nUse .cbor to decode to JSON.", true
	}
	if looksLikeProtobuf(trimmed) {
		return "Protocol Buffers\n\nUse protobuf to inspect schema-less fields.", true
	}
	return "", false
}

func prettyJSON(data []byte) string {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return string(data)
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "    ")
	if enc.Encode(v) != nil {
		return string(data)
	}
	return strings.TrimRight(out.String(), "\n")
}
