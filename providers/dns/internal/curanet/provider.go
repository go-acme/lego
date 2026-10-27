// Package curanet implements a DNS provider for solving the DNS-01 challenge using Curanet.
package curanet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/internal/curanet/internal"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ClientID     string
	ClientSecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int

	APIKey     string       // Deprecated
	HTTPClient *http.Client // Deprecated
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProviderConfig return a DNSProvider instance configured for Curanet.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	client, err := newClient(config)
	if err != nil {
		return nil, err
	}

	if baseURL != "" {
		client.BaseURL, err = url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return err
	}

	record := internal.Record{
		Name: subDomain,
		Type: "TXT",
		TTL:  d.config.TTL,
		Data: info.Value,
	}

	err = d.client.CreateRecord(ctx, dns01.UnFqdn(authZone), record)
	if err != nil {
		return fmt.Errorf("create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return err
	}

	record, err := d.findRecord(ctx, dns01.UnFqdn(authZone), subDomain, info.Value)
	if err != nil {
		return err
	}

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), record.ID)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findRecord(ctx context.Context, zone, subDomain, value string) (*internal.Record, error) {
	records, err := d.client.GetRecords(ctx, zone, subDomain, "TXT")
	if err != nil {
		return nil, fmt.Errorf("get records: %w", err)
	}

	for _, record := range records {
		if record.Name == subDomain && record.Type == "TXT" && record.Data == value {
			return &record, nil
		}
	}

	return nil, errors.New("record not found")
}

func newClient(config *Config) (*internal.Client, error) {
	if config.APIKey != "" {
		return internal.NewClient(
			clientdebug.Wrap(
				internal.OAuthStaticAccessToken(config.HTTPClient, config.APIKey),
			),
		)
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.New("credentials missing")
	}

	return internal.NewClient(
		clientdebug.Wrap(
			internal.CreateOAuthClient(context.Background(), config.ClientID, config.ClientSecret),
		),
	)
}
