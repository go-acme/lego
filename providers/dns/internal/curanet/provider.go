// Package curanet implements a DNS provider for solving the DNS-01 challenge using Curanet.
package curanet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
func NewDNSProviderConfig(config *Config, urlCfg URLConfiguration) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	client, err := newClient(config, urlCfg.TokenURL)
	if err != nil {
		return nil, err
	}

	if urlCfg.BaseURL != "" {
		client.BaseURL, err = url.Parse(urlCfg.BaseURL)
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

	record := internal.Record{
		Name: dns01.UnFqdn(info.EffectiveFQDN),
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

	record, err := d.findRecord(ctx, dns01.UnFqdn(authZone), dns01.UnFqdn(info.EffectiveFQDN), info.Value)
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

func (d *DNSProvider) findRecord(ctx context.Context, zone, fqdn, value string) (*internal.Record, error) {
	records, err := d.client.GetRecords(ctx, zone, fqdn, "TXT")
	if err != nil {
		return nil, fmt.Errorf("get records: %w", err)
	}

	for _, record := range records {
		if record.Name == fqdn && record.Type == "TXT" && (record.Data == value || record.Data == strconv.Quote(value)) {
			return &record, nil
		}
	}

	return nil, errors.New("record not found")
}

func newClient(config *Config, tokenURL string) (*internal.Client, error) {
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
			internal.CreateOAuthConfig(tokenURL, config.ClientID, config.ClientSecret).
				Client(context.Background()),
		),
	)
}

type URLConfiguration struct {
	BaseURL  string
	TokenURL string
}

func CreateURLConfiguration(code string) URLConfiguration {
	return URLConfiguration{
		BaseURL:  fmt.Sprintf(internal.BaseURLTemplate, code),
		TokenURL: fmt.Sprintf(internal.BaseOAuthURLTemplate, code),
	}
}
