//go:build js && wasm

package webui

import (
	"testing"

	"github.com/takeshixx/deen/internal/pipeline"
)

func TestCompactStepCollapseStateKeepsLatestOpen(t *testing.T) {
	pipe = pipeline.New()
	pipe.SetSource([]byte("test"))
	pipe.AddStep("base64", false)
	pipe.AddStep("hex", false)
	pipe.AddStep("url", false)

	compactStepCollapseState()

	for i, want := range map[int]bool{0: true, 1: true, 2: false} {
		if got := stepCollapsed[i]; got != want {
			t.Fatalf("stepCollapsed[%d] = %v, want %v", i, got, want)
		}
	}
}
