// Package rackspace implements a DNS provider for solving the DNS-01 challenge using rackspace DNS.
package rackspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

// defaultBaseURL represents the Identity API endpoint to call.
const defaultBaseURL = "https://identity.api.rackspacecloud.com/v2.0/tokens"

// Environment variables names.
const (
	envNamespace = "RACKSPACE_"

	EnvUser   = envNamespace + "USER"
	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIUser            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config           *Config
	token            string
	cloudDNSEndpoint string
}

// NewDNSProvider returns a DNSProvider instance configured for Rackspace.
// Credentials must be passed in the environment variables:
// RACKSPACE_USER and RACKSPACE_API_KEY.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvUser, EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.APIUser = values[EnvUser]
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Rackspace.
// It authenticates against the API, also grabbing the DNS Endpoint.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rackspace: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, errors.New("rackspace: credentials missing")
	}

	identity, err := login(config)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %w", err)
	}

	// Iterate through the Service Catalog to get the DNS Endpoint
	var dnsEndpoint string
	for _, service := range identity.Access.ServiceCatalog {
		if service.Name == "cloudDNS" {
			dnsEndpoint = service.Endpoints[0].PublicURL
			break
		}
	}

	if dnsEndpoint == "" {
		return nil, errors.New("rackspace: failed to populate DNS endpoint, check Rackspace API for changes")
	}

	return &DNSProvider{
		config:           config,
		token:            identity.Access.Token.ID,
		cloudDNSEndpoint: dnsEndpoint,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	rec := Records{
		Record: []Record{{
			Name: dns01.UnFqdn(fqdn),
			Type: "TXT",
			Data: value,
			TTL:  d.config.TTL,
		}},
	}

	body, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	_, err = d.makeRequest(http.MethodPost, fmt.Sprintf("/domains/%d/records", zoneID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	record, err := d.findTxtRecord(fqdn, zoneID)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}

	_, err = d.makeRequest(http.MethodDelete, fmt.Sprintf("/domains/%d/records?id=%s", zoneID, record.ID), nil)
	if err != nil {
		return fmt.Errorf("rackspace: %w", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
