package acme

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"net/http"
	"sync"

	jose "gopkg.in/square/go-jose.v1"
)

type jws struct {
	directoryURL string
	privKey      crypto.PrivateKey

	mu     sync.Mutex
	nonces []string
}

func keyAsJWK(key interface{}) *jose.JsonWebKey {
	switch k := key.(type) {
	case *ecdsa.PublicKey:
		return &jose.JsonWebKey{Key: k, Algorithm: "EC"}
	case *rsa.PublicKey:
		return &jose.JsonWebKey{Key: k, Algorithm: "RSA"}

	default:
		return nil
	}
}

// Posts a JWS signed message to the specified URL
func (j *jws) post(url string, content []byte) (*http.Response, error) {
	signedContent, err := j.signContent(content)
	if err != nil {
		return nil, err
	}

	resp, err := httpPost(url, "application/jose+json", bytes.NewBuffer([]byte(signedContent.FullSerialize())))
	if err != nil {
		return nil, err
	}

	nonce, err := j.getNonceFromResponse(resp)
	if err != nil {
		return nil, err
	}
	j.addNonce(nonce)
	return resp, nil
}

func (j *jws) signContent(content []byte) (*jose.JsonWebSignature, error) {

	var alg jose.SignatureAlgorithm
	switch k := j.privKey.(type) {
	case *rsa.PrivateKey:
		alg = jose.RS256
	case *ecdsa.PrivateKey:
		if k.Curve == elliptic.P256() {
			alg = jose.ES256
		} else if k.Curve == elliptic.P384() {
			alg = jose.ES384
		}
	}

	signer, err := jose.NewSigner(alg, j.privKey)
	if err != nil {
		return nil, err
	}
	signer.SetNonceSource(j)

	signed, err := signer.Sign(content)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

func (j *jws) getNonceFromResponse(resp *http.Response) (string, error) {
	nonce := resp.Header.Get("Replay-Nonce")
	if nonce == "" {
		return "", fmt.Errorf("Server did not respond with a proper nonce header.")
	}
	return nonce, nil
}

func (j *jws) Nonce() (string, error) {
	nonce := j.claimNonce()
	if nonce != "" {
		return nonce, nil
	}
	resp, err := httpHead(j.directoryURL)
	if err != nil {
		return "", err
	}

	return j.getNonceFromResponse(resp)
}

func (j *jws) claimNonce() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	if len(j.nonces) == 0 {
		return ""
	}
	nonce := ""
	for len(j.nonces) != 0 && nonce == "" {
		nonce, j.nonces = j.nonces[len(j.nonces)-1], j.nonces[:len(j.nonces)-1]
	}
	return nonce
}

func (j *jws) addNonce(nonce string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.nonces = append(j.nonces, nonce)
}
