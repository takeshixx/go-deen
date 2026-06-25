package misc

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"strings"
	"testing"
)

// TestCertClonerSelfSign clones a certificate and verifies the output contains
// a fresh certificate and private key that parse correctly.
func TestCertClonerSelfSign(t *testing.T) {
	p := NewPluginCertCloner()
	fs := flag.NewFlagSet("certCloner", flag.ContinueOnError)
	p.RegisterFlags(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := p.Process(strings.NewReader(certRSA), &out, fs); err != nil {
		t.Fatalf("certCloner failed: %s", err)
	}

	rest := out.Bytes()
	var sawCert, sawKey bool
	for {
		block, remainder := pem.Decode(rest)
		if block == nil {
			break
		}
		switch block.Type {
		case "CERTIFICATE":
			if _, err := x509.ParseCertificate(block.Bytes); err != nil {
				t.Errorf("cloned certificate does not parse: %s", err)
			}
			sawCert = true
		case "RSA PRIVATE KEY":
			if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
				t.Errorf("cloned private key does not parse: %s", err)
			}
			sawKey = true
		}
		rest = remainder
	}
	if !sawCert || !sawKey {
		t.Errorf("clone output missing cert (%v) or key (%v)", sawCert, sawKey)
	}
}
