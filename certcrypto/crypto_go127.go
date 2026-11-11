//go:build go1.27

package certcrypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// ParsePEMPrivateKey parses a private key from key, which is a PEM block.
// Borrowed from Go standard library, to handle various private key and PEM block types.
// https://github.com/golang/go/blob/e9240c72ed38134ef108ccc61fc7e0b6ff2fd91b/src/crypto/tls/tls.go#L361-L379
// https://github.com/golang/go/blob/e9240c72ed38134ef108ccc61fc7e0b6ff2fd91b/src/crypto/tls/tls.go#L299
func ParsePEMPrivateKey(key []byte) (crypto.Signer, error) {
	keyBlockDER, _ := pem.Decode(key)
	if keyBlockDER == nil {
		return nil, errors.New("invalid PEM block")
	}

	if keyBlockDER.Type != "PRIVATE KEY" && !strings.HasSuffix(keyBlockDER.Type, " PRIVATE KEY") {
		return nil, fmt.Errorf("unknown PEM header %q", keyBlockDER.Type)
	}

	if key, err := x509.ParsePKCS1PrivateKey(keyBlockDER.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(keyBlockDER.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		case *ecdsa.PrivateKey:
			return key, nil
		case ed25519.PrivateKey:
			return key, nil
		case *mldsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping: %T", key)
		}
	}

	if key, err := x509.ParseECPrivateKey(keyBlockDER.Bytes); err == nil {
		return key, nil
	}

	return nil, errors.New("failed to parse private key")
}

func GeneratePrivateKey(keyType KeyType) (crypto.Signer, error) {
	switch keyType {
	case EC256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case EC384:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case RSA2048:
		return rsa.GenerateKey(rand.Reader, 2048)
	case RSA3072:
		return rsa.GenerateKey(rand.Reader, 3072)
	case RSA4096:
		return rsa.GenerateKey(rand.Reader, 4096)
	case RSA8192:
		return rsa.GenerateKey(rand.Reader, 8192)
	case MLDSA44:
		return mldsa.GenerateKey(mldsa.MLDSA44())
	case MLDSA65:
		return mldsa.GenerateKey(mldsa.MLDSA65())
	case MLDSA87:
		return mldsa.GenerateKey(mldsa.MLDSA87())
	}

	return nil, fmt.Errorf("invalid KeyType: %s", keyType)
}

func PEMBlock(data any) *pem.Block {
	var pemBlock *pem.Block

	switch key := data.(type) {
	case *ecdsa.PrivateKey, *rsa.PrivateKey, *mldsa.PrivateKey:
		keyBytes, _ := x509.MarshalPKCS8PrivateKey(key)
		pemBlock = &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}
	case *x509.CertificateRequest:
		pemBlock = &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: key.Raw}
	case DERCertificateBytes:
		pemBlock = &pem.Block{Type: "CERTIFICATE", Bytes: []byte(data.(DERCertificateBytes))}
	}

	return pemBlock
}
