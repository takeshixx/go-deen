package misc

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// firstCertAndKey extracts the first CERTIFICATE and the first private key
// found in PEM-encoded data. Either may be nil if not present.
func firstCertAndKey(data []byte) (cert *x509.Certificate, key crypto.PrivateKey, err error) {
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		switch {
		case block.Type == "CERTIFICATE" && cert == nil:
			cert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, nil, err
			}
		case strings.Contains(block.Type, "PRIVATE KEY") && key == nil:
			key, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, nil, err
			}
		}
		data = rest
	}
	return cert, key, nil
}

// parsePrivateKey parses a DER-encoded private key in PKCS#1, PKCS#8 or SEC1
// (EC) form.
func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("unsupported private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("failed to parse private key")
}

// generateMatchingKey creates a new private key of the same type (and size or
// curve) as the given certificate's public key.
func generateMatchingKey(cert *x509.Certificate) (crypto.Signer, error) {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return rsa.GenerateKey(rand.Reader, pub.N.BitLen())
	case *ecdsa.PublicKey:
		return ecdsa.GenerateKey(pub.Curve, rand.Reader)
	case ed25519.PublicKey:
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		return priv, err
	default:
		return nil, fmt.Errorf("unsupported public key type %T", cert.PublicKey)
	}
}

// publicKeyAlgorithm reports the algorithm of a public key.
func publicKeyAlgorithm(pub crypto.PublicKey) x509.PublicKeyAlgorithm {
	switch pub.(type) {
	case *rsa.PublicKey:
		return x509.RSA
	case *ecdsa.PublicKey:
		return x509.ECDSA
	case ed25519.PublicKey:
		return x509.Ed25519
	default:
		return x509.UnknownPublicKeyAlgorithm
	}
}

// sigAlgKeyType reports which public key algorithm a signature algorithm uses.
func sigAlgKeyType(sa x509.SignatureAlgorithm) x509.PublicKeyAlgorithm {
	switch sa {
	case x509.MD5WithRSA, x509.SHA1WithRSA, x509.SHA256WithRSA, x509.SHA384WithRSA,
		x509.SHA512WithRSA, x509.SHA256WithRSAPSS, x509.SHA384WithRSAPSS, x509.SHA512WithRSAPSS:
		return x509.RSA
	case x509.ECDSAWithSHA1, x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
		return x509.ECDSA
	case x509.PureEd25519:
		return x509.Ed25519
	default:
		return x509.UnknownPublicKeyAlgorithm
	}
}

// signatureAlgorithmByKey maps a (key algorithm, hash) pair to the matching
// x509.SignatureAlgorithm. An empty hash returns the algorithm's default.
var rsaSigAlgs = map[string]x509.SignatureAlgorithm{
	"": x509.SHA256WithRSA, "sha256": x509.SHA256WithRSA,
	"sha384": x509.SHA384WithRSA, "sha512": x509.SHA512WithRSA, "sha1": x509.SHA1WithRSA,
}
var ecdsaSigAlgs = map[string]x509.SignatureAlgorithm{
	"": x509.ECDSAWithSHA256, "sha256": x509.ECDSAWithSHA256,
	"sha384": x509.ECDSAWithSHA384, "sha512": x509.ECDSAWithSHA512, "sha1": x509.ECDSAWithSHA1,
}

// hashToken extracts a normalised hash name (e.g. "sha256") from a user-provided
// signature-algorithm string such as "SHA256", "ecdsa-with-SHA256" or
// "SHA256-RSA".
func hashToken(s string) string {
	s = strings.ToLower(s)
	for _, h := range []string{"sha512", "sha384", "sha256", "sha1"} {
		if strings.Contains(s, h) {
			return h
		}
	}
	return ""
}

// resolveSignatureAlgorithm decides the signature algorithm for the new
// certificate based on the requested name, the original certificate and the key
// that will actually sign (self key or CA key).
func resolveSignatureAlgorithm(requested string, original x509.SignatureAlgorithm, signerKey x509.PublicKeyAlgorithm) (x509.SignatureAlgorithm, error) {
	if requested != "" {
		hash := hashToken(requested)
		switch signerKey {
		case x509.RSA:
			return rsaSigAlgs[hash], nil
		case x509.ECDSA:
			return ecdsaSigAlgs[hash], nil
		case x509.Ed25519:
			return x509.PureEd25519, nil
		default:
			return 0, fmt.Errorf("unsupported signing key type")
		}
	}
	// Inherit the original algorithm when it matches the signing key type;
	// otherwise return 0 so x509 picks a sensible default for the key.
	if sigAlgKeyType(original) == signerKey {
		return original, nil
	}
	return 0, nil
}

