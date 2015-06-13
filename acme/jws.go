package acme

import (
	"bytes"
	"crypto/rsa"
	"net/http"

	"github.com/square/go-jose"
)

type jws struct {
	privKey *rsa.PrivateKey
}

// Posts a JWS signed message to the specified URL
func (j *jws) post(url string, content []byte) (*http.Response, error) {
	// TODO: support other algorithms - RS512
	signer, err := jose.NewSigner(jose.RS256, j.privKey)
	if err != nil {
		return nil, err
	}

	signed, err := signer.Sign(content)
	if err != nil {
		return nil, err
	}
	signedContent := signed.FullSerialize()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(signedContent)))
	if err != nil {
		return nil, err
	}

	return resp, err
}
