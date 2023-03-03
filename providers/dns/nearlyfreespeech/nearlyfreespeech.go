// Package nearlyfreespeech implements a DNS provider for solving the DNS-01 challenge using NearlyFreeSpeech.NET.
package nearlyfreespeech

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/nearlyfreespeech/internal"
)

// Environment variables names.
const (
	envNamespace = "NEARLYFREESPEECH_"

	EnvLogin  = envNamespace + "LOGIN"
	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string
	Login  string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, dns01.DefaultPropagationTimeout),
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

// NewDNSProvider returns a DNSProvider instance configured for NearlyFreeSpeech.NET.
// Credentials must be passed in the environment variable: NEARLYFREESPEECH_LOGIN, NEARLYFREESPEECH_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvLogin)
	if err != nil {
		return nil, fmt.Errorf("nearlyfreespeech: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.Login = values[EnvLogin]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for NearlyFreeSpeech.NET.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nearlyfreespeech: the configuration of the DNS provider is nil")
	}

	if config.Login == "" || config.APIKey == "" {
		return nil, errors.New("nearlyfreespeech: API credentials are missing")
	}

	client := internal.NewClient(config.Login, config.APIKey)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
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

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: could not determine zone for domain %q: %w", info.EffectiveFQDN, err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	record := internal.Record{
		Name: recordName,
		Type: "TXT",
		Data: info.Value,
		TTL:  d.config.TTL,
	}

	err = d.client.AddRecord(authZone, record)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: could not determine zone for domain %q: %w", info.EffectiveFQDN, err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	record := internal.Record{
		Name: recordName,
		Type: "TXT",
		Data: info.Value,
	}

	err = d.client.RemoveRecord(domain, record)
	if err != nil {
		return fmt.Errorf("nearlyfreespeech: %w", err)
	}

	return nil
}
