package pipeline

import (
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
		return string(data), false
	}
	return string(data[:safeTextCut(data, TextPreviewLimit)]) + truncatedMessage(len(data), TextPreviewLimit), true
}

// HexDisplay returns a UI-safe hex dump and whether it was capped.
func HexDisplay(data []byte) (string, bool) {
	if len(data) <= HexPreviewLimit {
		return hex.Dump(data), false
	}
	return hex.Dump(data[:HexPreviewLimit]) + truncatedMessage(len(data), HexPreviewLimit), true
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
