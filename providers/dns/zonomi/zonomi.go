// Package zonomi implements a DNS provider for solving the DNS-01 challenge using Zonomi DNS.
package zonomi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/rimuhosting"
)

// Environment variables names.
const (
	envNamespace = "ZONOMI_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 3600),
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
	client *rimuhosting.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Zonomi.
// Credentials must be passed in the environment variables.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("zonomi: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Zonomi.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("zonomi: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("zonomi: incomplete credentials, missing API key")
	}

	client := rimuhosting.NewClient(config.APIKey)
	client.BaseURL = rimuhosting.DefaultZonomiBaseURL

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	records, err := d.client.FindTXTRecords(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("zonomi: failed to find record(s) for %s: %w", domain, err)
	}

	actions := []rimuhosting.ActionParameter{
		rimuhosting.NewAddRecordAction(dns01.UnFqdn(info.EffectiveFQDN), info.Value, d.config.TTL),
	}

	for _, record := range records {
		actions = append(actions, rimuhosting.NewAddRecordAction(record.Name, record.Content, d.config.TTL))
	}

	_, err = d.client.DoActions(ctx, actions...)
	if err != nil {
		return fmt.Errorf("zonomi: failed to add record(s) for %s: %w", domain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	action := rimuhosting.NewDeleteRecordAction(dns01.UnFqdn(info.EffectiveFQDN), info.Value)

	_, err := d.client.DoActions(context.Background(), action)
	if err != nil {
		return fmt.Errorf("zonomi: failed to delete record for %s: %w", domain, err)
	}

	return nil
}
