package misc

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/takeshixx/deen/pkg/types"
)

// readCertsAndKeyFromFile calls readCertFromBuffer() with
// bytes read from a given file.
func readCertsAndKeyFromFile(path string) (cert *x509.Certificate, key interface{}, err error) {
	data, err := ioutil.ReadFile(path)
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

// NewPluginCertCloner creates a new certificate cloner
// plugin.
func NewPluginCertCloner() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "certCloner"
	p.Aliases = []string{"clone"}
	p.Category = "misc"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		// Example processing with SHA1
		var err error
		hasher := sha1.New()
		if _, err := io.Copy(hasher, reader); err != nil {
			return nil, err
		}
		hashSum := hasher.Sum(nil)
		return hashSum, err
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		caCertFlag := flags.Lookup("ca-cert")
		caCertVal := caCertFlag.Value.String()
		if caCertVal != "" {
			if _, err := os.Stat(caCertVal); err != nil {
				return nil, err
			}
		}

		inBuf := new(bytes.Buffer)
		inBuf.ReadFrom(reader)
		inputCert := inBuf.Bytes()

		cert, _, err := readCertFromBuffer(inputCert)
		if err != nil {
			return nil, err
		}

		// Prepare keys for new certificate
		newPrivateKey, newPublicKey, err := generateKeyPair(cert)
		if err != nil {
			return nil, err
		}

		// Create the new certificate with all extensions
		// and the new public key.
		newCert := cert
		newCert.ExtraExtensions = newCert.Extensions
		newCert.PublicKey = newPublicKey

		var issuerCert *x509.Certificate
		var issuerKey interface{}

		if caCertVal != "" {
			issuerCert, issuerKey, err = readCertsAndKeyFromFile(caCertVal)
			if err != nil {
				return nil, err
			}
		} else {
			// Clone the original cert and make sure
			//	- extensions are copied
			//	- Subject set to original certs Issuer
			//  - RawIssuer set to original certs RawIssuer
			issuerCert = cert
			issuerCert.ExtraExtensions = cert.Extensions
			issuerCert.Subject = cert.Issuer
			issuerCert.RawSubject = cert.RawIssuer
			// For self-signing, we use the new certs' private key
			issuerKey = newPrivateKey
		}

		var outBytes, derBytes []byte
		outBytesBuf := bytes.NewBuffer(outBytes)

		// Create the certificate
		derBytes, err = x509.CreateCertificate(rand.Reader, newCert, issuerCert, newPublicKey, issuerKey)
		if err != nil {
			return nil, err
		}

		if err = pem.Encode(outBytesBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
			return nil, err
		}

		if err = pem.Encode(outBytesBuf, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(newPrivateKey.(*rsa.PrivateKey)),
		}); err != nil {
			return nil, err
		}

		return outBytesBuf.Bytes(), nil
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "x509 certificate cloner.\n\n")
			flags.PrintDefaults()
		}
		flags.String("ca-cert", "", "CA certificate and private key in PEM format")
		flags.Parse(args)
		return flags
	}
	return
}
