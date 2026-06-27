package pipeline

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestChainEncode(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false) // -> dGVzdA==
	p.AddStep("hex", false)    // -> hex of dGVzdA==

	if got := string(p.Output(0)); got != "dGVzdA==" {
		t.Errorf("step 0: got %q", got)
	}
	// hex of "dGVzdA==" is 6447567a64413d3d
	if got := string(p.Output(1)); got != "6447567a64413d3d" {
		t.Errorf("step 1: got %q", got)
	}
}

func TestChainRecomputesOnSourceChange(t *testing.T) {
	p := New()
	p.SetSource([]byte("a"))
	p.AddStep("base64", false)
	first := string(p.Result())
	p.SetSource([]byte("b"))
	if string(p.Result()) == first {
		t.Error("result did not change after source edit")
	}
	if got := string(p.Result()); got != "Yg==" {
		t.Errorf("got %q, want Yg==", got)
	}
}

func TestDecodeDirection(t *testing.T) {
	p := New()
	p.SetSource([]byte("dGVzdA=="))
	p.AddStep("base64", true) // decode
	if got := string(p.Result()); got != "test" {
		t.Errorf("decode: got %q", got)
	}
}

func TestOneWayDecodeError(t *testing.T) {
	p := New()
	p.SetSource([]byte("x"))
	p.AddStep("sha256", true) // hashes can't decode
	if p.Err(0) == nil {
		t.Error("expected an error decoding with a one-way plugin")
	}
}

func TestOptionsApplied(t *testing.T) {
	p := New()
	p.SetSource([]byte("<<<>>>"))
	i := p.AddStep("base64", false)
	std := string(p.Output(i))
	p.SetOption(i, "url", "true")
	url := string(p.Output(i))
	if std == url {
		t.Error("setting -url did not change base64 output")
	}
}

func TestEditOutputRecomputesDownstream(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false) // step 0 -> dGVzdA==
	p.AddStep("base64", true)  // step 1 decodes -> test

	// Override step 0 with a different base64 value; step 1 must follow.
	p.EditOutput(0, []byte("YQ==")) // base64("a")
	if got := string(p.Output(0)); got != "YQ==" {
		t.Errorf("override not applied: %q", got)
	}
	if got := string(p.Output(1)); got != "a" {
		t.Errorf("downstream did not recompute: %q", got)
	}

	// Editing the source clears the override.
	p.SetSource([]byte("test"))
	if got := string(p.Output(0)); got != "dGVzdA==" {
		t.Errorf("override not cleared on source edit: %q", got)
	}
}

func TestRemoveStep(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	p.AddStep("hex", false)
	p.RemoveStep(1)
	if p.Len() != 1 {
		t.Fatalf("expected 1 step, got %d", p.Len())
	}
	if got := string(p.Result()); got != "dGVzdA==" {
		t.Errorf("after remove: got %q", got)
	}
}

func TestMoveStep(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	p.AddStep("hex", false)
	p.MoveStep(1, 0)

	if got := p.Steps()[0].Plugin; got != "hex" {
		t.Fatalf("step 0 = %q, want hex", got)
	}
	if got := string(p.Result()); got != "NzQ2NTczNzQ=" {
		t.Fatalf("unexpected result after move: %q", got)
	}
	if !p.Undo() {
		t.Fatal("expected move to be undoable")
	}
	if got := p.Steps()[0].Plugin; got != "base64" {
		t.Fatalf("undo step 0 = %q, want base64", got)
	}
}

func TestDuplicateStep(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	p.SetOption(0, "raw", "true")
	p.DuplicateStep(0)

	if p.Len() != 2 {
		t.Fatalf("expected 2 steps, got %d", p.Len())
	}
	if p.Steps()[1].Plugin != "base64" {
		t.Fatalf("duplicated plugin = %q", p.Steps()[1].Plugin)
	}
	if got := p.Steps()[1].Options["raw"]; got != "true" {
		t.Fatalf("duplicated option = %q, want true", got)
	}
}

func TestDisableStep(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	p.SetStepDisabled(0, true)
	if got := string(p.Result()); got != "test" {
		t.Fatalf("disabled result = %q, want test", got)
	}
	p.SetStepDisabled(0, false)
	if got := string(p.Result()); got != "dGVzdA==" {
		t.Fatalf("enabled result = %q, want dGVzdA==", got)
	}
}

func TestBinarySafeChain(t *testing.T) {
	src := make([]byte, 256)
	for i := range src {
		src[i] = byte(i)
	}
	p := New()
	p.SetSource(src)
	p.AddStep("gzip", false)
	p.AddStep("gzip", true)
	if !bytes.Equal(p.Result(), src) {
		t.Error("binary round-trip through the pipeline failed")
	}
}

func TestPluginOptions(t *testing.T) {
	opts := PluginOptions("base64")
	if len(opts) == 0 {
		t.Fatal("base64 should report options")
	}
	var foundURL bool
	for _, o := range opts {
		if o.Name == "url" {
			foundURL = true
			if !o.IsBool {
				t.Error("-url should be a bool option")
			}
		}
	}
	if !foundURL {
		t.Error("base64 options missing -url")
	}
	if PluginOptions("sha256") != nil {
		t.Error("sha256 should have no options")
	}
}

