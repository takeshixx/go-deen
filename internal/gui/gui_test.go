//go:build gui

package gui

import "testing"

func TestGUISafeSuggestionPreviewEscapesControls(t *testing.T) {
	got := guiSafeSuggestionPreview("ok\x00\n\x1b[31m")
	want := `ok\x00` + "\n" + `\x1b[31m`
	if got != want {
		t.Fatalf("guiSafeSuggestionPreview() = %q, want %q", got, want)
	}
}
