package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/xenolf/lego/storage"
)

func generatePrivateKey(path string, storage storage.StorageProvider) (crypto.PrivateKey, error) {

	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	pemKey := pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	pem_data := pem.EncodeToMemory(&pemKey)
	err = storage.WritePath(path, pem_data)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPrivateKey(path string, storage storage.StorageProvider) (crypto.PrivateKey, error) {
	keyBytes, err := storage.ReadPath(path)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("Unknown private key type.")
}
