package pipeline

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"net/url"
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
			Name:        "JWT claim tampering check",
			Description: "Decode a signed JWT, change the user claim with jq, recreate the token with the original signature, then decode it again.",
			Source:      []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30"),
			Steps: []PresetStep{
				{Plugin: "jwt", Unprocess: true},
				{Plugin: "jq", Options: map[string]string{
					"q":        `.payload.name = "Max Power"`,
					"no-color": "true",
				}},
				{Plugin: "jwt", Options: map[string]string{"r": "true"}},
				{Plugin: "jwt", Unprocess: true},
			},
			WantContains: `"name": "Max Power"`,
		},
		{
			Name:        "Compressed webhook payload",
			Description: "URL-decode a copied request parameter, Base64-decode and gunzip it, then use jq to extract the fields worth triaging.",
			Source:      urlEscapedBase64Gzip([]byte(`{"event":"checkout.session.completed","actor":{"id":"user_123","ip":"203.0.113.42"},"target":"invoice_987","risk":{"score":87,"signals":["new_device","impossible_travel"]},"received_at":"2024-01-15T10:23:54Z"}`)),
			Steps: []PresetStep{
				{Plugin: "url", Unprocess: true},
				{Plugin: "base64", Unprocess: true},
				{Plugin: "gzip", Unprocess: true},
				{Plugin: "jq", Options: map[string]string{
					"q":        `{event, actor: .actor.id, target, risk: .risk.score, signals: .risk.signals}`,
					"no-color": "true",
				}},
			},
			WantContains: `"actor": "user_123"`,
		},
		{
			Name:        "SAML redirect request",
			Description: "Decode a URL-escaped SAMLRequest parameter from an HTTP redirect into readable XML.",
			Source:      samlRedirect([]byte(`<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="_a123" Version="2.0" IssueInstant="2024-01-15T10:23:54Z" AssertionConsumerServiceURL="https://app.example.test/sso/callback"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">https://idp.example.test</saml:Issuer><samlp:NameIDPolicy Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress" AllowCreate="true"/></samlp:AuthnRequest>`)),
			Steps: []PresetStep{
				{Plugin: "saml", Options: map[string]string{"url": "true", "deflate": "true"}},
			},
			WantContains: "https://idp.example.test",
		},
		{
			Name:        "Protobuf login request",
			Description: "Inspect a captured binary gRPC-style login request without a .proto schema.",
			Source:      base64Input(loginRequestProto()),
			Steps: []PresetStep{
				{Plugin: "base64", Unprocess: true},
				{Plugin: "protobuf"},
			},
			WantContains: `2: string "correct-horse-battery-staple"`,
		},
		{
			Name:        "URL token policy extraction",
			Description: "Unescape and decode a Base64 JSON policy token, then pull out the access decision and service list.",
			Source:      []byte(url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(`{"principal":"svc-billing","tenant":"acme-prod","decision":"allow","scope":["invoice:read","invoice:refund"],"conditions":{"mfa":true,"source_cidr":"10.10.0.0/16"}}`)))),
			Steps: []PresetStep{
				{Plugin: "url", Unprocess: true},
				{Plugin: "base64", Unprocess: true},
				{Plugin: "jq", Options: map[string]string{
					"q":        `{principal, decision, scope, source: .conditions.source_cidr}`,
					"no-color": "true",
				}},
			},
			WantContains: `"principal": "svc-billing"`,
		},
		{
			Name:        "MessagePack event from logs",
			Description: "Decode a Base64 MessagePack log event and query the fields needed for incident notes.",
			Source:      base64Input([]byte{0x84, 0xa2, 'i', 'd', 0xa7, 'e', 'v', 't', '_', '7', '3', '1', 0xa4, 't', 'y', 'p', 'e', 0xaa, 'a', 'p', 'i', '.', 'd', 'e', 'n', 'i', 'e', 'd', 0xa5, 'a', 'l', 'e', 'r', 't', 0xc3, 0xa6, 's', 'o', 'u', 'r', 'c', 'e', 0xac, 'e', 'd', 'g', 'e', '-', 'g', 'a', 't', 'e', 'w', 'a', 'y'}),
			Steps: []PresetStep{
				{Plugin: "base64", Unprocess: true},
				{Plugin: "msgpack", Unprocess: true},
				{Plugin: "jq", Options: map[string]string{
					"q":        `{id, type, alert, source}`,
					"no-color": "true",
				}},
			},
			WantContains: `"type": "api.denied"`,
		},
		{
			Name:        "CBOR device attestation",
			Description: "Decode a Base64 CBOR attestation blob and reduce it to the device state that matters.",
			Source:      base64Input([]byte{0xa4, 0x66, 'd', 'e', 'v', 'i', 'c', 'e', 0x6b, 'i', 'o', 's', '-', 'p', 'h', 'o', 'n', 'e', '-', '7', 0x66, 's', 't', 'a', 't', 'u', 's', 0x69, 'j', 'a', 'i', 'l', 'b', 'r', 'e', 'a', 'k', 0x6a, 's', 'e', 'q', 'u', 'e', 'n', 'c', 'e', 'I', 'd', 0x1a, 0x00, 0x10, 0x4f, 0xa3, 0x66, 'c', 'l', 'a', 'i', 'm', 's', 0x82, 0x6b, 's', 'e', 'c', 'u', 'r', 'e', '-', 'b', 'o', 'o', 't', 0x68, 'b', 'i', 'o', 'm', 'e', 't', 'r', 'y'}),
			Steps: []PresetStep{
				{Plugin: "base64", Unprocess: true},
				{Plugin: "cbor", Unprocess: true},
				{Plugin: "jq", Options: map[string]string{
					"q":        `{device, status, claims}`,
					"no-color": "true",
				}},
			},
			WantContains: `"status": "jailbreak"`,
		},
		{
			Name:        "DNS name from packet bytes",
			Description: "Decode a hex-encoded DNS label sequence from a packet capture into the queried host name.",
			Source:      []byte("2161646d696e2d617574682d736572766963652d70726f642d656467652d30303178076578616d706c6503636f6d00"),
			Steps: []PresetStep{
				{Plugin: "hex", Unprocess: true},
				{Plugin: "dns", Unprocess: true},
			},
			WantContains: "admin-auth-service-prod-edge-001x.example.com.",
		},
		{
			Name:        "AES-GCM API secret",
			Description: "Base64-decode an encrypted API secret, decrypt it with AES-GCM and authenticated data, then format the JSON.",
			Source:      aesGCMBase64([]byte(`{"api_key":"sk_live_redacted","scope":["invoice:read"],"tenant":"acme-prod"}`), "000102030405060708090a0b0c0d0e0f", "000102030405060708090a0b", "request-id=req_123"),
			Steps: []PresetStep{
				{Plugin: "base64", Unprocess: true},
				{Plugin: "aes", Unprocess: true, Options: map[string]string{
					"mode":  "gcm",
					"key":   "000102030405060708090a0b0c0d0e0f",
					"nonce": "000102030405060708090a0b",
					"aad":   "request-id=req_123",
				}},
				{Plugin: "json"},
			},
			WantContains: `"api_key"`,
		},
		{
			Name:        "Create encrypted support fixture",
			Description: "Take a JSON support case, encrypt it with AES-GCM, then Base64-encode the ciphertext for sharing in a ticket.",
			Source:      []byte(`{"ticket":"SEC-1842","tenant":"acme-prod","token_seen":true,"sample":"redacted"}`),
			Steps: []PresetStep{
				{
					Plugin: "aes",
					Options: map[string]string{
						"mode":  "gcm",
						"key":   "000102030405060708090a0b0c0d0e0f",
						"nonce": "000102030405060708090a0b",
						"aad":   "ticket=SEC-1842",
					},
				},
				{Plugin: "base64"},
			},
		},
		{
			Name:        "QR payload fixture",
			Description: "Format an incident handoff URL as a compact JSON object, create a QR PNG, then Base64-encode it for embedding.",
			Source:      []byte(`{"case":"SEC-1842","url":"https://deen.adversec.com/cases/SEC-1842","expires":"2024-01-16T10:23:54Z"}`),
			Steps: []PresetStep{
				{Plugin: "jq", Options: map[string]string{
					"q":     `.url`,
					"plain": "true",
				}},
				{
					Plugin:  "qr",
					Options: map[string]string{"size": "192"},
				},
				{Plugin: "base64"},
			},
			WantContains: "iVBOR",
		},
	}
}

