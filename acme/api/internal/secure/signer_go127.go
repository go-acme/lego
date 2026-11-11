//go:build go1.27

package secure

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rsa"

	"github.com/go-jose/go-jose/v4"
)

func signatureAlgorithm(privKey crypto.Signer) jose.SignatureAlgorithm {
	var alg jose.SignatureAlgorithm

	switch k := privKey.Public().(type) {
	case *rsa.PublicKey:
		alg = jose.RS256
	case *ecdsa.PublicKey:
		switch k.Curve {
		case elliptic.P256():
			alg = jose.ES256
		case elliptic.P384():
			alg = jose.ES384
		case elliptic.P521():
			alg = jose.ES512
		}
	case *mldsa.PublicKey:
		switch k.Parameters() {
		case mldsa.MLDSA44():
			alg = jose.MLDSA44
		case mldsa.MLDSA65():
			alg = jose.MLDSA65
		case mldsa.MLDSA87():
			alg = jose.MLDSA87
		}
	}

	return alg
}
