package api

import (
	"bytes"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/emca/le/api/internal/secure"
	"github.com/xenolf/lego/emca/le/api/internal/sender"
	"github.com/xenolf/lego/log"
)

type Core struct {
	do           *sender.Do
	nonceManager *secure.NonceManager
	jws          *secure.JWS
	directory    le.Directory
}

func New(httpClient *http.Client, userAgent string, caDirURL, kid string, privKey crypto.PrivateKey) (*Core, error) {
	do := sender.NewDo(httpClient, userAgent)

	dir, err := getDirectory(do, caDirURL)
	if err != nil {
		return nil, err
	}

	nonceManager := secure.NewNonceManager(do, dir.NewNonceURL)

	jws := secure.NewJWS(privKey, kid, nonceManager)

	return &Core{do: do, nonceManager: nonceManager, jws: jws, directory: dir}, nil
}

// Post performs an HTTP POST request and parses the response body as JSON,
// into the provided respBody object.
func (a *Core) Post(uri string, reqBody, response interface{}) (*http.Response, error) {
	content, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal message")
	}

	return a.retrievablePost(uri, content, response)
}

// PostAsGet performs an HTTP POST ("POST-as-GET") request.
// https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-6.3
func (a *Core) PostAsGet(uri string, response interface{}) (*http.Response, error) {
	return a.retrievablePost(uri, []byte{}, response)
}

func (a *Core) retrievablePost(uri string, content []byte, response interface{}) (*http.Response, error) {
	resp, err := a.signedPost(uri, content, response)
	if err != nil {
		switch err.(type) {
		// Retry once if the nonce was invalidated
		case *le.NonceError:
			log.Infof("nonce error retry: %s", uri)
			resp, err = a.signedPost(uri, content, response)
			if err != nil {
				return resp, err
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

func (a *Core) signedPost(uri string, content []byte, response interface{}) (*http.Response, error) {
	signedContent, err := a.jws.SignContent(uri, content)
	if err != nil {
		return nil, fmt.Errorf("failed to post JWS message -> failed to sign content -> %v", err)
	}

	signedBody := bytes.NewBuffer([]byte(signedContent.FullSerialize()))

	resp, err := a.do.Post(uri, signedBody, "application/jose+json", response)

	nonce, nonceErr := secure.GetNonceFromResponse(resp)
	if nonceErr == nil {
		a.nonceManager.Push(nonce)
	}

	return resp, err
}

// Head performs a HEAD request with a proper User-Agent string.
// The response body (resp.Body) is already closed when this function returns.
func (a *Core) Head(url string) (*http.Response, error) {
	return a.do.Head(url)
}

func (a *Core) UpdateKID(kid string) {
	a.jws.SetKid(kid)
}

func (a *Core) GetKeyAuthorization(token string) (string, error) {
	return a.jws.GetKeyAuthorization(token)
}

func (a *Core) SignEABContent(newAccountURL, kid string, hmac []byte) ([]byte, error) {
	eabJWS, err := a.jws.SignEABContent(newAccountURL, kid, hmac)
	if err != nil {
		return nil, err
	}

	return []byte(eabJWS.FullSerialize()), nil
}

func (a *Core) GetDirectory() le.Directory {
	return a.directory
}

func getDirectory(do *sender.Do, caDirURL string) (le.Directory, error) {
	var dir le.Directory
	if _, err := do.Get(caDirURL, &dir); err != nil {
		return dir, fmt.Errorf("get directory at '%s': %v", caDirURL, err)
	}

	if dir.NewAccountURL == "" {
		return dir, errors.New("directory missing new registration URL")
	}
	if dir.NewOrderURL == "" {
		return dir, errors.New("directory missing new order URL")
	}

	return dir, nil
}
