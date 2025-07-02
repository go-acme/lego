// Package googledomains implements a DNS provider for solving the DNS-01 challenge using Google Domains DNS API.
package googledomains

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// Environment variables names.
const (
	envNamespace = "GOOGLE_DOMAINS_"

	EnvAccessToken        = envNamespace + "ACCESS_TOKEN"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessToken        string
	PollingInterval    time.Duration
	PropagationTimeout time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{}
}

type DNSProvider struct{}

// NewDNSProvider returns the Google Domains DNS provider with a default configuration.
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(&Config{})
}

// NewDNSProviderConfig returns the Google Domains DNS provider with the provided config.
func NewDNSProviderConfig(_ *Config) (*DNSProvider, error) {
	return nil, errors.New("googledomains: provider has shut down")
}

func (d *DNSProvider) Present(_, _, _ string) error {
	return nil
}

func (d *DNSProvider) CleanUp(_, _, _ string) error {
	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval
}
