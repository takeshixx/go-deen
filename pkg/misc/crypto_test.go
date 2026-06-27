package misc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"os"
	"testing"
)

func TestAESGCMRoundTrip(t *testing.T) {
	p := NewPluginAES()
	args := []string{"-mode", "gcm", "-key", "000102030405060708090a0b0c0d0e0f", "-nonce", "000102030405060708090a0b"}
	ciphertext := runMisc(t, p.RegisterFlags, p.Process, []byte("secret"), args...)
	plain := runMisc(t, p.RegisterFlags, p.Unprocess, ciphertext, args...)
	if string(plain) != "secret" {
		t.Fatalf("AES roundtrip = %q", string(plain))
	}
}

func TestAESAcceptsBase64KeyAndNonce(t *testing.T) {
	p := NewPluginAES()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	nonce := base64.StdEncoding.EncodeToString([]byte("abcdefghijkl"))
	args := []string{"-mode", "gcm", "-key", key, "-nonce", nonce}
	ciphertext := runMisc(t, p.RegisterFlags, p.Process, []byte("secret"), args...)
	plain := runMisc(t, p.RegisterFlags, p.Unprocess, ciphertext, args...)
	if string(plain) != "secret" {
		t.Fatalf("AES base64 roundtrip = %q", string(plain))
	}
}

func TestAESGCMRejectsTamperedCiphertext(t *testing.T) {
	p := NewPluginAES()
	args := []string{"-mode", "gcm", "-key", "000102030405060708090a0b0c0d0e0f", "-nonce", "000102030405060708090a0b"}
	ciphertext := runMisc(t, p.RegisterFlags, p.Process, []byte("secret"), args...)
	ciphertext[len(ciphertext)-1] ^= 0xff
	if err := runMiscErr(p.RegisterFlags, p.Unprocess, ciphertext, args...); err == nil {
		t.Fatal("expected tampered ciphertext to fail")
	}
}

func TestChaChaRoundTrip(t *testing.T) {
	p := NewPluginChaCha20Poly1305()
	args := []string{"-key", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "-nonce", "000102030405060708090a0b"}
	ciphertext := runMisc(t, p.RegisterFlags, p.Process, []byte("secret"), args...)
	plain := runMisc(t, p.RegisterFlags, p.Unprocess, ciphertext, args...)
	if string(plain) != "secret" {
		t.Fatalf("ChaCha roundtrip = %q", string(plain))
	}
}

func TestChaChaAcceptsRawBase64URLKeyAndNonce(t *testing.T) {
	p := NewPluginChaCha20Poly1305()
	key := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	nonce := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 12))
	args := []string{"-key", key, "-nonce", nonce}
	ciphertext := runMisc(t, p.RegisterFlags, p.Process, []byte("secret"), args...)
	plain := runMisc(t, p.RegisterFlags, p.Unprocess, ciphertext, args...)
	if string(plain) != "secret" {
		t.Fatalf("ChaCha base64 roundtrip = %q", string(plain))
	}
}

func TestSignEd25519(t *testing.T) {
	p := NewPluginSign()
	seed := bytes.Repeat([]byte{1}, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	sigHex := string(runMisc(t, p.RegisterFlags, p.Process, []byte("message"), "-alg", "ed25519", "-key", hex.EncodeToString(seed)))
	out := runMisc(t, p.RegisterFlags, p.Unprocess, []byte("message"), "-alg", "ed25519", "-pub", hex.EncodeToString(pub), "-sig", sigHex)
	if string(out) != "valid" {
		t.Fatalf("verify output = %q", string(out))
	}
}

func TestSignEd25519AcceptsBase64KeysAndSignature(t *testing.T) {
	p := NewPluginSign()
	seed := bytes.Repeat([]byte{2}, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	sigHex := string(runMisc(t, p.RegisterFlags, p.Process, []byte("message"), "-alg", "ed25519", "-key", base64.StdEncoding.EncodeToString(seed)))
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		t.Fatal(err)
	}
	out := runMisc(t, p.RegisterFlags, p.Unprocess, []byte("message"), "-alg", "ed25519", "-pub", base64.StdEncoding.EncodeToString(pub), "-sig", base64.StdEncoding.EncodeToString(sig))
	if string(out) != "valid" {
		t.Fatalf("verify output = %q", string(out))
	}
}

func TestSignEd25519RejectsInvalidSignature(t *testing.T) {
	p := NewPluginSign()
	seed := bytes.Repeat([]byte{3}, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	sigHex := string(runMisc(t, p.RegisterFlags, p.Process, []byte("message"), "-alg", "ed25519", "-key", hex.EncodeToString(seed)))
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		t.Fatal(err)
	}
	sig[0] ^= 0xff
	if err := runMiscErr(p.RegisterFlags, p.Unprocess, []byte("message"), "-alg", "ed25519", "-pub", hex.EncodeToString(pub), "-sig", hex.EncodeToString(sig)); err == nil {
		t.Fatal("expected invalid signature to fail")
	}
}

func TestSignRSAAndECDSA(t *testing.T) {
	for _, tc := range []struct {
		name string
		alg  string
		priv any
		pub  any
	}{
		{"rsa", "rsa-pss", mustRSAKey(t), nil},
		{"ecdsa", "ecdsa", mustECDSAKey(t), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var pub any
			switch k := tc.priv.(type) {
			case *rsa.PrivateKey:
				pub = &k.PublicKey
			case *ecdsa.PrivateKey:
				pub = &k.PublicKey
			}
			privPath := writePEMKey(t, "PRIVATE KEY", mustMarshalPKCS8(t, tc.priv))
			pubPath := writePEMKey(t, "PUBLIC KEY", mustMarshalPKIX(t, pub))
			p := NewPluginSign()
			sigHex := string(runMisc(t, p.RegisterFlags, p.Process, []byte("message"), "-alg", tc.alg, "-key", privPath))
			out := runMisc(t, p.RegisterFlags, p.Unprocess, []byte("message"), "-alg", tc.alg, "-pub", pubPath, "-sig", sigHex)
			if string(out) != "valid" {
				t.Fatalf("verify output = %q", string(out))
			}
		})
	}
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func mustECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func mustMarshalPKCS8(t *testing.T, key any) []byte {
	t.Helper()
	b, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mustMarshalPKIX(t *testing.T, key any) []byte {
	t.Helper()
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func writePEMKey(t *testing.T, typ string, der []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "key-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	if err := pem.Encode(f, &pem.Block{Type: typ, Bytes: der}); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestAESRejectsMissingKey(t *testing.T) {
	p := NewPluginAES()
	if err := p.Process(bytes.NewReader([]byte("x")), &bytes.Buffer{}, flag.NewFlagSet("aes", flag.ContinueOnError)); err == nil {
		t.Fatal("expected missing key to fail")
	}
}
