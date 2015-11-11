package acme

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/http"
	"sync"

	"github.com/letsencrypt/go-jose"
)

type jws struct {
	privKey    *rsa.PrivateKey
	nonces     []string
	nonceMutex sync.Mutex
}

func keyAsJWK(key *ecdsa.PublicKey) jose.JsonWebKey {
	return jose.JsonWebKey{
		Key:       key,
		Algorithm: "EC",
	}
}

// Posts a JWS signed message to the specified URL
func (j *jws) post(url string, content []byte) (*http.Response, error) {
	err := j.getNonce(url)
	if err != nil {
		return nil, fmt.Errorf("Could not get a nonce for request: %s\n\t\tError: %v", url, err)
	}

	signedContent, err := j.signContent(content)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/jose+json", bytes.NewBuffer([]byte(signedContent.FullSerialize())))
	if err != nil {
		return nil, err
	}

	j.getNonceFromResponse(resp)

	return resp, err
}

func (j *jws) signContent(content []byte) (*jose.JsonWebSignature, error) {
	// TODO: support other algorithms - RS512
	signer, err := jose.NewSigner(jose.RS256, j.privKey)
	if err != nil {
		return nil, err
	}

	signed, err := signer.Sign(content, j.consumeNonce())
	if err != nil {
		return nil, err
	}
	return signed, nil
}

func (j *jws) getNonceFromResponse(resp *http.Response) error {
	nonce := resp.Header.Get("Replay-Nonce")
	if nonce == "" {
		return fmt.Errorf("Server did not respond with a proper nonce header.")
	}

	j.nonceMutex.Lock()
	j.nonces = append(j.nonces, nonce)
	j.nonceMutex.Unlock()
	return nil
}

func (j *jws) getNonce(url string) error {
	j.nonceMutex.Lock()
	if len(j.nonces) > 0 {
		j.nonceMutex.Unlock()
		return nil
	}
	j.nonceMutex.Unlock()

	resp, err := http.Head(url)
	if err != nil {
		return err
	}

	return j.getNonceFromResponse(resp)
}

func (j *jws) consumeNonce() string {
	j.nonceMutex.Lock()
	defer j.nonceMutex.Unlock()

	nonce := ""
	if len(j.nonces) == 0 {
		return nonce
	}

	nonce, j.nonces = j.nonces[len(j.nonces)-1], j.nonces[:len(j.nonces)-1]
	return nonce
}
