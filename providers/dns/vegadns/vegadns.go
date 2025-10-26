// Package vegadns implements a DNS provider for solving the DNS-01 challenge using VegaDNS.
package vegadns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/nrdcg/vegadns"
)

// Environment variables names.
const (
	envNamespace = "VEGADNS_"

	EnvKey    = "SECRET_VEGADNS_KEY"
	EnvSecret = "SECRET_VEGADNS_SECRET"
	EnvURL    = envNamespace + "URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	BaseURL   string
	APIKey    string
	APISecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 10),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 12*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, time.Minute),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *vegadns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for VegaDNS.
// Credentials must be passed in the environment variables:
// VEGADNS_URL, SECRET_VEGADNS_KEY, SECRET_VEGADNS_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvURL)
	if err != nil {
		return nil, fmt.Errorf("vegadns: %w", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values[EnvURL]
	config.APIKey = env.GetOrFile(EnvKey)
	config.APISecret = env.GetOrFile(EnvSecret)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for VegaDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vegadns: the configuration of the DNS provider is nil")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	config.HTTPClient = clientdebug.Wrap(config.HTTPClient)

	client, err := vegadns.NewClient(config.BaseURL,
		vegadns.WithOAuth(config.APIKey, config.APISecret),
		vegadns.WithHTTPClient(config.HTTPClient),
	)
	if err != nil {
		return nil, fmt.Errorf("vegadns: %w", err)
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	domainID, err := d.findDomainID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("vegadns: find domain ID for %s: %w", info.EffectiveFQDN, err)
	}

	err = d.client.CreateTXTRecord(ctx, domainID, dns01.UnFqdn(info.EffectiveFQDN), info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("vegadns: create TXT record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	domainID, err := d.findDomainID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("vegadns: find domain ID for %s: %w", info.EffectiveFQDN, err)
	}

	recordID, err := d.client.GetRecordID(ctx, domainID, dns01.UnFqdn(info.EffectiveFQDN), "TXT")
	if err != nil {
		return fmt.Errorf("vegadns: get Record ID: %w", err)
	}

	err = d.client.DeleteRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("vegadns: %w", err)
	}

	return nil
}

func (d *DNSProvider) findDomainID(ctx context.Context, fqdn string) (int, error) {
	for host := range dns01.UnFqdnDomainsSeq(fqdn) {
		id, err := d.client.GetDomainID(ctx, host)
		if err != nil {
			continue
		}

		return id, nil
	}

	return 0, errors.New("domain not found")
}
