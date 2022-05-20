package webnames

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "WEBNAMES_"

	EnvApiKey = envNamespace + "APIKEY" // FIXME

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string // FIXME

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	// FIXME client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for reg.ru.
// Credentials must be passed in the environment variable: WEBNAMES_APIKEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvApiKey) // FIXME
	if err != nil {
		return nil, fmt.Errorf("webnames: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvApiKey] // FIXME

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for reg.ru.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("webnames: the configuration of the DNS provider is nil")
	}
	// FIXME
	panic("not yet implemented")
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	// FIXME
	// fqdn, value := dns01.GetRecord(domain, keyAuth)
	//
	// authZone, err := dns01.FindZoneByFqdn(fqdn)
	// if err != nil {
	// 	return fmt.Errorf("webnames: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	// }
	// subDomain := dns01.UnFqdn(strings.TrimSuffix(fqdn, authZone))

	panic("not yet implemented")
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	// FIXME
	// fqdn, value := dns01.GetRecord(domain, keyAuth)
	//
	// authZone, err := dns01.FindZoneByFqdn(fqdn)
	// if err != nil {
	// 	return fmt.Errorf("webnames: could not find zone for domain %q and fqdn %q : %w", domain, fqdn, err)
	// }
	// subDomain := dns01.UnFqdn(strings.TrimSuffix(fqdn, authZone))

	panic("not yet implemented")
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
