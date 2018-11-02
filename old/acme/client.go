// Package acme implements the ACME protocol for Let's Encrypt and other conforming providers.
package acme

import (
	"crypto"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/xenolf/lego/log"
)

const (
	// maxBodySize is the maximum size of body that we will read.
	maxBodySize = 1024 * 1024

	// overallRequestLimit is the overall number of request per second
	// limited on the "new-reg", "new-authz" and "new-cert" endpoints.
	// From the documentation the limitation is 20 requests per second,
	// but using 20 as value doesn't work but 18 do
	overallRequestLimit = 18

	statusValid   = "valid"
	statusInvalid = "invalid"
)

// User interface is to be implemented by users of this library.
// It is used by the client type to get user specific information.
type User interface {
	GetEmail() string
	GetRegistration() *RegistrationResource
	GetPrivateKey() crypto.PrivateKey
}

type validateFunc func(j *jws, domain, uri string, chlng challenge) error

// Client is the user-friendly way to ACME
type Client struct {
	directory directory
	user      User
	jws       *jws
	keyType   KeyType
	solvers   map[Challenge]solver
}

// NewClient creates a new ACME client on behalf of the user.
// The client will depend on the ACME directory located at caDirURL for the rest of its actions.
// A private key of type keyType (see KeyType constants) will be generated
// when requesting a new certificate if one isn't provided.
func NewClient(caDirURL string, user User, keyType KeyType) (*Client, error) {
	privKey := user.GetPrivateKey()
	if privKey == nil {
		return nil, errors.New("private key was nil")
	}

	var dir directory
	if _, err := getJSON(caDirURL, &dir); err != nil {
		return nil, fmt.Errorf("get directory at '%s': %v", caDirURL, err)
	}

	if dir.NewAccountURL == "" {
		return nil, errors.New("directory missing new registration URL")
	}
	if dir.NewOrderURL == "" {
		return nil, errors.New("directory missing new order URL")
	}

	jws := newJWS(privKey, dir.NewNonceURL)
	if reg := user.GetRegistration(); reg != nil {
		jws.setKid(reg.URI)
	}

	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME spec to this map.
	// Otherwise they won`t be found.
	solvers := map[Challenge]solver{
		HTTP01:    &httpChallenge{jws: jws, validate: validate, provider: &HTTPProviderServer{}},
		TLSALPN01: &tlsALPNChallenge{jws: jws, validate: validate, provider: &TLSALPNProviderServer{}},
	}

	return &Client{directory: dir, user: user, jws: jws, keyType: keyType, solvers: solvers}, nil
}

// GetToSURL returns the current ToS URL from the Directory
func (c *Client) GetToSURL() string {
	return c.directory.Meta.TermsOfService
}

// GetExternalAccountRequired returns the External Account Binding requirement of the Directory
func (c *Client) GetExternalAccountRequired() bool {
	return c.directory.Meta.ExternalAccountRequired
}

// validate makes the ACME server start validating a
// challenge response, only returning once it is done.
func validate(j *jws, domain, uri string, c challenge) error {
	var chlng challenge

	// Challenge initiation is done by sending a JWS payload containing the
	// trivial JSON object `{}`. We use an empty struct instance as the postJSON
	// payload here to achieve this result.
	hdr, err := j.postJSON(uri, struct{}{}, &chlng)
	if err != nil {
		return err
	}

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	for {
		switch chlng.Status {
		case statusValid:
			log.Infof("[%s] The server validated our request", domain)
			return nil
		case "pending":
		case "processing":
		case statusInvalid:
			return handleChallengeError(chlng)
		default:
			return errors.New("the server returned an unexpected state")
		}

		ra, err := strconv.Atoi(hdr.Get("Retry-After"))
		if err != nil {
			// The ACME server MUST return a Retry-After.
			// If it doesn't, we'll just poll hard.
			ra = 5
		}
		time.Sleep(time.Duration(ra) * time.Second)

		hdr, err = getJSON(uri, &chlng)
		if err != nil {
			return err
		}
	}
}
