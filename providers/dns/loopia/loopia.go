// Package loopia implements a DNS provider for solving the DNS-01 challenge using loopia DNS.
package loopia

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "LOOPIA_"

	EnvAPIUser     = envNamespace + "API_USER"
	EnvAPIPassword = envNamespace + "API_PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL            string
	APIUser            string
	APIPassword        string
	Client             *dnsClient
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 40*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 60*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 60*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config         *Config
	client         dnsClient
	findZoneByFqdn func(fqdn string) (string, error)
	inProgressInfo map[string]int
}

// NewDNSProvider returns a DNSProvider instance configured for Loopia.
// Credentials must be passed in the environment variables LOOPIA_API_USER and LOOPIA_API_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("loopia: %w", err)
	}
	config := NewDefaultConfig()

	config.APIUser = values[EnvAPIUser]
	config.APIPassword = values[EnvAPIPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Loopia.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("loopia: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIPassword == "" {
		return nil, errors.New("loopia: credentials missing")
	}
	// Min value for TTL is 300
	if config.TTL < 300 {
		config.TTL = 300
	}

	client := NewClient(config.APIUser, config.APIPassword)
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{
		config:         config,
		client:         &client,
		findZoneByFqdn: dns01.FindZoneByFqdn,
		inProgressInfo: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	subdomain, authZone := d.splitDomain(fqdn)
	err := d.client.addTXTRecord(authZone, subdomain, d.config.TTL, value)
	if err != nil {
		return err
	}
	txtRecords, err := d.client.getTXTRecords(authZone, subdomain)
	if err != nil {
		return err
	}
	for _, r := range txtRecords {
		if r.Rdata == value {
			d.inProgressInfo[token] = r.RecordID
			return nil
		}
	}
	return fmt.Errorf("loopia: Failed to get id for TXT record")
}

// CleanUp removes the TXT record matching the specified
// parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	subdomain, authZone := d.splitDomain(fqdn)
	err := d.client.removeTXTRecord(authZone, subdomain, d.inProgressInfo[token])
	if err != nil {
		return err
	}
	records, err := d.client.getTXTRecords(authZone, subdomain)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		err = d.client.removeSubdomain(authZone, subdomain)
	}
	return err
}

func (d *DNSProvider) splitDomain(fqdn string) (string, string) {
	authZone, _ := d.findZoneByFqdn(fqdn)
	authZone = dns01.UnFqdn(authZone)
	unFqdn := dns01.UnFqdn(fqdn)
	subdomain := strings.TrimSuffix(unFqdn, "."+authZone)
	return subdomain, authZone
}

// Timeout returns the values (40*time.Minute, 60*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with Loopia.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