func base64Input(data []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(data))
}

func aesGCMBase64(plain []byte, keyHex, nonceHex, aad string) []byte {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil
	}
	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}
	return base64Input(gcm.Seal(nil, nonce, plain, []byte(aad)))
}

func urlEscapedBase64Gzip(data []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(data)
	_ = gz.Close()
	return []byte(url.QueryEscape(base64.StdEncoding.EncodeToString(buf.Bytes())))
}

func samlRedirect(xml []byte) []byte {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil
	}
	_, _ = w.Write(xml)
	_ = w.Close()
	return []byte("SAMLRequest=" + url.QueryEscape(base64.StdEncoding.EncodeToString(buf.Bytes())) + "&RelayState=prod")
}

func loginRequestProto() []byte {
	var out []byte
	out = appendProtoString(out, 1, "alice@example.com")
	out = appendProtoString(out, 2, "correct-horse-battery-staple")
	var device []byte
	device = appendProtoString(device, 1, "ios")
	device = appendProtoString(device, 2, "17.3")
	out = appendProtoBytes(out, 3, device)
	out = appendProtoVarint(out, 4<<3|0)
	out = appendProtoVarint(out, 1)
	return out
}

func appendProtoString(out []byte, field uint64, value string) []byte {
	return appendProtoBytes(out, field, []byte(value))
}

func appendProtoBytes(out []byte, field uint64, value []byte) []byte {
	out = appendProtoVarint(out, field<<3|2)
	out = appendProtoVarint(out, uint64(len(value)))
	return append(out, value...)
}

func appendProtoVarint(out []byte, value uint64) []byte {
	for value >= 0x80 {
		out = append(out, byte(value)|0x80)
		value >>= 7
	}
	return append(out, byte(value))
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
