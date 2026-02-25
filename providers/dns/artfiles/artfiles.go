// Package artfiles implements a DNS provider for solving the DNS-01 challenge using ArtFiles.
package artfiles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/env"
	"github.com/go-acme/lego/v5/providers/dns/artfiles/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "ARTFILES_"

	EnvUsername = envNamespace + "USERNAME"
	EnvPassword = envNamespace + "PASSWORD"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 6*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
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

// NewDNSProvider returns a DNSProvider instance configured for ArtFiles.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvPassword)
	if err != nil {
		return nil, fmt.Errorf("artfiles: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Password = values[EnvPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ArtFiles.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("artfiles: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("artfiles: %w", err)
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
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("artfiles: %w", err)
	}

	records, err := d.client.GetRecords(ctx, zone)
	if err != nil {
		return fmt.Errorf("artfiles: get records: %w", err)
	}

	rv := internal.RecordValue{}

	if len(records["TXT"]) > 0 {
		var raw string

		err = json.Unmarshal(records["TXT"], &raw)
		if err != nil {
			return fmt.Errorf("artfiles: unmarshal TXT records: %w", err)
		}

		rv = internal.ParseRecordValue(raw)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("artfiles: %w", err)
	}

	rv.Add(subDomain, info.Value)

	err = d.client.SetRecords(ctx, zone, "TXT", rv)
	if err != nil {
		return fmt.Errorf("artfiles: set TXT records: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zone, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("artfiles: %w", err)
	}

	records, err := d.client.GetRecords(ctx, zone)
	if err != nil {
		return fmt.Errorf("artfiles: get records: %w", err)
	}

	var raw string

	err = json.Unmarshal(records["TXT"], &raw)
	if err != nil {
		return fmt.Errorf("artfiles: unmarshal TXT records: %w", err)
	}

	rv := internal.ParseRecordValue(raw)

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("artfiles: %w", err)
	}

	rv.RemoveValue(subDomain, info.Value)

	err = d.client.SetRecords(ctx, zone, "TXT", rv)
	if err != nil {
		return fmt.Errorf("artfiles: set TXT records: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (string, error) {
	domains, err := d.client.GetDomains(ctx)
	if err != nil {
		return "", fmt.Errorf("artfiles: get domains: %w", err)
	}

	var zone string

	for s := range dns01.UnFqdnDomainsSeq(fqdn) {
		if slices.Contains(domains, s) {
			zone = s
		}
	}

	if zone == "" {
		return "", fmt.Errorf("artfiles: could not find the zone for domain %q", fqdn)
	}

	return zone, nil
}