// NewPluginCertCloner creates a new certificate cloner plugin.
func NewPluginCertCloner() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "certCloner"
	p.Aliases = []string{"clone"}
	p.Category = "misc"
	p.Description = "Clone an x509 certificate with a freshly generated key of the\nsame type. Self-signs by default, or signs with a given CA."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("ca-cert", "", "CA certificate (PEM); may also contain the CA key")
		flags.String("ca-key", "", "CA private key (PEM), if not in -ca-cert")
		flags.String("sig-alg", "", "signature hash algorithm (sha256/sha384/sha512); default inherits the original")
		flags.String("o", "", "also write <prefix>.crt, <prefix>.key and <prefix>.pem files")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		original, _, err := firstCertAndKey(input)
		if err != nil {
			return err
		}
		if original == nil {
			return fmt.Errorf("no certificate found in input")
		}

		newKey, err := generateMatchingKey(original)
		if err != nil {
			return err
		}
		newPub := newKey.Public()

		// Build a clean template that preserves the original's appearance.
		// Only raw fields are copied; typed extension fields are left zero so
		// x509.CreateCertificate does not emit duplicate extensions.
		template := &x509.Certificate{
			SerialNumber:    original.SerialNumber,
			RawSubject:      original.RawSubject,
			NotBefore:       original.NotBefore,
			NotAfter:        original.NotAfter,
			ExtraExtensions: original.Extensions,
		}

		var parent *x509.Certificate
		var signerKey crypto.PrivateKey

		caCertVal := helpers.StringFlag(flags, "ca-cert")
		caKeyVal := helpers.StringFlag(flags, "ca-key")
		if caCertVal != "" {
			caCert, caKey, err := loadCA(caCertVal, caKeyVal)
			if err != nil {
				return err
			}
			parent = caCert
			signerKey = caKey
		} else {
			// Self-sign: the issuer keeps the original's issuer DN and the new
			// key signs the certificate.
			parent = &x509.Certificate{RawSubject: original.RawIssuer}
			signerKey = newKey
		}

		signerPub, ok := signerKey.(crypto.Signer)
		if !ok {
			return fmt.Errorf("signing key does not implement crypto.Signer")
		}
		sigAlg, err := resolveSignatureAlgorithm(
			helpers.StringFlag(flags, "sig-alg"),
			original.SignatureAlgorithm,
			publicKeyAlgorithm(signerPub.Public()),
		)
		if err != nil {
			return err
		}
		template.SignatureAlgorithm = sigAlg

		derBytes, err := x509.CreateCertificate(rand.Reader, template, parent, newPub, signerKey)
		if err != nil {
			return err
		}

		keyDER, err := x509.MarshalPKCS8PrivateKey(newKey)
		if err != nil {
			return err
		}
		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

		if prefix := helpers.StringFlag(flags, "o"); prefix != "" {
			if err := writeCloneFiles(prefix, certPEM, keyPEM); err != nil {
				return err
			}
		}

		if _, err := w.Write(certPEM); err != nil {
			return err
		}
		_, err = w.Write(keyPEM)
		return err
	}
	return p
}

// loadCA loads a CA certificate and private key. The key may live in the same
// file as the certificate, or in a separate file given by keyPath.
func loadCA(certPath, keyPath string) (*x509.Certificate, crypto.PrivateKey, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}
	cert, key, err := firstCertAndKey(certData)
	if err != nil {
		return nil, nil, err
	}
	if cert == nil {
		return nil, nil, fmt.Errorf("no CA certificate found in %s", certPath)
	}
	if keyPath != "" {
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, nil, err
		}
		_, key, err = firstCertAndKey(keyData)
		if err != nil {
			return nil, nil, err
		}
	}
	if key == nil {
		return nil, nil, fmt.Errorf("no CA private key found (provide -ca-key)")
	}
	return cert, key, nil
}

// writeCloneFiles writes the certificate, key and a combined PEM bundle.
func writeCloneFiles(prefix string, certPEM, keyPEM []byte) error {
	if err := os.WriteFile(prefix+".crt", certPEM, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(prefix+".key", keyPEM, 0o600); err != nil {
		return err
	}
	combined := append(append([]byte{}, certPEM...), keyPEM...)
	return os.WriteFile(prefix+".pem", combined, 0o600)
}
