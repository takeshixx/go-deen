package pipeline

import (
	"bytes"
	"fmt"
	"math"
	"unicode/utf8"
)

// Metadata summarizes a byte slice for UI inspection.
type Metadata struct {
	Bytes      int
	Lines      int
	UTF8       bool
	Encoding   string
	BOM        string
	Printable  int
	Entropy    float64
	InputBytes int
	Sampled    bool
}

// MetadataField is a single labeled metadata value for UI presentation.
type MetadataField struct {
	Label string
	Value string
}

// DataMetadata returns byte-level metadata for data. inputBytes can be zero
// when no ratio should be displayed; otherwise it is used for compression or
// expansion ratio against the previous pipeline value.
func DataMetadata(data []byte, inputBytes int) Metadata {
	sample := data
	sampled := false
	if len(sample) > LargeDataThreshold {
		sample = sample[:LargeDataThreshold]
		sampled = true
	}
	m := Metadata{
		Bytes:      len(data),
		UTF8:       utf8.Valid(sample),
		Encoding:   likelyTextEncoding(sample),
		BOM:        textBOM(sample),
		Entropy:    entropy(sample),
		InputBytes: inputBytes,
		Sampled:    sampled,
	}
	if len(data) == 0 {
		return m
	}
	m.Lines = bytes.Count(sample, []byte{'\n'}) + 1
	m.Printable = printablePercent(sample)
	return m
}

// Summary returns a compact human-readable metadata line.
func (m Metadata) Summary() string {
	fields := m.Fields()
	if len(fields) == 0 {
		return ""
	}
	parts := fields[0].Value
	for _, field := range fields[1:] {
		if field.Label == "sample" {
			continue
		}
		if field.Label == "lines" {
			parts += " · " + field.Value + " lines"
			continue
		}
		if field.Label == "BOM" {
			parts += " · BOM " + field.Value
			continue
		}
		if field.Label == "printable" {
			parts += " · " + field.Value + " printable"
			continue
		}
		if field.Label == "entropy" {
			parts += " · " + field.Value + " entropy"
			continue
		}
		if field.Label == "ratio" {
			parts += " · " + field.Value + " input"
			continue
		}
		parts += " · " + field.Value
	}
	if m.Sampled {
		parts += " · metadata sampled"
	}
	return parts
}

// Fields returns metadata as labeled UI-ready values.
func (m Metadata) Fields() []MetadataField {
	encoding := m.Encoding
	if encoding == "" {
		encoding = "binary"
	}
	fields := []MetadataField{
		{"size", fmt.Sprintf("%d bytes%s", m.Bytes, sizeUnits(m.Bytes))},
		{"lines", fmt.Sprintf("%d", m.Lines)},
		{"encoding", encoding},
	}
	if m.BOM != "" && m.BOM != "none" {
		fields = append(fields, MetadataField{"BOM", m.BOM})
	}
	fields = append(fields,
		MetadataField{"printable", fmt.Sprintf("%d%%", m.Printable)},
		MetadataField{"entropy", fmt.Sprintf("%.2f bits/byte", m.Entropy)},
	)
	if m.InputBytes > 0 {
		fields = append(fields, MetadataField{"ratio", fmt.Sprintf("%.2fx", float64(m.Bytes)/float64(m.InputBytes))})
	}
	if m.Sampled {
		fields = append(fields, MetadataField{"sample", "sampled"})
	}
	return fields
}

func sizeUnits(bytes int) string {
	if bytes <= LargeDataThreshold {
		return ""
	}
	if bytes >= 1<<30 {
		return fmt.Sprintf(" (%.2f GB)", float64(bytes)/(1<<30))
	}
	return fmt.Sprintf(" (%.2f MB)", float64(bytes)/(1<<20))
}

func textBOM(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}):
		return "UTF-8"
	case bytes.HasPrefix(data, []byte{0xff, 0xfe, 0x00, 0x00}):
		return "UTF-32LE"
	case bytes.HasPrefix(data, []byte{0x00, 0x00, 0xfe, 0xff}):
		return "UTF-32BE"
	case bytes.HasPrefix(data, []byte{0xff, 0xfe}):
		return "UTF-16LE"
	case bytes.HasPrefix(data, []byte{0xfe, 0xff}):
		return "UTF-16BE"
	default:
		return "none"
	}
}

func likelyTextEncoding(data []byte) string {
	if len(data) == 0 {
		return "empty"
	}
	if bom := textBOM(data); bom != "none" {
		return bom
	}
	nulls := nullPattern(data)
	var evenNulls, oddNulls int
	for i, count := range nulls {
		if i%2 == 0 {
			evenNulls += count
		} else {
			oddNulls += count
		}
	}
	switch {
	case len(data) >= 8 && nulls[1] > 0 && nulls[2] > 0 && nulls[3] > 0 && nulls[0] == 0:
		return "likely UTF-32LE"
	case len(data) >= 8 && nulls[0] > 0 && nulls[1] > 0 && nulls[2] > 0 && nulls[3] == 0:
		return "likely UTF-32BE"
	case len(data) >= 4 && oddNulls >= 2 && evenNulls == 0:
		return "likely UTF-16LE"
	case len(data) >= 4 && evenNulls >= 2 && oddNulls == 0:
		return "likely UTF-16BE"
	}
	if utf8.Valid(data) {
		if asciiOnly(data) {
			return "ASCII / UTF-8"
		}
		return "UTF-8"
	}
	return "binary"
}

func asciiOnly(data []byte) bool {
	for _, b := range data {
		if b >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func nullPattern(data []byte) [4]int {
	var counts [4]int
	for i, b := range data {
		if b == 0 {
			counts[i%4]++
		}
	}
	return counts
}

func entropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	var counts [256]int
	for _, b := range data {
		counts[b]++
	}
	var e float64
	for _, count := range counts {
		if count == 0 {
			continue
		}
		p := float64(count) / float64(len(data))
		e -= p * math.Log2(p)
	}
	return e
}

func printablePercent(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	var printable int
	for _, b := range data {
		if b == '\n' || b == '\r' || b == '\t' || (b >= 0x20 && b < 0x7f) {
			printable++
		}
	}
	return printable * 100 / len(data)
}
