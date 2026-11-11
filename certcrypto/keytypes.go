package certcrypto

import (
	"crypto"
	"crypto/elliptic"
	"crypto/x509"
	"fmt"
	"math/big"
	"slices"
)

// KeyType represents the key algo as well as the key size or curve to use.
type KeyType string

func (k KeyType) String() string {
	return string(k)
}

// IsSupported checks if the key type is supported.
func IsSupported(keyType KeyType) bool {
	return slices.Contains(AllKeyTypes(), keyType)
}

// GetPrivateKeyType gets the key type based on the public key from crypto.Signer.
func GetPrivateKeyType(signer crypto.Signer) (KeyType, error) {
	return GetKeyType(signer.Public())
}

// GetCertificateKeyType gets the key type based on the public key from x509.Certificate.
func GetCertificateKeyType(cert *x509.Certificate) (KeyType, error) {
	return GetKeyType(cert.PublicKey)
}

// GetCSRKeyType gets the key type based on the public key from x509.CertificateRequest.
func GetCSRKeyType(csr *x509.CertificateRequest) (KeyType, error) {
	return GetKeyType(csr.PublicKey)
}

func getRSAKeyType(n *big.Int) (KeyType, error) {
	switch n.BitLen() {
	case 2048:
		return RSA2048, nil
	case 3072:
		return RSA3072, nil
	case 4096:
		return RSA4096, nil
	case 8192:
		return RSA8192, nil
	default:
		return "", fmt.Errorf("unsupported RSA key: %d", n.BitLen())
	}
}

func getECDSAKeyType(curve elliptic.Curve) (KeyType, error) {
	switch curve {
	case elliptic.P256():
		return EC256, nil
	case elliptic.P384():
		return EC384, nil
	default:
		return "", fmt.Errorf("unsupported ECDSA curve: %d", curve.Params().BitSize)
	}
}
