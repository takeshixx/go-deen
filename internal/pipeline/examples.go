package pipeline

import (
	"encoding/base64"
	"strings"
)

// Example is a runnable sample that loads both input data and a transform chain.
type Example struct {
	Name         string
	Description  string
	Source       []byte
	Steps        []PresetStep
	WantContains string
}

// BuiltinExamples returns runnable samples with concrete input data.
func BuiltinExamples() []Example {
	return []Example{
		{
			Name:         "Base64 encode text",
			Description:  "Encode plain text as Base64.",
			Source:       []byte("hello world"),
			Steps:        []PresetStep{{Plugin: "base64"}},
			WantContains: "aGVsbG8gd29ybGQ=",
		},
		{
			Name:         "Decode JWT",
			Description:  "Decode a small unsigned JWT into readable header and payload JSON.",
			Source:       []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.e30"),
			Steps:        []PresetStep{{Plugin: "jwt", Unprocess: true}},
			WantContains: `"sub": "123"`,
		},
		{
			Name:        "URL Base64 JSON",
			Description: "URL-decode a parameter value, Base64-decode it, then format the JSON.",
			Source:      []byte("eyJvayI6dHJ1ZX0%3D"),
			Steps: []PresetStep{
				{Plugin: "url", Unprocess: true},
				{Plugin: "base64", Unprocess: true},
				{Plugin: "json"},
			},
			WantContains: `"ok": true`,
		},
		{
			Name:         "Inspect protobuf wire data",
			Description:  "Decode schema-less protobuf bytes with a varint and a string field.",
			Source:       base64Input([]byte{0x08, 0x96, 0x01, 0x12, 0x04, 'd', 'e', 'e', 'n'}),
			Steps:        []PresetStep{{Plugin: "base64", Unprocess: true}, {Plugin: "protobuf"}},
			WantContains: `2: string "deen"`,
		},
		{
			Name:         "Decode MessagePack",
			Description:  "Decode a MessagePack map into formatted JSON.",
			Source:       base64Input([]byte{0x81, 0xa2, 'o', 'k', 0xc3}),
			Steps:        []PresetStep{{Plugin: "base64", Unprocess: true}, {Plugin: "msgpack", Unprocess: true}},
			WantContains: `"ok": true`,
		},
		{
			Name:         "Decode CBOR",
			Description:  "Decode a CBOR map into formatted JSON.",
			Source:       base64Input([]byte{0xa1, 0x62, 'o', 'k', 0xf5}),
			Steps:        []PresetStep{{Plugin: "base64", Unprocess: true}, {Plugin: "cbor", Unprocess: true}},
			WantContains: `"ok": true`,
		},
		{
			Name:         "Decode DNS wire name",
			Description:  "Decode DNS label bytes into a dotted name.",
			Source:       base64Input([]byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}),
			Steps:        []PresetStep{{Plugin: "base64", Unprocess: true}, {Plugin: "dns", Unprocess: true}},
			WantContains: "www.example.com.",
		},
		{
			Name:         "Inspect UUID",
			Description:  "Inspect UUID version, variant, and raw bytes.",
			Source:       []byte("550e8400-e29b-41d4-a716-446655440000"),
			Steps:        []PresetStep{{Plugin: "uuid", Options: map[string]string{"info": "true"}}},
			WantContains: "version: 4",
		},
		{
			Name:        "Extract IDs with regex",
			Description: "Extract capture group values from text.",
			Source:      []byte("id=123 id=456"),
			Steps: []PresetStep{{
				Plugin:  "regex",
				Options: map[string]string{"re": `id=(\d+)`, "group": "1"},
			}},
			WantContains: "123\n456",
		},
		{
			Name:        "Milliseconds timestamp",
			Description: "Convert Unix milliseconds into an RFC3339 timestamp.",
			Source:      []byte("1700000000123"),
			Steps: []PresetStep{{
				Plugin:  "timestamp",
				Options: map[string]string{"unit": "ms"},
			}},
			WantContains: "2023-11-14T22:13:20.123Z",
		},
		{
			Name:        "AES-GCM encrypt",
			Description: "Encrypt text with AES-GCM using a fixed sample key and nonce, then encode the ciphertext as Base64.",
			Source:      []byte("secret"),
			Steps: []PresetStep{
				{
					Plugin: "aes",
					Options: map[string]string{
						"mode":  "gcm",
						"key":   "000102030405060708090a0b0c0d0e0f",
						"nonce": "000102030405060708090a0b",
					},
				},
				{Plugin: "base64"},
			},
		},
		{
			Name:        "QR code",
			Description: "Create a QR PNG from a URL, then encode the PNG bytes as Base64.",
			Source:      []byte("https://deen.adversec.com"),
			Steps: []PresetStep{
				{
					Plugin:  "qr",
					Options: map[string]string{"size": "192"},
				},
				{Plugin: "base64"},
			},
			WantContains: "iVBOR",
		},
		{
			Name:         "Detect PDF magic",
			Description:  "Identify a PDF payload from its magic bytes.",
			Source:       []byte("%PDF-1.7\n"),
			Steps:        []PresetStep{{Plugin: "magic"}},
			WantContains: "PDF document",
		},
	}
}

func base64Input(data []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(data))
}

// ExampleResult runs an example in an isolated pipeline and returns its result.
func ExampleResult(example Example) ([]byte, error) {
	p := New()
	p.ApplyExample(example)
	for _, step := range p.Steps() {
		if step.err != nil {
			return p.Result(), step.err
		}
	}
	return p.Result(), nil
}

// ExampleMatches reports whether an example matches a user search query.
func ExampleMatches(example Example, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	var parts []string
	parts = append(parts, example.Name, example.Description, string(example.Source), example.WantContains)
	for _, step := range example.Steps {
		parts = append(parts, step.Plugin)
		if step.Unprocess {
			parts = append(parts, "."+step.Plugin, "decode")
		} else {
			parts = append(parts, "encode")
		}
		for k, v := range step.Options {
			parts = append(parts, k, v)
		}
	}
	return strings.Contains(strings.ToLower(strings.Join(parts, " ")), query)
}

// ApplyExample replaces the current input and chain with a runnable example.
// The previous pipeline state remains undoable.
func (p *Pipeline) ApplyExample(example Example) {
	p.record()
	p.source = append([]byte(nil), example.Source...)
	p.steps = presetSteps(example.Steps)
	p.Compute()
}

func presetSteps(steps []PresetStep) []*Step {
	out := make([]*Step, 0, len(steps))
	for _, ps := range steps {
		opts := make(map[string]string, len(ps.Options))
		for k, v := range ps.Options {
			opts[k] = v
		}
		out = append(out, &Step{
			Plugin:    ps.Plugin,
			Unprocess: ps.Unprocess,
			Options:   opts,
		})
	}
	return out
}
