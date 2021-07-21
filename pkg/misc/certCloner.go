package misc

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/pkg/errors"
	"github.com/spacemonkeygo/openssl"
	"github.com/takeshixx/deen/pkg/types"
)

// #include "shim.h"
import "C"

type Certificate struct {
	openssl.Certificate
}

func (c *Certificate) getIssuerDate() (ret interface{}, err error) {
	ret = C.X_X509_get0_notBefore(c.x)
	return
}

func parseCertificate(path string) (cert *x509.Certificate, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	block, _ := pem.Decode(data)
	if err != nil {
		return
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	return
}

// NewPluginStreamExample creates a stream-based plugin that can
// be used for plugins that do not implement readers/writers.
// This often applies to plugins that have fixed-size outputs
// like hashs, that return byte arrays instead of writing to
// writers directly. These types of plugins might also not
// have a unprocessing functions because they implement a
// one-way process.
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
		caFlag := flags.Lookup("ca")
		caVal := caFlag.Value.String()
		if caVal != "" {
			fmt.Printf("Got CA cert: %s\n", caVal)
		}

		inBuf := new(bytes.Buffer)
		inBuf.ReadFrom(reader)
		inputCert := inBuf.Bytes()

		cert, err := openssl.LoadCertificateFromPEM(inputCert)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Parsed certificate: %v\n", cert)

		serial := new(big.Int)
		serial, ok := serial.SetString(cert.GetSerialNumberHex(), 16)
		if !ok {
			return nil, errors.New("Failed to parse certificate serial")
		}
		subjectName, err := cert.GetSubjectName()
		if err != nil {
			return nil, err
		}
		commonName, ok := subjectName.GetEntry(openssl.NID_commonName)
		if !ok {
			return nil, errors.New("Failed to get CN from certificate")
		}
		org, ok := subjectName.GetEntry(openssl.NID_organizationName)
		if !ok {
			return nil, errors.New("Failed to get org from certificate")
		}
		country, ok := subjectName.GetEntry(openssl.NID_countryName)
		if !ok {
			return nil, errors.New("Failed to get country from certificate")
		}

		certInfo := &openssl.CertificateInfo{
			Serial:       serial,
			CommonName:   commonName,
			Organization: org,
			Country:      country,
		}
		newCert := openssl.NewCertificate()

		return nil, errors.New("Default error")
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			// Add a description for the plugin.
			fmt.Fprintf(os.Stderr, "Plugin description...\n\n")
			// Additional arguments will be listed automatically
			// in the help page, they should not be mentioned in
			// the above description.
			flags.PrintDefaults()
		}
		// Adding additional flags:
		flags.String("ca", "", "CA certificate and key in PEM format")

		// Different options for processing and unprocessing
		// can be added by checking:
		//if self.Unprocess {
		//
		//}

		flags.Parse(args)
		return flags
	}
	return
}
