// Package compat provides compatibility with lego/v4.
package compat

import (
	"github.com/go-acme/lego/v5/certcrypto"
)

const (
	EC256   = KeyTypeCompat(certcrypto.EC256)
	EC384   = KeyTypeCompat(certcrypto.EC384)
	RSA2048 = KeyTypeCompat(certcrypto.RSA2048)
	RSA3072 = KeyTypeCompat(certcrypto.RSA3072)
	RSA4096 = KeyTypeCompat(certcrypto.RSA4096)
	RSA8192 = KeyTypeCompat(certcrypto.RSA8192)
)

type KeyTypeCompat certcrypto.KeyType

func (k *KeyTypeCompat) UnmarshalText(text []byte) error {
	switch string(text) {
	case `P256`:
		// Compatibility with versions before lego/v5.
		*k = EC256
	case `P384`:
		// Compatibility with versions before lego/v5.
		*k = EC384
	case `2048`:
		// Compatibility with versions before lego/v5.
		*k = RSA2048
	case `3072`:
		// Compatibility with versions before lego/v5.
		*k = RSA3072
	case `4096`:
		// Compatibility with versions before lego/v5.
		*k = RSA4096
	case `8192`:
		// Compatibility with versions before lego/v5.
		*k = RSA8192
	default:
		kt, err := certcrypto.GetKeyType(string(text))
		if err != nil {
			return err
		}

		*k = KeyTypeCompat(kt)
	}

	return nil
}
