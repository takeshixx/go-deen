package misc

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// runCloneErr runs the cloner and returns the resulting error (if any).
func runCloneErr(certPEM []byte, args ...string) error {
	p := NewPluginCertCloner()
	fs := flag.NewFlagSet("certCloner", flag.ContinueOnError)
	p.RegisterFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	var out bytes.Buffer
	return p.Process(bytes.NewReader(certPEM), &out, fs)
}

func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCloneEd25519SelfSigned(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, orig := makeTestCert(t, key)
	clone, cloneKey := runClone(t, certPEM)
	assertClonePreserved(t, orig, clone)
	if _, ok := cloneKey.(ed25519.PrivateKey); !ok {
		t.Errorf("expected Ed25519 clone key, got %T", cloneKey)
	}
	if clone.SignatureAlgorithm != x509.PureEd25519 {
		t.Errorf("expected PureEd25519 signature, got %s", clone.SignatureAlgorithm)
	}
}

func TestCloneECDSASigAlgOverride(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, _ := makeTestCert(t, key)
	clone, _ := runClone(t, certPEM, "-sig-alg", "sha384")
	if clone.SignatureAlgorithm != x509.ECDSAWithSHA384 {
		t.Errorf("got %s, want ECDSA-SHA384", clone.SignatureAlgorithm)
	}
}

func TestCloneSigAlgUnknownDefaults(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, _ := makeTestCert(t, key)
	// "md5" is not a recognised token, so it falls back to the RSA default.
	clone, _ := runClone(t, certPEM, "-sig-alg", "md5")
	if clone.SignatureAlgorithm != x509.SHA256WithRSA {
		t.Errorf("got %s, want SHA256-RSA default", clone.SignatureAlgorithm)
	}
}

// caFiles builds a CA cert and writes its key in the requested PEM form,
// returning the cert path and key path (key path empty when combined).
func caFiles(t *testing.T, caKey crypto.Signer, keyForm string) (certPath, keyPath, dir string) {
	t.Helper()
	caPEM, _ := makeTestCert(t, caKey)
	dir = t.TempDir()

	var keyBlock *pem.Block
	switch keyForm {
	case "pkcs1":
		keyBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey.(*rsa.PrivateKey))}
	case "sec1":
		der, err := x509.MarshalECPrivateKey(caKey.(*ecdsa.PrivateKey))
		if err != nil {
			t.Fatal(err)
		}
		keyBlock = &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
	default: // pkcs8
		der, err := x509.MarshalPKCS8PrivateKey(caKey)
		if err != nil {
			t.Fatal(err)
		}
		keyBlock = &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	}
	keyPEM := pem.EncodeToMemory(keyBlock)

	if keyForm == "combined" {
		certPath = writeFile(t, dir, "ca.pem", append(append([]byte{}, caPEM...), keyPEM...))
		return certPath, "", dir
	}
	certPath = writeFile(t, dir, "ca.crt", caPEM)
	keyPath = writeFile(t, dir, "ca.key", keyPEM)
	return certPath, keyPath, dir
}

func cloneAndVerifyCA(t *testing.T, caKey crypto.Signer, keyForm string) {
	t.Helper()
	caPEM, caCert := makeTestCert(t, caKey)
	_ = caPEM
	certPath, keyPath, _ := caFiles(t, caKey, keyForm)

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	leafPEM, _ := makeTestCert(t, leafKey)

	args := []string{"-ca-cert", certPath}
	if keyPath != "" {
		args = append(args, "-ca-key", keyPath)
	}
	clone, _ := runClone(t, leafPEM, args...)
	if err := clone.CheckSignatureFrom(caCert); err != nil {
		t.Errorf("[%s] clone does not verify against CA: %s", keyForm, err)
	}
}

func TestCloneCACombinedFile(t *testing.T) {
	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	cloneAndVerifyCA(t, caKey, "combined")
}

func TestCloneCAKeyPKCS1(t *testing.T) {
	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	cloneAndVerifyCA(t, caKey, "pkcs1")
}

func TestCloneCAKeySEC1EC(t *testing.T) {
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cloneAndVerifyCA(t, caKey, "sec1")
}

func TestCloneNoCertInput(t *testing.T) {
	if err := runCloneErr([]byte("this is not a certificate")); err == nil {
		t.Error("expected an error for input without a certificate")
	}
}

func TestCloneMalformedCert(t *testing.T) {
	bad := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not der")})
	if err := runCloneErr(bad); err == nil {
		t.Error("expected an error for a malformed certificate block")
	}
}

func TestCloneCAMissingKey(t *testing.T) {
	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	caPEM, _ := makeTestCert(t, caKey)
	dir := t.TempDir()
	certPath := writeFile(t, dir, "ca-only.crt", caPEM)

	leafKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	leafPEM, _ := makeTestCert(t, leafKey)
	// CA cert without a key and no -ca-key flag should error.
	if err := runCloneErr(leafPEM, "-ca-cert", certPath); err == nil {
		t.Error("expected an error when the CA key is missing")
	}
}

func TestCloneMalformedKeyInInput(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	certPEM, _ := makeTestCert(t, key)
	badKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not a key")})
	input := append(append([]byte{}, certPEM...), badKey...)
	if err := runCloneErr(input); err == nil {
		t.Error("expected an error for a malformed private key block in the input")
	}
}

func TestCloneEd25519SigAlgIgnoresHash(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certPEM, _ := makeTestCert(t, key)
	// Ed25519 has no separate hash; the requested hash is ignored.
	clone, _ := runClone(t, certPEM, "-sig-alg", "sha256")
	if clone.SignatureAlgorithm != x509.PureEd25519 {
		t.Errorf("got %s, want PureEd25519", clone.SignatureAlgorithm)
	}
}

func TestCloneCAKeyOnlyFile(t *testing.T) {
	// -ca-cert pointing at a key-only file should error (no CA certificate).
	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	der, _ := x509.MarshalPKCS8PrivateKey(caKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	dir := t.TempDir()
	keyOnly := writeFile(t, dir, "key-only.pem", keyPEM)

	leafKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	leafPEM, _ := makeTestCert(t, leafKey)
	if err := runCloneErr(leafPEM, "-ca-cert", keyOnly); err == nil {
		t.Error("expected an error when -ca-cert has no certificate")
	}
}

func TestCloneMissingCAKeyFile(t *testing.T) {
	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	caPEM, _ := makeTestCert(t, caKey)
	dir := t.TempDir()
	certPath := writeFile(t, dir, "ca.crt", caPEM)
	leafKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	leafPEM, _ := makeTestCert(t, leafKey)
	if err := runCloneErr(leafPEM, "-ca-cert", certPath, "-ca-key", "/nonexistent/k.pem"); err == nil {
		t.Error("expected an error for a missing -ca-key file")
	}
}

func TestCloneMissingCAFile(t *testing.T) {
	leafKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	leafPEM, _ := makeTestCert(t, leafKey)
	if err := runCloneErr(leafPEM, "-ca-cert", "/nonexistent/ca.pem"); err == nil {
		t.Error("expected an error for a missing CA file")
	}
}
