package emca

import (
	"errors"
	"fmt"

	"github.com/xenolf/lego/emca/certificate"
	"github.com/xenolf/lego/emca/challenge/resolver"
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/internal/sender"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/emca/registration"
)

// Client is the user-friendly way to ACME
type Client struct {
	Certificate  *certificate.Certifier
	Challenge    *resolver.SolverManager
	Registration *registration.Registrar
	directory    le.Directory
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

	var kid string
	if reg := config.user.GetRegistration(); reg != nil {
		kid = reg.URI
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
	jws.SetKid(kid)

	solversManager := resolver.NewSolversManager(jws)
	prober := resolver.NewProber(jws, solversManager)

	return &Client{
		Certificate:  certificate.NewCertifier(jws, config.keyType, dir, prober),
		Challenge:    solversManager,
		Registration: registration.NewRegistrar(jws, config.user, dir),
		directory:    dir,
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
