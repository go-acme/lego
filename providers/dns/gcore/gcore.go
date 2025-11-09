package gcore

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
	"github.com/go-acme/lego/v4/providers/dns/internal/gcore"
)

// Environment variables names.
const (
	envNamespace = "GCORE_"

	EnvPermanentAPIToken = envNamespace + "PERMANENT_API_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	defaultPropagationTimeout = 360 * time.Second
	defaultPollingInterval    = 20 * time.Second
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config for DNSProvider.
type Config struct {
	APIToken           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}
}

// DNSProvider an implementation of challenge.Provider contract.
type DNSProvider struct {
	config *Config
	client *gcore.Client
}

// NewDNSProvider returns an instance of DNSProvider configured for G-Core DNS API.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPermanentAPIToken)
	if err != nil {
		return nil, fmt.Errorf("gcore: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvPermanentAPIToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for G-Core DNS API.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gcore: the configuration of the DNS provider is nil")
	}

	if config.APIToken == "" {
		return nil, errors.New("gcore: incomplete credentials provided")
	}

	client := gcore.NewClient(config.APIToken)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zone, err := d.guessZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gcore: %w", err)
	}

	err = d.client.AddRRSet(ctx, zone, dns01.UnFqdn(info.EffectiveFQDN), info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("gcore: add txt record: %w", err)
	}

	return nil
}

// CleanUp removes the record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	zone, err := d.guessZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gcore: %w", err)
	}

	err = d.client.DeleteRRSet(ctx, zone, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("gcore: remove txt record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) guessZone(ctx context.Context, fqdn string) (string, error) {
	var lastErr error

	for zone := range dns01.UnFqdnDomainsSeq(fqdn) {
		dnsZone, err := d.client.GetZone(ctx, zone)
		if err != nil {
			lastErr = err
			continue
		}

		return dnsZone.Name, nil
	}

	return "", fmt.Errorf("zone %q not found: %w", fqdn, lastErr)
}
