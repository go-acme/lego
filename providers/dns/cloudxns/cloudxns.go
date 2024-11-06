// Package cloudxns implements a DNS provider for solving the DNS-01 challenge using CloudXNS DNS.
package cloudxns

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
)

// Environment variables names.
const (
	envNamespace = "CLOUDXNS_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvSecretKey = envNamespace + "SECRET_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	SecretKey          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct{}

// NewDNSProvider returns a DNSProvider instance configured for CloudXNS.
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(&Config{})
}

// NewDNSProviderConfig return a DNSProvider instance configured for CloudXNS.
func NewDNSProviderConfig(_ *Config) (*DNSProvider, error) {
	return nil, errors.New("cloudxns: provider has shut down")
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(_, _, _ string) error {
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(_, _, _ string) error {
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval
}
