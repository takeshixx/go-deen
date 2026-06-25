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
	"regexp"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

func printHexDumpLevel(outBuf *bytes.Buffer, data []byte, level int) {
	re := regexp.MustCompile("..")
	hex := fmt.Sprintf("%x", data)
	if len(hex)%2 == 1 {
		hex = "0" + hex
	}
	formatted := strings.TrimRight(re.ReplaceAllString(hex, "$0:"), ":")
	for j := level; j > 0; j-- {
		fmt.Fprint(outBuf, "    ")
	}

	strLength := len(formatted)
	var stop int
	for i := 0; i < strLength; i += 45 {
		stop = i + 45
		if stop > strLength {
			stop = strLength
		}
		fmt.Fprintf(outBuf, "%s\n", formatted[i:stop])
		if i+45 > strLength {
			break
		}
		for j := level; j > 0; j-- {
			fmt.Fprint(outBuf, "    ")
		}
	}
}

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
	printHexDumpLevel(outBuf, cert.SerialNumber.Bytes(), 3)
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
		printHexDumpLevel(outBuf, cert.PublicKey.(*rsa.PublicKey).N.Bytes(), 5)
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
	printHexDumpLevel(outBuf, cert.Signature, 2)

	return
}

// NewPluginCertPrinter creates a x509 certificate pretty
// printing plugin.
func NewPluginCertPrinter() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "certPrinter"
	p.Aliases = []string{"cert", "x509"}
	p.Category = "misc"
	p.Description = "x509 certificate pretty printer."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		inputCert, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var cert *x509.Certificate
		for {
			block, rest := pem.Decode(inputCert)
			if block == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				cert, err = x509.ParseCertificate(block.Bytes)
				if err != nil {
					return err
				}
				break
			}
			inputCert = rest
		}
		if cert == nil {
			return fmt.Errorf("no certificate found in input")
		}

		outBuf := new(bytes.Buffer)
		if err = prettyPrintCert(cert, outBuf); err != nil {
			return err
		}
		// Append the full PEM encoded cert.
		if err = pem.Encode(outBuf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return err
		}
		_, err = w.Write(outBuf.Bytes())
		return err
	}
	return p
}
