// Package checkdomain implements a DNS provider for solving the DNS-01 challenge using CheckDomain DNS.
package checkdomain

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/checkdomain/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "CHECKDOMAIN_"

	EnvEndpoint = envNamespace + "ENDPOINT"
	EnvToken    = envNamespace + "TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Endpoint           *url.URL
	Token              string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 7*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for CheckDomain.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("checkdomain: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]

	endpoint, err := url.Parse(env.GetOrDefaultString(EnvEndpoint, internal.DefaultEndpoint))
	if err != nil {
		return nil, fmt.Errorf("checkdomain: invalid %s: %w", EnvEndpoint, err)
	}

	config.Endpoint = endpoint

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.Endpoint == nil {
		return nil, errors.New("checkdomain: invalid endpoint")
	}

	if config.Token == "" {
		return nil, errors.New("checkdomain: missing token")
	}

	client := internal.NewClient(
		clientdebug.Wrap(
			internal.OAuthStaticAccessToken(config.HTTPClient, config.Token),
		),
	)

	if config.Endpoint != nil {
		client.BaseURL = config.Endpoint
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	domainID, err := d.client.GetDomainIDByName(ctx, domain)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	err = d.client.CheckNameservers(ctx, domainID)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	err = d.client.CreateRecord(ctx, domainID, &internal.Record{
		Name:  info.EffectiveFQDN,
		TTL:   d.config.TTL,
		Type:  "TXT",
		Value: info.Value,
	})
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record previously created.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	domainID, err := d.client.GetDomainIDByName(ctx, domain)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	err = d.client.CheckNameservers(ctx, domainID)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	info := dns01.GetChallengeInfo(domain, keyAuth)

	defer d.client.CleanCache(info.EffectiveFQDN)

	err = d.client.DeleteTXTRecord(ctx, domainID, info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("checkdomain: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
