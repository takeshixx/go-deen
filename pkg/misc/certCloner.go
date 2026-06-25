package misc

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// readCertsAndKeyFromFile calls readCertFromBuffer() with
// bytes read from a given file.
func readCertsAndKeyFromFile(path string) (cert *x509.Certificate, key interface{}, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return readCertFromBuffer(data)
}

// readCertFromBuffer parses a cert and key from a
// given byte slice.
func readCertFromBuffer(data []byte) (cert *x509.Certificate, key interface{}, err error) {
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			if cert != nil {
				return nil, nil, fmt.Errorf("additional certificate found, aborting")
			}
			cert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, nil, err
			}

		} else {
			if key != nil {
				return nil, nil, fmt.Errorf("additional private key found, aborting")
			}
			key, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, nil, err
			}
		}
		data = rest
	}
	return
}

// parsePrivateKey returns a private key parsed from the
// input bytes.
func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("failed to parse private key")
}

// generateKeyPair creates a random key pair with the same
// type as the given cert.
func generateKeyPair(cert *x509.Certificate) (newPrivateKey, newPublicKey interface{}, err error) {
	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		newPrivateKey, err = rsa.GenerateKey(rand.Reader, cert.PublicKey.(*rsa.PublicKey).Size()*8)
		if err != nil {
			return nil, nil, err
		}
		newPublicKey = &newPrivateKey.(*rsa.PrivateKey).PublicKey
	case x509.ECDSA:
		newPrivateKey, err = ecdsa.GenerateKey(cert.PublicKey.(elliptic.Curve), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		newPublicKey = &newPrivateKey.(*ecdsa.PrivateKey).PublicKey
	case x509.DSA:
	case x509.Ed25519:
		fmt.Printf("Public key algorithm %s not implemented yet.\n", cert.PublicKeyAlgorithm)
	}
	return
}

// NewPluginCertCloner creates a new certificate cloner plugin.
func NewPluginCertCloner() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "certCloner"
	p.Aliases = []string{"clone"}
	p.Category = "misc"
	p.Description = "x509 certificate cloner."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("ca-cert", "", "CA certificate and private key in PEM format")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		caCertVal := helpers.StringFlag(flags, "ca-cert")
		if caCertVal != "" {
			if _, err := os.Stat(caCertVal); err != nil {
				return err
			}
		}

		inputCert, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		cert, _, err := readCertFromBuffer(inputCert)
		if err != nil {
			return err
		}

		// Prepare keys for the new certificate.
		newPrivateKey, newPublicKey, err := generateKeyPair(cert)
		if err != nil {
			return err
		}

		// Create the new certificate with all extensions and the new key.
		newCert := cert
		newCert.ExtraExtensions = newCert.Extensions
		newCert.PublicKey = newPublicKey

		var issuerCert *x509.Certificate
		var issuerKey interface{}

		if caCertVal != "" {
			issuerCert, issuerKey, err = readCertsAndKeyFromFile(caCertVal)
			if err != nil {
				return err
			}
		} else {
			// Self-sign: copy extensions and set subject to the original issuer.
			issuerCert = cert
			issuerCert.ExtraExtensions = cert.Extensions
			issuerCert.Subject = cert.Issuer
			issuerCert.RawSubject = cert.RawIssuer
			issuerKey = newPrivateKey
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, newCert, issuerCert, newPublicKey, issuerKey)
		if err != nil {
			return err
		}

		if err = pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
			return err
		}
		return pem.Encode(w, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(newPrivateKey.(*rsa.PrivateKey)),
		})
	}
	return p
}
