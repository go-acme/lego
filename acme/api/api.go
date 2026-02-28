package api

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api/internal/nonces"
	"github.com/go-acme/lego/v5/acme/api/internal/secure"
	"github.com/go-acme/lego/v5/acme/api/internal/sender"
	"github.com/go-acme/lego/v5/log"
)

// Core ACME/LE core API.
type Core struct {
	doer         *sender.Doer
	nonceManager *nonces.Manager
	directory    acme.Directory

	HTTPClient *http.Client

	privateKey crypto.PrivateKey
	kid        string

	common         service // Reuse a single struct instead of allocating one for each service on the heap.
	Accounts       *AccountService
	Authorizations *AuthorizationService
	Certificates   *CertificateService
	Challenges     *ChallengeService
	Orders         *OrderService
}

// New Creates a new Core.
func New(httpClient *http.Client, userAgent, caDirURL, kid string, privateKey crypto.PrivateKey) (*Core, error) {
	doer := sender.NewDoer(httpClient, userAgent)

	// NOTE(ldez) add context as a parameter of the constructor?
	dir, err := getDirectory(context.Background(), doer, caDirURL)
	if err != nil {
		return nil, err
	}

	nonceManager := nonces.NewManager(doer, dir.NewNonceURL)

	c := &Core{
		doer:         doer,
		nonceManager: nonceManager,
		directory:    dir,

		privateKey: privateKey,
		kid:        kid,

		HTTPClient: httpClient,
	}

	c.common.core = c
	c.Accounts = (*AccountService)(&c.common)
	c.Authorizations = (*AuthorizationService)(&c.common)
	c.Certificates = (*CertificateService)(&c.common)
	c.Challenges = (*ChallengeService)(&c.common)
	c.Orders = (*OrderService)(&c.common)

	return c, nil
}

func (a *Core) jws() *secure.JWS {
	return secure.NewJWS(a.privateKey, a.kid, a.nonceManager)
}

// setKid Sets the key identifier (account URI).
func (a *Core) setKid(kid string) {
	if kid != "" {
		a.kid = kid
	}
}

func (a *Core) GetKid() string {
	return a.kid
}

// post performs an HTTP POST request and parses the response body as JSON,
// into the provided respBody object.
func (a *Core) post(ctx context.Context, uri string, reqBody, response any) (*http.Response, error) {
	content, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal message")
	}

	return a.retrievablePost(ctx, uri, content, response)
}

// postAsGet performs an HTTP POST ("POST-as-GET") request.
// https://www.rfc-editor.org/rfc/rfc8555.html#section-6.3
func (a *Core) postAsGet(ctx context.Context, uri string, response any) (*http.Response, error) {
	return a.retrievablePost(ctx, uri, []byte{}, response)
}

func (a *Core) retrievablePost(ctx context.Context, uri string, content []byte, response any) (*http.Response, error) {
	// during tests, allow to support ~90% of bad nonce with a minimum of attempts.
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 200 * time.Millisecond
	bo.MaxInterval = 5 * time.Second

	operation := func() (*http.Response, error) {
		resp, err := a.signedPost(ctx, uri, content, response)
		if err != nil {
			// Retry if the nonce was invalidated
			var e *acme.NonceError
			if errors.As(err, &e) {
				return resp, err
			}

			return resp, backoff.Permanent(err)
		}

		return resp, nil
	}

	notify := func(err error, duration time.Duration) {
		log.Warn("Retry.", log.ErrorAttr(err))
	}

	return backoff.Retry(ctx, operation,
		backoff.WithBackOff(bo),
		backoff.WithMaxElapsedTime(20*time.Second),
		backoff.WithNotify(notify))
}

func (a *Core) signedPost(ctx context.Context, uri string, content []byte, response any) (*http.Response, error) {
	signedContent, err := a.jws().SignContent(ctx, uri, content)
	if err != nil {
		return nil, fmt.Errorf("failed to post JWS message: failed to sign content: %w", err)
	}

	signedBody := bytes.NewBufferString(signedContent.FullSerialize())

	resp, err := a.doer.Post(ctx, uri, signedBody, "application/jose+json", response)

	// nonceErr is ignored to keep the root error.
	nonce, nonceErr := nonces.GetFromResponse(resp)
	if nonceErr == nil {
		a.nonceManager.Push(nonce)
	}

	return resp, err
}

func (a *Core) signEABContent(newAccountURL, kid string, hmac []byte) ([]byte, error) {
	eabJWS, err := a.jws().SignEABContent(newAccountURL, kid, hmac)
	if err != nil {
		return nil, err
	}

	return []byte(eabJWS.FullSerialize()), nil
}

// GetKeyAuthorization Gets the key authorization.
func (a *Core) GetKeyAuthorization(token string) (string, error) {
	return a.jws().GetKeyAuthorization(token)
}

func (a *Core) GetDirectory() acme.Directory {
	return a.directory
}

func getDirectory(ctx context.Context, do *sender.Doer, caDirURL string) (acme.Directory, error) {
	var dir acme.Directory
	if _, err := do.Get(ctx, caDirURL, &dir); err != nil {
		return dir, fmt.Errorf("get directory at '%s': %w", caDirURL, err)
	}

	if dir.NewAccountURL == "" {
		return dir, errors.New("directory missing new registration URL")
	}

	if dir.NewOrderURL == "" {
		return dir, errors.New("directory missing new order URL")
	}

	return dir, nil
}
