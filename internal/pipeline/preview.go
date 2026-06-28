package pipeline

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
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
	if looksLikeXML(trimmed) {
		if formatted, ok := prettyXMLPreview(trimmed); ok {
			return "XML\n\n" + formatted, true
		}
	}
	if formatted, ok := prettyTOMLPreview(trimmed); ok {
		return "TOML\n\n" + formatted, true
	}
	if formatted, ok := prettyYAMLPreview(trimmed); ok {
		return "YAML\n\n" + formatted, true
	}
	if formatted, ok := prettyCSVPreview(trimmed); ok {
		return "CSV\n\n" + formatted, true
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
	return "", false
}

func previewBody(preview, label string) (string, int, bool) {
	prefix := label + "\n\n"
	if !strings.HasPrefix(preview, prefix) {
		return "", 0, false
	}
	return preview[len(prefix):], len(prefix), true
}

func prettyXMLPreview(data []byte) (string, bool) {
	var out bytes.Buffer
	dec := xml.NewDecoder(bytes.NewReader(data))
	enc := xml.NewEncoder(&out)
	enc.Indent("", "    ")
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", false
		}
		if cd, ok := tok.(xml.CharData); ok && len(bytes.TrimSpace(cd)) == 0 {
			continue
		}
		if err := enc.EncodeToken(tok); err != nil {
			return "", false
		}
	}
	if err := enc.Flush(); err != nil {
		return "", false
	}
	return strings.TrimRight(out.String(), "\n"), true
}

func prettyTOMLPreview(data []byte) (string, bool) {
	var decoded map[string]interface{}
	if _, err := toml.Decode(string(data), &decoded); err != nil || len(decoded) == 0 {
		return "", false
	}
	var out bytes.Buffer
	if err := toml.NewEncoder(&out).Encode(decoded); err != nil {
		return "", false
	}
	return strings.TrimRight(out.String(), "\n"), true
}

func prettyYAMLPreview(data []byte) (string, bool) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil || len(node.Content) == 0 {
		return "", false
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode && root.Kind != yaml.SequenceNode {
		return "", false
	}
	out, err := yaml.Marshal(root)
	if err != nil {
		return "", false
	}
	return strings.TrimRight(string(out), "\n"), true
}

func prettyCSVPreview(data []byte) (string, bool) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil || len(rows) == 0 {
		return "", false
	}
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	if maxCols < 2 {
		return "", false
	}
	var out strings.Builder
	for _, row := range rows {
		out.WriteString(strings.Join(row, "\t"))
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n"), true
}

// HasStructuredPreview reports whether data can produce a structured preview.
func HasStructuredPreview(data []byte) bool {
	if IsLargeData(data) {
		return false
	}
	_, ok := StructuredPreview(data)
	return ok
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
