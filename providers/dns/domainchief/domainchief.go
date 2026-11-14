// Package domainchief implements a DNS provider for solving the DNS-01 challenge using Domain Chief.
package domainchief

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/wait"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/domainchief/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DOMAINCHIEF_"

	EnvAPIToken = envNamespace + "API_TOKEN"
	EnvTeamSlug = envNamespace + "_TEAM_SLUG"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIToken string
	TeamSlug string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Domain Chief.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken)
	if err != nil {
		return nil, fmt.Errorf("domainchief: %w", err)
	}

	config := NewDefaultConfig()
	config.APIToken = values[EnvAPIToken]
	config.TeamSlug = env.GetOrFile(EnvTeamSlug)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Domain Chief.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("domainchief: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIToken, config.TeamSlug)
	if err != nil {
		return nil, fmt.Errorf("domainchief: %w", err)
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

	payload := internal.Challenge{
		Name:  dns01.UnFqdn(info.EffectiveFQDN),
		Value: info.Value,
	}

	err := wait.Retry(ctx,
		func() error {
			record, err := d.client.CreateACMERecord(ctx, payload)
			if err != nil {
				return backoff.Permanent(fmt.Errorf("create ACME record: %w", err))
			}

			switch record.Status {
			case internal.StatusPublished:
				return nil
			case internal.StatusFailed:
				return backoff.Permanent(fmt.Errorf("%s: ACME record publication failure", record.Status))
			default:
				return fmt.Errorf("ACME record status: %s", record.Status)
			}
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(d.config.PollingInterval)),
		backoff.WithMaxElapsedTime(d.config.PropagationTimeout),
		backoff.WithNotify(wait.SimpleNotify("domainchief: waiting for ACME record publication")),
	)
	if err != nil {
		return fmt.Errorf("domainchief: publishing: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	payload := internal.Challenge{
		Name:  dns01.UnFqdn(info.EffectiveFQDN),
		Value: info.Value,
	}

	err := d.client.DeleteACMERecord(ctx, payload)
	if err != nil {
		return fmt.Errorf("domainchief: delete ACME record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
