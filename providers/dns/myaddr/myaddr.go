// Package myaddr implements a DNS provider for solving the DNS-01 challenge using myaddr.{tools,dev,io}.
package myaddr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/myaddr/internal"
)

// Environment variables names.
const (
	envNamespace = "MYADDR_"

	EnvPrivateKeysMapping = envNamespace + "PRIVATE_KEYS_MAPPING"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Credentials map[string]string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for myaddr.{tools,dev,io}.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPrivateKeysMapping)
	if err != nil {
		return nil, fmt.Errorf("myaddr: %w", err)
	}

	config := NewDefaultConfig()

	credentials, err := env.ParsePairs(values[EnvPrivateKeysMapping])
	if err != nil {
		return nil, fmt.Errorf("myaddr: %w", err)
	}

	config.Credentials = credentials

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for myaddr.{tools,dev,io}.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("myaddr: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Credentials)
	if err != nil {
		return nil, fmt.Errorf("myaddr: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("myaddr: could not find zone for domain %q: %w", domain, err)
	}

	fullSubdomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("myaddr: %w", err)
	}

	_, after, found := strings.Cut(fullSubdomain, ".")
	if !found {
		return fmt.Errorf("myaddr: subdomain not found in: %q (%s)", fullSubdomain, info.EffectiveFQDN)
	}

	err = d.client.AddTXTRecord(context.Background(), after, info.Value)
	if err != nil {
		return fmt.Errorf("myaddr: add TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	// There is no API endpoint to delete a TXT record:
	// TXT records are automatically removed after a few minutes.
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}
