package certcrypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"math/big"
	"slices"
	"strings"
)

// Constants for all key types we support.
const (
	EC256   = KeyType("EC256")
	EC384   = KeyType("EC384")
	RSA2048 = KeyType("RSA2048")
	RSA3072 = KeyType("RSA3072")
	RSA4096 = KeyType("RSA4096")
	RSA8192 = KeyType("RSA8192")
)

// KeyType represents the key algo as well as the key size or curve to use.
type KeyType string

func (k KeyType) String() string {
	return string(k)
}

// ToKeyType gets a key type from a string.
func ToKeyType(keyType string) (KeyType, error) {
	switch strings.ToUpper(keyType) {
	case string(RSA2048):
		return RSA2048, nil
	case string(RSA3072):
		return RSA3072, nil
	case string(RSA4096):
		return RSA4096, nil
	case string(RSA8192):
		return RSA8192, nil
	case string(EC256):
		return EC256, nil
	case string(EC384):
		return EC384, nil
	}

	return "", fmt.Errorf("unsupported key type: %s", keyType)
}

func AllKeyTypes() []KeyType {
	return []KeyType{
		EC256, EC384,
		RSA2048, RSA3072, RSA4096, RSA8192,
	}
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

// GetKeyType gets the key type.
func GetKeyType(key any) (KeyType, error) {
	switch k := key.(type) {
	case *rsa.PublicKey:
		return getRSAKeyType(k.N)

	case *rsa.PrivateKey:
		return getRSAKeyType(k.N)

	case *ecdsa.PublicKey:
		return getECDSAKeyType(k.Curve)

	case *ecdsa.PrivateKey:
		return getECDSAKeyType(k.Curve)

	default:
		return "", fmt.Errorf("unsupported key type: %T", k)
	}
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
