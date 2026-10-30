// Package mythicbeasts implements a DNS provider for solving the DNS-01 challenge using Mythic Beasts API.
package mythicbeasts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/mythicbeasts/internal"
)

// Environment variables names.
const (
	envNamespace = "MYTHICBEASTS_"

	EnvClientID     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"

	EnvUserName = envNamespace + "USERNAME" // Deprecated: use EnvClientID instead.
	EnvPassword = envNamespace + "PASSWORD" // Deprecated: use EnvClientSecret instead.

	EnvAPIEndpoint     = envNamespace + "API_ENDPOINT"
	EnvAuthAPIEndpoint = envNamespace + "AUTH_API_ENDPOINT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	EnvHTTPTimeout = envNamespace + "HTTP_TIMEOUT" // Deprecated
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	UserName string // Deprecated: use ClientID instead.
	Password string // Deprecated: use ClientSecret instead.

	ClientID     string
	ClientSecret string

	HTTPClient *http.Client // Deprecated

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	APIEndpoint        *url.URL
	AuthAPIEndpoint    *url.URL
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() (*Config, error) {
	apiEndpoint, err := url.Parse(env.GetOrDefaultString(EnvAPIEndpoint, internal.APIBaseURL))
	if err != nil {
		return nil, fmt.Errorf("mythicbeasts: Unable to parse API URL: %w", err)
	}

	authEndpoint, err := url.Parse(env.GetOrDefaultString(EnvAuthAPIEndpoint, internal.DefaultOAuthURL))
	if err != nil {
		return nil, fmt.Errorf("mythicbeasts: Unable to parse AUTH API URL: %w", err)
	}

	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		APIEndpoint:        apiEndpoint,
		AuthAPIEndpoint:    authEndpoint,
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 10*time.Second),
		},
	}, nil
}

func (c *Config) getClientID() string {
	if c.UserName != "" {
		return c.UserName
	}

	return c.ClientID
}

func (c *Config) getClientSecret() string {
	if c.Password != "" {
		return c.Password
	}

	return c.ClientSecret
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for mythicbeasts DNSv2 API.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.GetWithFallback([]string{EnvClientID, EnvUserName}, []string{EnvClientSecret, EnvPassword})
	if err != nil {
		return nil, fmt.Errorf("mythicbeasts: %w", err)
	}

	config, err := NewDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("mythicbeasts: %w", err)
	}

	config.ClientID = values[EnvClientID]
	config.ClientSecret = values[EnvClientSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for mythicbeasts DNSv2 API.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("mythicbeasts: the configuration of the DNS provider is nil")
	}

	if config.getClientID() == "" || config.getClientSecret() == "" {
		return nil, errors.New("mythicbeasts: incomplete credentials, missing client ID and/or client secret")
	}

	var authEndpoint string

	if config.AuthAPIEndpoint != nil {
		authEndpoint = config.AuthAPIEndpoint.String()
	}

	client := internal.NewClient(
		clientdebug.Wrap(
			internal.CreateOAuthConfig(authEndpoint, config.getClientID(), config.getClientSecret()).
				Client(context.Background()),
		),
	)

	if config.APIEndpoint != nil {
		client.APIEndpoint = config.APIEndpoint
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("mythicbeasts: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("mythicbeasts: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	err = d.client.CreateTXTRecord(ctx, authZone, subDomain, info.Value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("mythicbeasts: CreateTXTRecord: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("mythicbeasts: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("mythicbeasts: %w", err)
	}

	authZone = dns01.UnFqdn(authZone)

	err = d.client.RemoveTXTRecord(ctx, authZone, subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("mythicbeasts: RemoveTXTRecord: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
