// Package beget implements a DNS provider for solving the DNS-01 challenge using beget.com DNS.
package beget

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/beget/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "BEGET_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 30*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for beget.com.
// Credentials must be passed in the environment variables:
// BEGET_USERNAME and BEGET_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("beget: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for beget.com.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("beget: the configuration of the DNS provider is nil")
	}

	if config.Username == "" || config.Password == "" {
		return nil, errors.New("beget: incomplete credentials, missing username and/or password")
	}

	client := internal.NewClient(config.Username, config.Password)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	records, err := d.client.GetTXTRecords(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("beget: get TXT records: %w", err)
	}

	records = append(records, internal.Record{
		Value:    info.Value,
		Data:     "", // NOTE: there are 2 fields in the API for the same thing.
		Priority: 10,
		TTL:      d.config.TTL,
	})

	err = d.client.ChangeTXTRecord(ctx, dns01.UnFqdn(info.EffectiveFQDN), records)
	if err != nil {
		return fmt.Errorf("beget: failed to create TXT records [domain: %s]: %w",
			dns01.UnFqdn(info.EffectiveFQDN), err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	records, err := d.client.GetTXTRecords(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("beget: get TXT records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	var updatedRecords []internal.Record
	for _, record := range records {
		if record.Data == info.Value {
			continue
		}

		updatedRecords = append(updatedRecords, record)
	}

	err = d.client.ChangeTXTRecord(ctx, dns01.UnFqdn(info.EffectiveFQDN), updatedRecords)
	if err != nil {
		return fmt.Errorf("beget: failed to remove TXT records [domain: %s]: %w",
			dns01.UnFqdn(info.EffectiveFQDN), err)
	}

	return nil
}
