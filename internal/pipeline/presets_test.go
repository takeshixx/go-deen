package pipeline

import "testing"

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
