package misc

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// makeTestCert builds a self-signed certificate with a couple of extensions and
// returns it PEM-encoded together with the parsed certificate.
func makeTestCert(t *testing.T, key crypto.Signer) ([]byte, *x509.Certificate) {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(0x1234abcd),
		Subject:               pkix.Name{CommonName: "clone.example.com", Organization: []string{"deen"}},
		DNSNames:              []string{"clone.example.com", "www.clone.example.com"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), cert
}

// runClone feeds certPEM through the cloner plugin with the given args and
// returns the parsed cloned certificate and its private key.
func runClone(t *testing.T, certPEM []byte, args ...string) (*x509.Certificate, crypto.PrivateKey) {
	t.Helper()
	p := NewPluginCertCloner()
	fs := flag.NewFlagSet("certCloner", flag.ContinueOnError)
	p.RegisterFlags(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := p.Process(bytes.NewReader(certPEM), &out, fs); err != nil {
		t.Fatalf("clone failed: %s", err)
	}
	cert, key, err := firstCertAndKey(out.Bytes())
	if err != nil {
		t.Fatalf("clone output did not parse: %s", err)
	}
	if cert == nil || key == nil {
		t.Fatalf("clone output missing cert (%v) or key (%v)", cert != nil, key != nil)
	}
	return cert, key
}

func pubBytes(t *testing.T, pub crypto.PublicKey) []byte {
	t.Helper()
	b, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func assertClonePreserved(t *testing.T, orig, clone *x509.Certificate) {
	t.Helper()
	if !bytes.Equal(orig.RawSubject, clone.RawSubject) {
		t.Errorf("subject not preserved: %s vs %s", orig.Subject, clone.Subject)
	}
	if orig.SerialNumber.Cmp(clone.SerialNumber) != 0 {
		t.Errorf("serial not preserved: %s vs %s", orig.SerialNumber, clone.SerialNumber)
	}
	if bytes.Equal(pubBytes(t, orig.PublicKey), pubBytes(t, clone.PublicKey)) {
		t.Error("public key was not regenerated")
	}
	// SANs are carried as a copied extension.
	if len(clone.DNSNames) != len(orig.DNSNames) {
		t.Errorf("SANs not preserved: %v vs %v", orig.DNSNames, clone.DNSNames)
	}
}

func TestCloneRSASelfSigned(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, orig := makeTestCert(t, key)
	clone, cloneKey := runClone(t, certPEM)
	assertClonePreserved(t, orig, clone)
	if _, ok := cloneKey.(*rsa.PrivateKey); !ok {
		t.Errorf("expected RSA clone key, got %T", cloneKey)
	}
}

func TestCloneECDSASelfSigned(t *testing.T) {
	// This is the case that previously crashed (PKCS1 marshal of an EC key).
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, orig := makeTestCert(t, key)
	clone, cloneKey := runClone(t, certPEM)
	assertClonePreserved(t, orig, clone)
	ecKey, ok := cloneKey.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("expected ECDSA clone key, got %T", cloneKey)
	}
	if ecKey.Curve != elliptic.P384() {
		t.Errorf("clone key curve = %s, want P-384", ecKey.Curve.Params().Name)
	}
}

func TestCloneCASigned(t *testing.T) {
	// Build a CA and write its cert+key to files.
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	caPEM, caCert := makeTestCert(t, caKey)
	caKeyDER, err := x509.MarshalPKCS8PrivateKey(caKey)
	if err != nil {
		t.Fatal(err)
	}
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: caKeyDER})

	dir := t.TempDir()
	caCertFile := filepath.Join(dir, "ca.crt")
	caKeyFile := filepath.Join(dir, "ca.key")
	if err := os.WriteFile(caCertFile, caPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caKeyFile, caKeyPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	leafPEM, _ := makeTestCert(t, leafKey)

	clone, _ := runClone(t, leafPEM, "-ca-cert", caCertFile, "-ca-key", caKeyFile)
	if err := clone.CheckSignatureFrom(caCert); err != nil {
		t.Errorf("CA-signed clone does not verify against the CA: %s", err)
	}
}

func TestCloneSigAlgOverride(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, _ := makeTestCert(t, key)
	clone, _ := runClone(t, certPEM, "-sig-alg", "sha512")
	if clone.SignatureAlgorithm != x509.SHA512WithRSA {
		t.Errorf("sig-alg override failed: got %s, want SHA512-RSA", clone.SignatureAlgorithm)
	}
}

func TestCloneFileOutput(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, _ := makeTestCert(t, key)
	prefix := filepath.Join(t.TempDir(), "out")
	runClone(t, certPEM, "-o", prefix)
	for _, ext := range []string{".crt", ".key", ".pem"} {
		if _, err := os.Stat(prefix + ext); err != nil {
			t.Errorf("expected output file %s: %s", prefix+ext, err)
		}
	}
}
