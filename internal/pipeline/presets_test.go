package pipeline

import (
	"strings"
	"testing"
)

func TestBuiltinPresetsAreValid(t *testing.T) {
	presets := BuiltinPresets()
	if len(presets) == 0 {
		t.Fatal("expected built-in presets")
	}
	for _, preset := range presets {
		if preset.Name == "" {
			t.Fatal("preset with empty name")
		}
		if preset.Description == "" {
			t.Fatalf("%s has empty description", preset.Name)
		}
		if len(preset.Steps) == 0 {
			t.Fatalf("%s has no steps", preset.Name)
		}
		for _, step := range preset.Steps {
			if step.Plugin == "" {
				t.Fatalf("%s has empty plugin step", preset.Name)
			}
		}
	}
}

func TestApplyPresetPreservesSourceAndIsUndoable(t *testing.T) {
	p := New()
	p.SetSource([]byte("dGVzdA=="))
	p.AddStep("hex", false)

	p.ApplyPreset(Preset{
		Name:        "test",
		Description: "test preset",
		Steps: []PresetStep{
			{Plugin: "base64", Unprocess: true},
		},
	})
	if got := string(p.Source()); got != "dGVzdA==" {
		t.Fatalf("source = %q", got)
	}
	if p.Len() != 1 || p.Steps()[0].Plugin != "base64" || !p.Steps()[0].Unprocess {
		t.Fatalf("unexpected preset steps: %#v", p.Steps())
	}
	if got := string(p.Result()); got != "test" {
		t.Fatalf("result = %q, want test", got)
	}
	if !p.Undo() {
		t.Fatal("expected preset application to be undoable")
	}
	if p.Len() != 1 || p.Steps()[0].Plugin != "hex" {
		t.Fatalf("undo did not restore previous chain: %#v", p.Steps())
	}
}

func TestBuiltinPresetNames(t *testing.T) {
	names := map[string]bool{}
	for _, preset := range BuiltinPresets() {
		names[preset.Name] = true
	}
	for _, want := range []string{"Decode JWT", "Decode SAML Redirect", "JWK Thumbprint", "Timestamp ms"} {
		if !names[want] {
			t.Fatalf("missing preset %q", want)
		}
	}
}

func TestBuiltinExamplesAreRunnable(t *testing.T) {
	examples := BuiltinExamples()
	if len(examples) == 0 {
		t.Fatal("expected built-in examples")
	}
	for _, example := range examples {
		t.Run(example.Name, func(t *testing.T) {
			if example.Description == "" {
				t.Fatal("missing description")
			}
			if len(example.Steps) == 0 {
				t.Fatal("missing steps")
			}
			p := New()
			p.ApplyExample(example)
			if len(p.Source()) == 0 {
				t.Fatal("example did not set source")
			}
			for i, step := range p.Steps() {
				if step.err != nil {
					t.Fatalf("step %d failed: %v", i+1, step.err)
				}
			}
			if example.WantContains != "" && !strings.Contains(string(p.Result()), example.WantContains) {
				t.Fatalf("result %q does not contain %q", string(p.Result()), example.WantContains)
			}
		})
	}
}

func TestApplyExampleIsUndoable(t *testing.T) {
	p := New()
	p.SetSource([]byte("before"))
	p.AddStep("hex", false)
	p.ApplyExample(Example{
		Name:        "test",
		Description: "test example",
		Source:      []byte("dGVzdA=="),
		Steps:       []PresetStep{{Plugin: "base64", Unprocess: true}},
	})
	if got := string(p.Source()); got != "dGVzdA==" {
		t.Fatalf("source = %q", got)
	}
	if got := string(p.Result()); got != "test" {
		t.Fatalf("result = %q, want test", got)
	}
	if !p.Undo() {
		t.Fatal("expected example application to be undoable")
	}
	if got := string(p.Source()); got != "before" {
		t.Fatalf("undo source = %q", got)
	}
	if p.Len() != 1 || p.Steps()[0].Plugin != "hex" {
		t.Fatalf("undo did not restore previous chain: %#v", p.Steps())
	}
}
