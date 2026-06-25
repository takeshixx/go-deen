package formatters

import (
	"strings"
	"testing"
)

// A well-known HS256 token (the jwt.io sample) used to exercise the decode path.
const sampleHS256Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
	"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

func TestNewPluginJwt(t *testing.T) {
	p := NewPluginJwt()
	if p.Name != "jwt" {
		t.Errorf("unexpected plugin name: %s", p.Name)
	}
}

func TestListAlgs(t *testing.T) {
	out := listAlgs()
	for _, want := range []string{"HS256", "RS256", "ES256", "EdDSA"} {
		if !strings.Contains(out, want) {
			t.Errorf("listAlgs output missing %q", want)
		}
	}
}

// TestDecodeToken exercises the go-jose v4 decode path (ParseSigned now requires
// an explicit allow-list of signature algorithms).
func TestDecodeToken(t *testing.T) {
	p := NewPluginJwt()
	out, err := tryFormat(p.Unprocess, p.RegisterFlags, []byte(sampleHS256Token))
	if err != nil {
		t.Fatalf("jwt decode failed: %s", err)
	}
	got := string(out)
	for _, want := range []string{
		`"alg": "HS256"`,
		`"sub": "1234567890"`,
		`"name": "John Doe"`,
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("decoded token missing %q\ngot: %s", want, got)
		}
	}
}

// TestAllowedSignatureAlgorithms ensures the allow-list passed to ParseSigned
// covers the algorithms deen advertises support for.
func TestAllowedSignatureAlgorithms(t *testing.T) {
	if got := len(allowedSignatureAlgorithms()); got != len(signatureAlgs) {
		t.Errorf("allowed algorithms = %d, want %d", got, len(signatureAlgs))
	}
}
