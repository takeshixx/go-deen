package pipeline

// Preset is a named transform-chain recipe.
type Preset struct {
	Name        string
	Description string
	Steps       []PresetStep
}

// PresetStep describes a step in a built-in preset.
type PresetStep struct {
	Plugin    string
	Unprocess bool
	Options   map[string]string
}

// BuiltinPresets returns starter recipes for common inspection workflows.
func BuiltinPresets() []Preset {
	return []Preset{
		{
			Name:        "Decode JWT",
			Description: "Decode a JSON Web Token into readable header and claim JSON.",
			Steps: []PresetStep{
				{Plugin: "jwt", Unprocess: true},
			},
		},
		{
			Name:        "Decode SAML Redirect",
			Description: "Decode a URL-escaped, DEFLATE-compressed SAMLRequest or SAMLResponse into XML.",
			Steps: []PresetStep{
				{Plugin: "saml", Options: map[string]string{"url": "true", "deflate": "true"}},
			},
		},
		{
			Name:        "URL Base64 JSON",
			Description: "URL-decode, Base64-decode and format JSON copied from a parameter or token.",
			Steps: []PresetStep{
				{Plugin: "url", Unprocess: true},
				{Plugin: "base64", Unprocess: true},
				{Plugin: "json"},
			},
		},
		{
			Name:        "Base64 Gzip",
			Description: "Base64-decode and gzip-decompress a compact payload.",
			Steps: []PresetStep{
				{Plugin: "base64", Unprocess: true},
				{Plugin: "gzip", Unprocess: true},
			},
		},
		{
			Name:        "Hex Protobuf",
			Description: "Hex-decode and inspect schema-less protobuf wire data.",
			Steps: []PresetStep{
				{Plugin: "hex", Unprocess: true},
				{Plugin: "protobuf"},
			},
		},
		{
			Name:        "JWK Thumbprint",
			Description: "Normalize a JWK or JWKS and compute SHA-256 thumbprints.",
			Steps: []PresetStep{
				{Plugin: "jwk", Options: map[string]string{"thumbprint": "true", "public": "true"}},
			},
		},
		{
			Name:        "Timestamp ms",
			Description: "Convert Unix milliseconds into RFC3339 time.",
			Steps: []PresetStep{
				{Plugin: "timestamp", Options: map[string]string{"unit": "ms"}},
			},
		},
	}
}

// ApplyPreset replaces the current chain with a preset while preserving the
// current source input. The previous pipeline state remains undoable.
func (p *Pipeline) ApplyPreset(preset Preset) {
	p.record()
	source := append([]byte(nil), p.source...)
	p.source = source
	p.steps = presetSteps(preset.Steps)
	p.Compute()
}
