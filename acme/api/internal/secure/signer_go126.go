//go:build !go1.27

package secure

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"

	"github.com/go-jose/go-jose/v4"
)

func signatureAlgorithm(privKey crypto.Signer) jose.SignatureAlgorithm {
	var alg jose.SignatureAlgorithm

	switch k := privKey.Public().(type) {
	case *rsa.PublicKey:
		alg = jose.RS256
	case *ecdsa.PublicKey:
		if k.Curve == elliptic.P256() {
			alg = jose.ES256
		} else if k.Curve == elliptic.P384() {
			alg = jose.ES384
		}
	}

	return alg
}
