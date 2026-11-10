// Package versio implements a DNS provider for solving the DNS-01 challenge using versio DNS.
package versio

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
)

// Environment variables names.
const (
	envNamespace = "VERSIO_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"
	EnvEndpoint = envNamespace + "ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            *url.URL
	TTL                int
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct{}

// NewDNSProvider returns a DNSProvider instance.
//
// Deprecated: The Versio DNS provider shut down in 2024.
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(&Config{})
}

// NewDNSProviderConfig return a DNSProvider instance configured for Versio.
//
// Deprecated: The Versio DNS provider shut down in 2024.
func NewDNSProviderConfig(_ *Config) (*DNSProvider, error) {
	return nil, errors.New("stackpath: provider has shut down")
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	return nil
}
