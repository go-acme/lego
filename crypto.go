package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
)

func generateRsaKey(length int, file string) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, length)
	if err != nil {
		return nil, err
	}

	pemKey := pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

	certOut, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	pem.Encode(certOut, &pemKey)
	certOut.Close()

	return privateKey, nil
}

func loadRsaKey(file string) (*rsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)
	return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
}
