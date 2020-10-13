package utils

import "C"

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/spacemonkeygo/openssl"
	"github.com/takeshixx/deen/pkg/types"
)

func parseCert(certData []byte, der bool) (cert *x509.Certificate, err error) {
	var rawCert bytes.Buffer
	if !der {
		block, _ := pem.Decode(certData)
		blockReader := bytes.NewReader(block.Bytes)
		_, err = io.Copy(&rawCert, blockReader)
		if err != nil {
			return
		}
	} else {
		rawCert.Write(certData)
	}
	cert, err = x509.ParseCertificate(rawCert.Bytes())
	return
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

func cloneCert(srcCert *x509.Certificate, privKey interface{}) (dstCert []byte, err error) {
	dstCert, err = x509.CreateCertificate(nil, srcCert, srcCert, publicKey(privKey), privKey)
	if err != nil {
		return
	}

	return
}

// NewPluginX509Clone creates a X509 certificate cloner instance.
func NewPluginX509Clone() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "x509clone"
	p.Aliases = []string{"clone"}
	p.Category = "utils"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {

	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		derFlag := flags.Lookup("der")
		der, err := strconv.ParseBool(derFlag.Value.String())
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse DER flag")
			}()
			return
		}
		fmt.Printf("%v\n", der)
		certData, err := ioutil.ReadAll(task.Reader)
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to read certificate")
			}()
			return
		}

		otherCert, err := parseCert(certData, der)
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse certificate")
			}()
			return
		}
		dur := &time.Duration{}
		/* 		cert, err := parseCert(certData, der)
		   		if err != nil {
		   			go func() {
		   				defer task.Close()
		   				task.ErrChan <- errors.Wrap(err, "Failed to parse certificate")
		   			}()
		   			return
		   		}

		   		var newKey interface{}
		   		if cert.PublicKeyAlgorithm == x509.RSA {
		   			rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
		   			newKey, err = rsa.GenerateKey(rand.Reader, rsaPublicKey.N.BitLen())
		   		} else if cert.PublicKeyAlgorithm == x509.ECDSA {
		   			ecdsaPublicKey := cert.PublicKey.(*ecdsa.PublicKey)
		   			newKey, err = ecdsa.GenerateKey(ecdsaPublicKey.Curve, rand.Reader)
		   		} else if cert.PublicKeyAlgorithm == x509.Ed25519 {
		   			_, newKey, err = ed25519.GenerateKey(rand.Reader)
		   		} else {
		   			err = fmt.Errorf("Invalid public key algoritm")
		   		}
		   		if err != nil {
		   			go func() {
		   				defer task.Close()
		   				task.ErrChan <- errors.Wrap(err, "Failed to parse DER flag")
		   			}()
		   			return
		   		}

		   		cloned, err := cloneCert(cert, newKey)
		   		if err != nil {
		   			go func() {
		   				defer task.Close()
		   				task.ErrChan <- errors.Wrap(err, "Failed to clone cert")
		   			}()
		   			return
		   		}
		   		fmt.Printf("parsed cert extensions: %v\n", cert.Extensions)
		   		fmt.Printf("parsed cert extra extensions: %v\n", cert.ExtraExtensions)
		   		clonedCert, err := x509.ParseCertificate(cloned)
		   		fmt.Printf("cert extensions: %v\n", cert.Extensions == clonedCert.Extensions)
		   		fmt.Printf("cert extra extensions: %v\n", cert.ExtraExtensions)
		*/

		cert, err := openssl.LoadCertificateFromPEM(certData)
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert with openssl")
			}()
			return
		}

		newKey, err := openssl.GenerateRSAKey(2048)
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert with openssl")
			}()
			return
		}

		certSubj, err := cert.GetSubjectName()
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert with openssl")
			}()
			return
		}

		certCountry, ok := certSubj.GetEntry(openssl.NID_countryName)
		certOrg, ok := certSubj.GetEntry(openssl.NID_organizationName)
		certCN, ok := certSubj.GetEntry(openssl.NID_commonName)
		certSerial, ok := certSubj.GetEntry(openssl.NID_serialNumber)
		if !ok {
			fmt.Println("NOT OK")
		}

		converted, _ := strconv.Atoi(certSerial)
		certInfo := &openssl.CertificateInfo{
			Country:      certCountry,
			Organization: certOrg,
			CommonName:   certCN,
			Serial:       big.NewInt(int64(converted)),
		}

		clonedCert, err := openssl.NewCertificate(certInfo, newKey)
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert with openssl")
			}()
			return
		}

		certIssuer, err := cert.GetIssuerName()
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert with openssl")
			}()
			return
		}
		clonedCert.SetIssuerName(certIssuer)

		clonedCertMarshalled, err := clonedCert.MarshalPEM()
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert with openssl")
			}()
			return
		}

		clonedCertMarshalledReader := bytes.NewReader(clonedCertMarshalled)

		go func() {
			defer task.Close()
			_, err = io.Copy(task.PipeWriter, clonedCertMarshalledReader)
			if err != nil {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to copy cloned cert")
			}
		}()

		/* 		go func() {
			defer task.Close()
			clonedBuf := bytes.NewBuffer(cloned)
			_, err = io.Copy(task.PipeWriter, clonedBuf)
			if err != nil {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to copy cloned cert")
			}
		}() */
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {

	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		go func() {
			defer task.Close()

			wrappedReader := types.TrimReader{}
			wrappedReader.Rd = task.Reader
			decoder := hex.NewDecoder(wrappedReader)
			_, err := io.Copy(task.PipeWriter, decoder)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copy in Hex failed")
			}

		}()
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "X509 certificate cloner.\n\n")
			flags.PrintDefaults()
		}
		flags.Bool("der", false, "input certificate is in DER format")
		flags.Parse(args)
		return flags
	}
	return
}
