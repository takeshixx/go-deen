package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func PrintCert(certPEM []byte) (err error) {
	block, _ := pem.Decode(certPEM)
	rawCert := block.Bytes
	fmt.Printf("rawCert: %v\n", rawCert)
	return
}

func ParseCert(certPEM []byte) (err error) {
	block, _ := pem.Decode(certPEM)
	fmt.Printf("Cert type: %s\n", block.Type)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}
	// Issuer and Subject
	fmt.Printf("Issuer: %s\n", cert.Issuer)
	fmt.Printf("Subject: %s\n", cert.Subject)
	fmt.Printf("Public key algorithm: %s\n", cert.PublicKeyAlgorithm)
	fmt.Printf("Policy Identifiers: %v\n", cert.PolicyIdentifiers)
	fmt.Printf("SubjectKeyId: %v\n", cert.SubjectKeyId)

	var newPrivateKey, newPublicKey interface{}

	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		fmt.Printf("Creating RSA key with size: %d\n", cert.PublicKey.(*rsa.PublicKey).Size()*8)
		newPrivateKey, err = rsa.GenerateKey(rand.Reader, cert.PublicKey.(*rsa.PublicKey).Size()*8)
		if err != nil {
			return
		}
		newPublicKey = &newPrivateKey.(*rsa.PrivateKey).PublicKey
	case x509.ECDSA:
		newPrivateKey, err = ecdsa.GenerateKey(cert.PublicKey.(elliptic.Curve), rand.Reader)
		if err != nil {
			return
		}
		newPublicKey = &newPrivateKey.(*ecdsa.PrivateKey).PublicKey
	case x509.DSA:
	case x509.Ed25519:
		fmt.Printf("Public key algorithm %s not implemented yet.\n", cert.PublicKeyAlgorithm)
	}

	newCert := cert
	newCert.ExtraExtensions = newCert.Extensions
	newCert.PublicKey = newPublicKey
	fmt.Printf("newCert subject: %v\n", newCert.Subject)
	fmt.Printf("newCert issuer: %v\n", newCert.Issuer)

	issuerCert := *cert
	issuerCert.ExtraExtensions = issuerCert.Extensions
	issuerCert.Subject = cert.Issuer
	issuerCert.RawSubject = cert.RawIssuer
	fmt.Printf("CA cert subject: %v\n", issuerCert.Subject)
	fmt.Printf("CA cert issuer: %v\n", issuerCert.Issuer)

	_ = &x509.Certificate{
		SerialNumber:                cert.SerialNumber,
		SignatureAlgorithm:          cert.SignatureAlgorithm,
		PublicKeyAlgorithm:          cert.PublicKeyAlgorithm,
		PublicKey:                   newPublicKey,
		Version:                     cert.Version,
		Subject:                     cert.Subject,
		RawSubject:                  cert.RawSubject,
		Issuer:                      cert.Issuer,
		RawIssuer:                   cert.RawIssuer,
		NotBefore:                   cert.NotBefore,
		NotAfter:                    cert.NotAfter,
		SubjectKeyId:                cert.SubjectKeyId,
		ExtKeyUsage:                 cert.ExtKeyUsage,
		KeyUsage:                    cert.KeyUsage,
		Extensions:                  cert.Extensions,
		ExtraExtensions:             cert.ExtraExtensions,
		UnhandledCriticalExtensions: cert.UnhandledCriticalExtensions,
		UnknownExtKeyUsage:          cert.UnknownExtKeyUsage,
		BasicConstraintsValid:       cert.BasicConstraintsValid,
		IsCA:                        cert.IsCA,
		MaxPathLen:                  cert.MaxPathLen,
		MaxPathLenZero:              cert.MaxPathLenZero,
		OCSPServer:                  cert.OCSPServer,
		IssuingCertificateURL:       cert.IssuingCertificateURL,
		DNSNames:                    cert.DNSNames,
		EmailAddresses:              cert.EmailAddresses,
		IPAddresses:                 cert.IPAddresses,
		URIs:                        cert.URIs,
		PermittedDNSDomainsCritical: cert.PermittedDNSDomainsCritical,
		PermittedDNSDomains:         cert.PermittedDNSDomains,
		ExcludedDNSDomains:          cert.ExcludedDNSDomains,
		PermittedIPRanges:           cert.PermittedIPRanges,
		PermittedEmailAddresses:     cert.PermittedEmailAddresses,
		ExcludedEmailAddresses:      cert.ExcludedEmailAddresses,
		PermittedURIDomains:         cert.PermittedURIDomains,
		ExcludedURIDomains:          cert.ExcludedURIDomains,
		CRLDistributionPoints:       cert.CRLDistributionPoints,
		PolicyIdentifiers:           cert.PolicyIdentifiers,
	}
	//fmt.Printf("New cert Subject: %s\n", newCert.Subject)
	//fmt.Printf("New private key: %v\n", newPrivateKey)

	fmt.Println("Creating certificate")

	derBytes, err := x509.CreateCertificate(rand.Reader, newCert, &issuerCert, newPublicKey, newPrivateKey)
	if err != nil {
		return
	}

	fmt.Println("Creating output file")

	certOut, err := os.Create("/tmp/certs/newCert2")
	if err != nil {
		return
	}

	fmt.Println("Writing newCert to file")

	if err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return
	}

	if err = certOut.Close(); err != nil {
		return
	}

	return
}

/* func SaveCertToPath(x509.Certificate cert, string path) (err error) {

	return
} */

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [Certificate]\n", os.Args[0])
		return
	}

	certName := os.Args[1]
	if _, err := os.Stat(certName); err != nil {
		fmt.Printf("Certificate %s not found\n", certName)
		return
	}

	certData, err := ioutil.ReadFile(certName)
	if err != nil {
		fmt.Printf("Failed to read certificate: %s\n", err)
		return
	}

	//fmt.Printf("Read certificate data:\n\n%s\n", certData)

	//PrintCert([]byte(certData))
	err = ParseCert([]byte(certData))
	if err != nil {
		log.Fatal(err)
	}
}
