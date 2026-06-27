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
}

// DataMetadata returns byte-level metadata for data. inputBytes can be zero
// when no ratio should be displayed; otherwise it is used for compression or
// expansion ratio against the previous pipeline value.
func DataMetadata(data []byte, inputBytes int) Metadata {
	m := Metadata{
		Bytes:      len(data),
		UTF8:       utf8.Valid(data),
		Encoding:   likelyTextEncoding(data),
		BOM:        textBOM(data),
		Entropy:    entropy(data),
		InputBytes: inputBytes,
	}
	if len(data) == 0 {
		return m
	}
	m.Lines = bytes.Count(data, []byte{'\n'}) + 1
	m.Printable = printablePercent(data)
	return m
}

// Summary returns a compact human-readable metadata line.
func (m Metadata) Summary() string {
	encoding := m.Encoding
	if encoding == "" {
		encoding = "binary"
	}
	parts := fmt.Sprintf("%d bytes · %d lines · %s", m.Bytes, m.Lines, encoding)
	if m.BOM != "" && m.BOM != "none" {
		parts += " · BOM " + m.BOM
	}
	parts += fmt.Sprintf(" · %d%% printable · %.2f bits/byte entropy", m.Printable, m.Entropy)
	if m.InputBytes > 0 {
		parts += fmt.Sprintf(" · %.2fx input", float64(m.Bytes)/float64(m.InputBytes))
	}
	return parts
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
