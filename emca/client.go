package emca

import (
	"errors"
	"fmt"

	"github.com/xenolf/lego/emca/certificate"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/http01"
	"github.com/xenolf/lego/emca/challenge/tlsalpn01"
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/internal/sender"
	"github.com/xenolf/lego/emca/le"
)

const (
	statusValid   = "valid"
	statusInvalid = "invalid"
)

const (
	// maxBodySize is the maximum size of body that we will read.
	maxBodySize = 1024 * 1024

	// overallRequestLimit is the overall number of request per second
	// limited on the "new-reg", "new-authz" and "new-cert" endpoints.
	// From the documentation the limitation is 20 requests per second,
	// but using 20 as value doesn't work but 18 do
	overallRequestLimit = 18
)

// Client is the user-friendly way to ACME
type Client struct {
	directory le.Directory
	user      User
	jws       *secure.JWS
	keyType   certificate.KeyType
	solvers   map[challenge.Type]solver

	do *sender.Do
}

// NewClient creates a new ACME client on behalf of the user.
// The client will depend on the ACME directory located at caDirURL for the rest of its actions.
// A private key of type keyType (see KeyType constants) will be generated when requesting a new certificate if one isn't provided.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("a configuration must be provided")
	}

	privKey := config.user.GetPrivateKey()
	if privKey == nil {
		return nil, errors.New("private key was nil")
	}

	do := sender.NewDo(config.HTTPClient, config.userAgent)

	var dir le.Directory
	if _, err := do.Get(config.caDirURL, &dir); err != nil {
		return nil, fmt.Errorf("get directory at '%s': %v", config.caDirURL, err)
	}

	if dir.NewAccountURL == "" {
		return nil, errors.New("directory missing new registration URL")
	}
	if dir.NewOrderURL == "" {
		return nil, errors.New("directory missing new order URL")
	}

	jws := secure.NewJWS(do, privKey, dir.NewNonceURL)
	if reg := config.user.GetRegistration(); reg != nil {
		jws.SetKid(reg.URI)
	}

	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME spec to this map.
	// Otherwise they won't be found.
	solvers := map[challenge.Type]solver{
		challenge.HTTP01:    http01.NewChallenge(jws, validate, &http01.ProviderServer{}),
		challenge.TLSALPN01: tlsalpn01.NewChallenge(jws, validate, &tlsalpn01.ProviderServer{}),
	}

	return &Client{
		directory: dir,
		user:      config.user,
		jws:       jws,
		keyType:   config.keyType,
		solvers:   solvers,
		do:        do,
	}, nil
}

// GetToSURL returns the current ToS URL from the Directory
func (c *Client) GetToSURL() string {
	return c.directory.Meta.TermsOfService
}

// GetExternalAccountRequired returns the External Account Binding requirement of the Directory
func (c *Client) GetExternalAccountRequired() bool {
	return c.directory.Meta.ExternalAccountRequired
}
