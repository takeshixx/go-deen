package formatters

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"io"
	"strings"
	"testing"

	jose "github.com/go-jose/go-jose/v4"
)

func TestJWKPrettyPrintsKey(t *testing.T) {
	p := NewPluginJWK()
	input := []byte(`{"kty":"oct","kid":"hmac","k":"AyM1SysPpbyDfgZld3umTQ"}`)
	got := runJWK(t, p.Process, p.RegisterFlags, input)
	if !strings.Contains(got, "\n") || !strings.Contains(got, `"kid": "hmac"`) {
		t.Fatalf("expected pretty JWK output, got %s", got)
	}
}

func TestJWKCompactsKey(t *testing.T) {
	p := NewPluginJWK()
	input := []byte("{\n  \"kty\": \"oct\",\n  \"kid\": \"hmac\",\n  \"k\": \"AyM1SysPpbyDfgZld3umTQ\"\n}")
	got := runJWK(t, p.Unprocess, p.RegisterFlags, input)
	if strings.Contains(got, "\n") {
		t.Fatalf("expected compact JWK output, got %s", got)
	}
	if !strings.Contains(got, `"kid":"hmac"`) {
		t.Fatalf("expected kid in compact output, got %s", got)
	}
}

func TestJWKPublicOnlyRemovesPrivateMaterial(t *testing.T) {
	p := NewPluginJWK()
	key := rsaJWK(t)
	input, err := json.Marshal(key)
	if err != nil {
		t.Fatal(err)
	}
	got := runJWK(t, p.Process, p.RegisterFlags, input, "-public")
	if strings.Contains(got, `"d"`) || strings.Contains(got, `"p"`) || strings.Contains(got, `"q"`) {
		t.Fatalf("public JWK leaked private fields: %s", got)
	}
	if !strings.Contains(got, `"kty": "RSA"`) || !strings.Contains(got, `"n":`) {
		t.Fatalf("public RSA JWK missing public fields: %s", got)
	}
}

func TestJWKThumbprint(t *testing.T) {
	p := NewPluginJWK()
	key := rsaJWK(t)
	input, err := json.Marshal(key.Public())
	if err != nil {
		t.Fatal(err)
	}
	got := runJWK(t, p.Process, p.RegisterFlags, input, "-thumbprint")
	if !strings.Contains(got, `"kid": "rsa-test"`) || !strings.Contains(got, `"thumbprint_sha256":`) {
		t.Fatalf("thumbprint output missing fields: %s", got)
	}
}

func TestJWKSFormatsSet(t *testing.T) {
	p := NewPluginJWK()
	input := []byte(`{"keys":[{"kty":"oct","kid":"one","k":"AyM1SysPpbyDfgZld3umTQ"}]}`)
	got := runJWK(t, p.Process, p.RegisterFlags, input)
	if !strings.Contains(got, `"keys": [`) || !strings.Contains(got, `"kid": "one"`) {
		t.Fatalf("expected pretty JWKS output, got %s", got)
	}
}

func rsaJWK(t *testing.T) jose.JSONWebKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	return jose.JSONWebKey{Key: key, KeyID: "rsa-test"}
}

func tryJWK(fn func(io.Reader, io.Writer, *flag.FlagSet) error, registerFlags func(*flag.FlagSet), input []byte, args ...string) (string, error) {
	fs := flag.NewFlagSet("jwk", flag.ContinueOnError)
	if registerFlags != nil {
		registerFlags(fs)
	}
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	var out bytes.Buffer
	err := fn(bytes.NewReader(input), &out, fs)
	return out.String(), err
}

func runJWK(t *testing.T, fn func(io.Reader, io.Writer, *flag.FlagSet) error, registerFlags func(*flag.FlagSet), input []byte, args ...string) string {
	t.Helper()
	got, err := tryJWK(fn, registerFlags, input, args...)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
