package misc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/takeshixx/deen/pkg/types"
)

// levelPrinter is a print wrapper that allows to
// specify indentation levels.
func levelPrinter(outBuf *bytes.Buffer, level int, format string, args ...interface{}) {
	for i := level; i > 0; i-- {
		fmt.Fprint(outBuf, "    ")
	}
	fmt.Fprintf(outBuf, format, args...)
	fmt.Fprint(outBuf, "\n")
}

// prettyPrintCert writes the certificate data to outBuf.
func prettyPrintCert(cert *x509.Certificate, outBuf *bytes.Buffer) (err error) {
	levelPrinter(outBuf, 0, "Certificate:")
	levelPrinter(outBuf, 1, "Data:")
	levelPrinter(outBuf, 2, "Version: %d (%x)", cert.Version, cert.Version)
	levelPrinter(outBuf, 2, "Serial Number:")
	levelPrinter(outBuf, 3, "%s", cert.SerialNumber)
	levelPrinter(outBuf, 1, "Signature Algorithm: %s", cert.SignatureAlgorithm)
	levelPrinter(outBuf, 2, "Issuer: %s", cert.Issuer)
	levelPrinter(outBuf, 2, "Validity:")
	levelPrinter(outBuf, 3, "Not Before: %s", cert.NotBefore)
	levelPrinter(outBuf, 3, "Not After: %s", cert.NotAfter)
	levelPrinter(outBuf, 2, "Subject: %s", cert.Subject)
	levelPrinter(outBuf, 2, "Subject Public Key Info:")
	levelPrinter(outBuf, 3, "Public Key Algorithm: %s", cert.PublicKeyAlgorithm)

	var bitLen int
	switch privKey := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		bitLen = privKey.N.BitLen()
		levelPrinter(outBuf, 4, "Public-Key: (%d)", bitLen)
		levelPrinter(outBuf, 4, "Modulus:")
		levelPrinter(outBuf, 5, "%02X", cert.PublicKey.(*rsa.PublicKey).N)
		levelPrinter(outBuf, 4, "Exponent: %d (%4x)", cert.PublicKey.(*rsa.PublicKey).E, cert.PublicKey.(*rsa.PublicKey).E)
	case *ecdsa.PublicKey:
		bitLen = privKey.Curve.Params().BitSize
	case *ed25519.PublicKey:
		bitLen = 32 * 8
	default:
		log.Fatal("unsupported private key")
	}

	levelPrinter(outBuf, 2, "X509v3 extensions:")
	var critical string
	for _, ext := range cert.Extensions {
		if ext.Critical {
			critical = "critical"
		} else {
			critical = ""
		}
		levelPrinter(outBuf, 3, "%s: %s", ext.Id.String(), critical)
		levelPrinter(outBuf, 4, "%s", ext.Value)
	}

	levelPrinter(outBuf, 1, "Signature Algorithm: %s", cert.SignatureAlgorithm)
	levelPrinter(outBuf, 2, "%02X", cert.Signature)

	return
}

// NewPluginCertPrinter creates a x509 certificate pretty
// printing plugin.
func NewPluginCertPrinter() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "certPrinter"
	p.Aliases = []string{"cert", "x509"}
	p.Category = "misc"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return p.ProcessStreamWithCliFlagsFunc(nil, reader)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		inBuf := new(bytes.Buffer)
		inBuf.ReadFrom(reader)
		inputCert := inBuf.Bytes()
		var cert *x509.Certificate
		var err error

		for {
			block, rest := pem.Decode(inputCert)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				cert, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, err
				}
				break
			}
			inputCert = rest
		}

		var outBytes []byte
		outBytesBuf := bytes.NewBuffer(outBytes)

		err = prettyPrintCert(cert, outBytesBuf)
		if err != nil {
			return nil, err
		}

		// Print the full PEM encoded cert at the end
		err = pem.Encode(outBytesBuf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err != nil {
			return nil, err
		}

		return outBytesBuf.Bytes(), nil
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "x509 certificate pretty printer.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
