package lego

import (
	"errors"
	"net/url"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/resolver"
	"github.com/go-acme/lego/v5/registration"
)

// Client is the user-friendly way to ACME.
type Client struct {
	Certificate  *certificate.Certifier
	Challenge    *resolver.SolverManager
	Registration *registration.Registrar
	core         *api.Core
}

// NewClient creates a new ACME client on behalf of the user.
// The client will depend on the ACME directory located at CADirURL for the rest of its actions.
// A private key of type keyType (see KeyType constants) will be generated when requesting a new certificate if one isn't provided.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("a configuration must be provided")
	}

	_, err := url.Parse(config.CADirURL)
	if err != nil {
		return nil, err
	}

	if config.HTTPClient == nil {
		return nil, errors.New("the HTTP client cannot be nil")
	}

	privateKey := config.User.GetPrivateKey()
	if privateKey == nil {
		return nil, errors.New("private key was nil")
	}

	var kid string
	if reg := config.User.GetRegistration(); reg != nil {
		kid = reg.Location
	}

	core, err := api.New(config.HTTPClient, config.UserAgent, config.CADirURL, kid, privateKey)
	if err != nil {
		return nil, err
	}

	solversManager := resolver.NewSolversManager(core)

	prober := resolver.NewProber(solversManager)

	options := certificate.CertifierOptions{
		KeyType:             config.Certificate.KeyType,
		Timeout:             config.Certificate.Timeout,
		OverallRequestLimit: config.Certificate.OverallRequestLimit,
	}

	certifier := certificate.NewCertifier(core, prober, options)

	return &Client{
		Certificate:  certifier,
		Challenge:    solversManager,
		Registration: registration.NewRegistrar(core, config.User),
		core:         core,
	}, nil
}

// GetServerMetadata returns the current server metadata from the Directory.
func (c *Client) GetServerMetadata() acme.Meta {
	return c.core.GetDirectory().Meta
}
