// Package autodns implements a DNS provider for solving the DNS-01 challenge using auto DNS.
package autodns

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "AUTODNS_"

	EnvAPIUser            = envNamespace + "API_USER"
	EnvAPIPassword        = envNamespace + "API_PASSWORD"
	EnvAPIEndpoint        = envNamespace + "ENDPOINT"
	EnvAPIEndpointContext = envNamespace + "CONTEXT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	defaultEndpointContext int = 4
	defaultTTL             int = 600
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Endpoint           *url.URL
	Username           string
	Password           string
	Context            int
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig(conf map[string]string) *Config {
	endpoint, _ := url.Parse(env.GetOrDefaultString(conf, EnvAPIEndpoint, defaultEndpoint))

	return &Config{
		Endpoint:           endpoint,
		Context:            env.GetOrDefaultInt(conf, EnvAPIEndpointContext, defaultEndpointContext),
		TTL:                env.GetOrDefaultInt(conf, EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(conf, EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(conf, EnvPollingInterval, 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(conf, EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for autoDNS.
// Credentials must be passed in the environment variables.
func NewDNSProvider(conf map[string]string) (*DNSProvider, error) {
	values, err := env.Get(conf, EnvAPIUser, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("autodns: %w", err)
	}

	config := NewDefaultConfig(conf)
	config.Username = values[EnvAPIUser]
	config.Password = values[EnvAPIPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for autoDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("autodns: config is nil")
	}

	if config.Username == "" {
		return nil, errors.New("autodns: missing user")
	}

	if config.Password == "" {
		return nil, errors.New("autodns: missing password")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	records := []*ResourceRecord{{
		Name:  fqdn,
		TTL:   int64(d.config.TTL),
		Type:  "TXT",
		Value: value,
	}}

	_, err := d.addTxtRecord(domain, records)
	if err != nil {
		return fmt.Errorf("autodns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	records := []*ResourceRecord{{
		Name:  fqdn,
		TTL:   int64(d.config.TTL),
		Type:  "TXT",
		Value: value,
	}}

	if err := d.removeTXTRecord(domain, records); err != nil {
		return fmt.Errorf("autodns: %w", err)
	}

	return nil
}
