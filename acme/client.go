package acme

import (
	"errors"

	"github.com/xenolf/lego/certificate"
	"github.com/xenolf/lego/challenge/resolver"
	"github.com/xenolf/lego/le/api"
	"github.com/xenolf/lego/registration"
)

// Client is the user-friendly way to ACME
type Client struct {
	Certificate  *certificate.Certifier
	Challenge    *resolver.SolverManager
	Registration *registration.Registrar
	core         *api.Core
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

	core, err := api.New(config.HTTPClient, config.userAgent, config.caDirURL, kid, privKey)
	if err != nil {
		return nil, err
	}

	solversManager := resolver.NewSolversManager(core)

	prober := resolver.NewProber(solversManager)

	return &Client{
		Certificate:  certificate.NewCertifier(core, config.keyType, prober),
		Challenge:    solversManager,
		Registration: registration.NewRegistrar(core, config.user),
		core:         core,
	}, nil
}

// GetToSURL returns the current ToS URL from the Directory
func (c *Client) GetToSURL() string {
	return c.core.GetDirectory().Meta.TermsOfService
}

// GetExternalAccountRequired returns the External Account Binding requirement of the Directory
func (c *Client) GetExternalAccountRequired() bool {
	return c.core.GetDirectory().Meta.ExternalAccountRequired
}
