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
	utf := "binary"
	if m.UTF8 {
		utf = "utf-8"
	}
	parts := fmt.Sprintf("%d bytes · %d lines · %s · %d%% printable · %.2f bits/byte entropy",
		m.Bytes, m.Lines, utf, m.Printable, m.Entropy)
	if m.InputBytes > 0 {
		parts += fmt.Sprintf(" · %.2fx input", float64(m.Bytes)/float64(m.InputBytes))
	}
	return parts
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