func TestUndoRedoSourceAndSteps(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	if got := string(p.Result()); got != "dGVzdA==" {
		t.Fatalf("got %q, want dGVzdA==", got)
	}

	if !p.Undo() {
		t.Fatal("expected undo to remove the step")
	}
	if p.Len() != 0 {
		t.Fatalf("expected no steps after undo, got %d", p.Len())
	}
	if got := string(p.Result()); got != "test" {
		t.Fatalf("undo step result: got %q", got)
	}

	if !p.Undo() {
		t.Fatal("expected undo to clear the source")
	}
	if got := string(p.Result()); got != "" {
		t.Fatalf("undo source result: got %q", got)
	}

	if !p.Redo() {
		t.Fatal("expected redo to restore source")
	}
	if got := string(p.Result()); got != "test" {
		t.Fatalf("redo source result: got %q", got)
	}

	if !p.Redo() {
		t.Fatal("expected redo to restore step")
	}
	if got := string(p.Result()); got != "dGVzdA==" {
		t.Fatalf("redo step result: got %q", got)
	}
}

func TestUndoClearsRedoAfterNewEdit(t *testing.T) {
	p := New()
	p.SetSource([]byte("a"))
	p.SetSource([]byte("b"))
	if !p.Undo() {
		t.Fatal("expected undo")
	}
	p.SetSource([]byte("c"))
	if p.CanRedo() {
		t.Fatal("new edits should clear redo history")
	}
	if got := string(p.Result()); got != "c" {
		t.Fatalf("got %q, want c", got)
	}
}

func TestExportImportJSONPreservesChain(t *testing.T) {
	p := New()
	src := []byte{0, 1, 2, 3, 255}
	p.SetSource(src)
	p.AddStep("gzip", false)
	p.AddStep("base64", false)
	p.SetOption(1, "raw", "true")
	p.EditOutput(1, []byte("manual"))

	data, err := p.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("exported invalid JSON: %s", data)
	}

	loaded := New()
	if err := loaded.ImportJSON(data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(loaded.Source(), src) {
		t.Fatalf("source mismatch: got %v, want %v", loaded.Source(), src)
	}
	if loaded.Len() != 2 {
		t.Fatalf("expected 2 steps, got %d", loaded.Len())
	}
	if got := loaded.Steps()[1].Options["raw"]; got != "true" {
		t.Fatalf("option not restored: got %q", got)
	}
	if got := string(loaded.Result()); got != "manual" {
		t.Fatalf("override not restored: got %q", got)
	}
}

func TestExportImportJSONPreservesDisabledStep(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	p.SetStepDisabled(0, true)
	data, err := p.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}

	loaded := New()
	if err := loaded.ImportJSON(data); err != nil {
		t.Fatal(err)
	}
	if !loaded.Steps()[0].Disabled {
		t.Fatal("disabled state was not restored")
	}
	if got := string(loaded.Result()); got != "test" {
		t.Fatalf("disabled imported result = %q, want test", got)
	}
}

func TestExportJSONWithoutSourceKeepsRecipeOnly(t *testing.T) {
	p := New()
	p.SetSource([]byte("secret"))
	p.AddStep("base64", false)
	p.SetOption(0, "raw", "true")
	p.EditOutput(0, []byte("manual-secret"))

	data, err := p.ExportJSONWithoutSource()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("secret")) || bytes.Contains(data, []byte("manual-secret")) {
		t.Fatalf("recipe export leaked source or override: %s", data)
	}

	loaded := New()
	loaded.SetSource([]byte("input"))
	if err := loaded.ImportJSON(data); err != nil {
		t.Fatal(err)
	}
	if got := string(loaded.Source()); got != "" {
		t.Fatalf("recipe import should not include source, got %q", got)
	}
	if loaded.Len() != 1 {
		t.Fatalf("expected 1 step, got %d", loaded.Len())
	}
	if got := loaded.Steps()[0].Options["raw"]; got != "true" {
		t.Fatalf("option not restored: got %q", got)
	}
}

func TestImportJSONIsUndoable(t *testing.T) {
	p := New()
	p.SetSource([]byte("before"))

	other := New()
	other.SetSource([]byte("after"))
	data, err := other.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}
	if err := p.ImportJSON(data); err != nil {
		t.Fatal(err)
	}
	if got := string(p.Result()); got != "after" {
		t.Fatalf("import result: got %q", got)
	}
	if !p.Undo() {
		t.Fatal("expected import to be undoable")
	}
	if got := string(p.Result()); got != "before" {
		t.Fatalf("undo import result: got %q", got)
	}
}

func TestImportJSONRejectsUnknownPlugin(t *testing.T) {
	data := []byte(`{"version":1,"steps":[{"plugin":"does-not-exist"}]}`)
	if err := New().ImportJSON(data); err == nil {
		t.Fatal("expected unknown plugin error")
	}
}
