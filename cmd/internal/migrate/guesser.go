package migrate

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/go-acme/lego/v5/certcrypto"
)

func guessPrivateKeyType(key crypto.Signer) (certcrypto.KeyType, error) {
	return guessPublicKeyType(key.Public())
}

func guessCertificateKeyType(cert *x509.Certificate) (certcrypto.KeyType, error) {
	return guessPublicKeyType(cert.PublicKey)
}

func guessPublicKeyType(pubKey any) (certcrypto.KeyType, error) {
	switch pub := pubKey.(type) {
	case *rsa.PublicKey:
		switch pub.N.BitLen() {
		case 2048:
			return certcrypto.RSA2048, nil
		case 3072:
			return certcrypto.RSA3072, nil
		case 4096:
			return certcrypto.RSA4096, nil
		case 8192:
			return certcrypto.RSA8192, nil
		default:
			return "", fmt.Errorf("unsupported RSA key: %d", pub.N.BitLen())
		}

	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P256():
			return certcrypto.EC256, nil
		case elliptic.P384():
			return certcrypto.EC384, nil
		default:
			return "", fmt.Errorf("unsupported ECDSA curve: %d", pub.Curve.Params().BitSize)
		}

	default:
		return "", fmt.Errorf("unsupported key type: %T", pub)
	}
}
