package acme

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

	"gopkg.in/square/go-jose.v2"
)

type jws struct {
	getNonceURL string
	privKey     crypto.PrivateKey
	kid         string
	nonces      nonceManager
}

// postJSON performs an HTTP POST request and parses the response body
// as JSON, into the provided respBody object.
func (j *jws) postJSON(uri string, reqBody, respBody interface{}) (http.Header, error) {
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal network message")
	}

	signedContent, err := j.signContent(uri, jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to post JWS message -> failed to sign content -> %v", err)
	}

	data := bytes.NewBuffer([]byte(signedContent.FullSerialize()))
	resp, err := httpPost(uri, "application/jose+json", data)
	if err != nil {
		return nil, fmt.Errorf("failed to post JWS message -> failed to HTTP POST to %s -> %v", uri, err)
	}

	defer resp.Body.Close()

	nonce, nonceErr := getNonceFromResponse(resp)
	if nonceErr == nil {
		j.nonces.Push(nonce)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		err = handleHTTPError(resp)
		switch err.(type) {
		// Retry once if the nonce was invalidated
		case NonceError:
			signedContent, errP := j.signContent(uri, jsonBytes)
			if errP != nil {
				return nil, fmt.Errorf("failed to post JWS message -> failed to sign content -> %v", errP)
			}

			data := bytes.NewBuffer([]byte(signedContent.FullSerialize()))
			retryResp, errP := httpPost(uri, "application/jose+json", data)
			if errP != nil {
				return nil, fmt.Errorf("failed to post JWS message -> failed to HTTP POST to %s -> %v", uri, errP)
			}

			defer retryResp.Body.Close()

			nonce, nonceErr := getNonceFromResponse(resp)
			if nonceErr == nil {
				j.nonces.Push(nonce)
			}

			if retryResp.StatusCode >= http.StatusBadRequest {
				return retryResp.Header, handleHTTPError(retryResp)
			}

			if respBody == nil {
				return retryResp.Header, nil
			}

			return retryResp.Header, json.NewDecoder(retryResp.Body).Decode(respBody)

		default:
			return resp.Header, err
		}
	}

	if respBody == nil {
		return resp.Header, nil
	}

	return resp.Header, json.NewDecoder(resp.Body).Decode(respBody)
}

func (j *jws) signContent(url string, content []byte) (*jose.JSONWebSignature, error) {
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

func (j *jws) signEABContent(url, kid string, hmac []byte) (*jose.JSONWebSignature, error) {
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

func (j *jws) Nonce() (string, error) {
	if nonce, ok := j.nonces.Pop(); ok {
		return nonce, nil
	}

	return getNonce(j.getNonceURL)
}

func (j *jws) getKeyAuthorization(token string) (string, error) {
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
