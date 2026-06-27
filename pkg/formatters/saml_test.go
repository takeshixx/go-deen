package formatters

import (
	"bytes"
	"encoding/base64"
	"flag"
	"io"
	"net/url"
	"strings"
	"testing"
)

const samlXML = `<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="abc"><saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">issuer</saml:Issuer></samlp:AuthnRequest>`

func TestSAMLDecodePlainBase64(t *testing.T) {
	p := NewPluginSAML()
	input := []byte(base64.StdEncoding.EncodeToString([]byte(samlXML)))
	got := runSAML(t, p.Process, p.RegisterFlags, input)
	if !strings.Contains(got, "AuthnRequest") || !strings.Contains(got, "\n") {
		t.Fatalf("expected pretty SAML XML, got %s", got)
	}
}

func TestSAMLDecodeDeflatedQuery(t *testing.T) {
	p := NewPluginSAML()
	deflated := deflateRaw([]byte(samlXML))
	encoded := base64.StdEncoding.EncodeToString(deflated)
	query := "RelayState=x&SAMLRequest=" + url.QueryEscape(encoded)
	got := runSAML(t, p.Process, p.RegisterFlags, []byte(query), "-url")
	if !strings.Contains(got, "issuer") {
		t.Fatalf("expected decoded issuer, got %s", got)
	}
}

func TestSAMLEncodePlain(t *testing.T) {
	p := NewPluginSAML()
	got := runSAML(t, p.Unprocess, p.RegisterFlags, []byte(samlXML))
	decoded, err := base64.StdEncoding.DecodeString(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(decoded), "AuthnRequest") {
		t.Fatalf("expected XML in decoded payload, got %s", decoded)
	}
}

func TestSAMLEncodeDeflatedURL(t *testing.T) {
	p := NewPluginSAML()
	got := runSAML(t, p.Unprocess, p.RegisterFlags, []byte(samlXML), "-deflate", "-url")
	unescaped, err := url.QueryUnescape(got)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(unescaped)
	if err != nil {
		t.Fatal(err)
	}
	inflated, err := inflateRaw(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(inflated), "issuer") {
		t.Fatalf("expected issuer in inflated payload, got %s", inflated)
	}
}

func trySAML(fn func(io.Reader, io.Writer, *flag.FlagSet) error, registerFlags func(*flag.FlagSet), input []byte, args ...string) (string, error) {
	fs := flag.NewFlagSet("saml", flag.ContinueOnError)
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

func runSAML(t *testing.T, fn func(io.Reader, io.Writer, *flag.FlagSet) error, registerFlags func(*flag.FlagSet), input []byte, args ...string) string {
	t.Helper()
	got, err := trySAML(fn, registerFlags, input, args...)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
