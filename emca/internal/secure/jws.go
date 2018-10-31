package secure

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/xenolf/lego/emca/internal/sender"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
	"gopkg.in/square/go-jose.v2"
)

type JWS struct {
	do          *sender.Do
	privKey     crypto.PrivateKey
	getNonceURL string
	kid         string // Key identifier
	nonces      nonceManager
}

func NewJWS(do *sender.Do, privKey crypto.PrivateKey, nonceURL string) *JWS {
	return &JWS{
		do:          do,
		privKey:     privKey,
		getNonceURL: nonceURL,
	}
}

func (j *JWS) SetKid(kid string) {
	j.kid = kid
}

// PostJSON performs an HTTP POST request and parses the response body as JSON,
// into the provided respBody object.
func (j *JWS) PostJSON(uri string, reqBody, response interface{}) (http.Header, error) {
	content, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal message")
	}

	resp, err := j.retrievablePost(uri, content, response)
	if err != nil {
		return nil, err
	}

	return resp.Header, nil
}

// PostAsGet performs an HTTP POST ("POST-as-GET") request.
func (j *JWS) PostAsGet(uri string, response interface{}) (*http.Response, error) {
	return j.retrievablePost(uri, []byte{}, response)
}

func (j *JWS) retrievablePost(uri string, content []byte, response interface{}) (*http.Response, error) {
	resp, err := j.signedPost(uri, content, response)
	if err != nil {
		switch err.(type) {
		// Retry once if the nonce was invalidated
		case *le.NonceError:
			log.Infof("nonce error retry: %s", uri)
			resp, err = j.signedPost(uri, content, response)
			if err != nil {
				return resp, err
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

func (j *JWS) signedPost(uri string, content []byte, response interface{}) (*http.Response, error) {
	signedContent, err := j.signContent(uri, content)
	if err != nil {
		return nil, fmt.Errorf("failed to post JWS message -> failed to sign content -> %v", err)
	}

	signedBody := bytes.NewBuffer([]byte(signedContent.FullSerialize()))

	resp, err := j.do.Post(uri, signedBody, "application/jose+json", response)

	nonce, nonceErr := getNonceFromResponse(resp)
	if nonceErr == nil {
		j.nonces.Push(nonce)
	}

	return resp, err
}

func (j *JWS) signContent(url string, content []byte) (*jose.JSONWebSignature, error) {
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

	signKey := jose.SigningKey{
		Algorithm: alg,
		Key:       jose.JSONWebKey{Key: j.privKey, KeyID: j.kid},
	}

	options := jose.SignerOptions{
		NonceSource: j,
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"url": url,
		},
	}

	if j.kid == "" {
		options.EmbedJWK = true
	}

	signer, err := jose.NewSigner(signKey, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to create jose signer -> %v", err)
	}

	signed, err := signer.Sign(content)
	if err != nil {
		return nil, fmt.Errorf("failed to sign content -> %v", err)
	}
	return signed, nil
}

func (j *JWS) SignEABContent(url, kid string, hmac []byte) (*jose.JSONWebSignature, error) {
	jwk := jose.JSONWebKey{Key: j.privKey}
	jwkJSON, err := jwk.Public().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("acme: error encoding eab jwk key: %v", err)
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: hmac},
		&jose.SignerOptions{
			EmbedJWK: false,
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				"kid": kid,
				"url": url,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create External Account Binding jose signer -> %v", err)
	}

	signed, err := signer.Sign(jwkJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to External Account Binding sign content -> %v", err)
	}

	return signed, nil
}

func (j *JWS) Nonce() (string, error) {
	if nonce, ok := j.nonces.Pop(); ok {
		return nonce, nil
	}

	return j.getNonce()
}

func (j *JWS) getNonce() (string, error) {
	resp, err := j.do.Head(j.getNonceURL)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce from HTTP HEAD -> %v", err)
	}

	return getNonceFromResponse(resp)
}

func (j *JWS) GetKeyAuthorization(token string) (string, error) {
	var publicKey crypto.PublicKey
	switch k := j.privKey.(type) {
	case *ecdsa.PrivateKey:
		publicKey = k.Public()
	case *rsa.PrivateKey:
		publicKey = k.Public()
	}

	// Generate the Key Authorization for the challenge
	jwk := &jose.JSONWebKey{Key: publicKey}
	if jwk == nil {
		return "", errors.New("could not generate JWK from key")
	}

	thumbBytes, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}

	// unpad the base64URL
	keyThumb := base64.RawURLEncoding.EncodeToString(thumbBytes)

	return token + "." + keyThumb, nil
}
