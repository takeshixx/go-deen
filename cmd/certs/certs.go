package main

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

func PrintCert(certPEM []byte) (err error) {
	block, _ := pem.Decode(certPEM)
	rawCert := block.Bytes
	fmt.Printf("rawCert: %v\n", rawCert)
	return
}

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

	fmt.Printf("Read certificate data:\n\n%s\n", certData)

	PrintCert([]byte(certData))
}
