package pipeline

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const (
	LargeDataThreshold = 1 << 20
	TextPreviewLimit   = 64 << 10
	HexPreviewLimit    = 16 << 10
)

// IsLargeData reports whether data should be previewed instead of fully
// rendered into editable UI widgets.
func IsLargeData(data []byte) bool {
	return len(data) > LargeDataThreshold
}

// TextDisplay returns a UI-safe text representation and whether it was capped.
func TextDisplay(data []byte) (string, bool) {
	if len(data) <= TextPreviewLimit {
		return TextDisplayFull(data), false
	}
	return string(data[:safeTextCut(data, TextPreviewLimit)]) + truncatedMessage(len(data), TextPreviewLimit), true
}

// TextDisplayFull returns the complete text representation of data.
func TextDisplayFull(data []byte) string {
	return string(data)
}

// HexDisplay returns a UI-safe hex dump and whether it was capped.
func HexDisplay(data []byte) (string, bool) {
	if len(data) <= HexPreviewLimit {
		return HexDisplayFull(data), false
	}
	return hex.Dump(data[:HexPreviewLimit]) + truncatedMessage(len(data), HexPreviewLimit), true
}

// HexDisplayFull returns the complete hex dump of data.
func HexDisplayFull(data []byte) string {
	return hex.Dump(data)
}

// StringsDisplay returns printable ASCII strings found in data, one per line.
// It is intended as a quick binary-data skim, similar to the strings utility.
func StringsDisplay(data []byte) (string, bool) {
	return stringsDisplay(data, true)
}

// StringsDisplayFull returns every printable ASCII string found in data.
func StringsDisplayFull(data []byte) string {
	text, _ := stringsDisplay(data, false)
	return text
}

func stringsDisplay(data []byte, capOutput bool) (string, bool) {
	sample := data
	capped := false
	if capOutput && len(sample) > TextPreviewLimit {
		sample = sample[:TextPreviewLimit]
		capped = true
	}

	var out bytes.Buffer
	start := -1
	for i, b := range sample {
		if isStringByte(b) {
			if start < 0 {
				start = i
			}
			continue
		}
		writeStringRun(&out, sample, start, i)
		start = -1
	}
	writeStringRun(&out, sample, start, len(sample))

	if out.Len() == 0 {
		out.WriteString("(no printable strings found)")
	}
	if capped {
		out.WriteString(truncatedMessage(len(data), len(sample)))
	}
	return out.String(), capped
}

// LargeDataPlaceholder returns text for editable source/output fields whose
// complete content is intentionally not rendered.
func LargeDataPlaceholder(data []byte) string {
	return fmt.Sprintf("Large data: %d bytes. Full content is kept in memory; use previews or download the result.", len(data))
}

func truncatedMessage(total, shown int) string {
	return fmt.Sprintf("\n\n... preview truncated: showing %d of %d bytes ...", shown, total)
}

func safeTextCut(data []byte, limit int) int {
	if limit >= len(data) {
		return len(data)
	}
	cut := limit
	for cut > 0 && data[cut]&0xc0 == 0x80 {
		cut--
	}
	if cut == 0 {
		return limit
	}
	return cut
}

func isStringByte(b byte) bool {
	return b >= 0x20 && b <= 0x7e
}

func writeStringRun(out *bytes.Buffer, data []byte, start, end int) {
	if start < 0 || end-start < 4 {
		return
	}
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
	out.Write(data[start:end])
}
