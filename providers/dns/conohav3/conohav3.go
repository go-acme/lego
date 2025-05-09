// Package conohav3 implements a DNS provider for solving the DNS-01 challenge using ConoHa VPS Ver 3.0 DNS.
package conohav3

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/conohav3/internal"
)

// Environment variables names.
const (
	envNamespace = "CONOHAV3_"

	EnvRegion      = envNamespace + "REGION"
	EnvTenantID    = envNamespace + "TENANT_ID"
	EnvAPIUserID   = envNamespace + "API_USER_ID"
	EnvAPIPassword = envNamespace + "API_PASSWORD"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Region             string
	TenantID           string
	UserID             string
	Password           string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		Region:             env.GetOrDefaultString(EnvRegion, "c3j1"),
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
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
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ConoHa DNS.
// Credentials must be passed in the environment variables:
// CONOHAV3_TENANT_ID, CONOHAV3_API_USER_ID, CONOHAV3_API_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvTenantID, EnvAPIUserID, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("conohav3: %w", err)
	}

	config := NewDefaultConfig()
	config.TenantID = values[EnvTenantID]
	config.UserID = values[EnvAPIUserID]
	config.Password = values[EnvAPIPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ConoHa DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("conohav3: the configuration of the DNS provider is nil")
	}

	if config.TenantID == "" || config.UserID == "" || config.Password == "" {
		return nil, errors.New("conohav3: some credentials information are missing")
	}

	identifier, err := internal.NewIdentifier(config.Region)
	if err != nil {
		return nil, fmt.Errorf("conohav3: failed to create identity client: %w", err)
	}

	if config.HTTPClient != nil {
		identifier.HTTPClient = config.HTTPClient
	}

	auth := internal.Auth{
		Identity: internal.Identity{
			Methods: []string{"password"},
			Password: internal.Password{
				User: internal.User{
					ID:       config.UserID,
					Password: config.Password,
				},
			},
		},
		Scope: internal.Scope{
			Project: internal.Project{
				ID: config.TenantID,
			},
		},
	}

	token, err := identifier.GetToken(context.Background(), auth)
	if err != nil {
		return nil, fmt.Errorf("conohav3: failed to log in: %w", err)
	}

	client, err := internal.NewClient(config.Region, token)
	if err != nil {
		return nil, fmt.Errorf("conohav3: failed to create client: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("conohav3: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	id, err := d.client.GetDomainID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("conohav3: failed to get domain ID: %w", err)
	}

	record := internal.Record{
		Name: info.EffectiveFQDN,
		Type: "TXT",
		Data: info.Value,
		TTL:  d.config.TTL,
	}

	err = d.client.CreateRecord(ctx, id, record)
	if err != nil {
		return fmt.Errorf("conohav3: failed to create record: %w", err)
	}

	return nil
}

// CleanUp clears ConoHa DNS TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("conohav3: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	domID, err := d.client.GetDomainID(ctx, authZone)
	if err != nil {
		return fmt.Errorf("conohav3: failed to get domain ID: %w", err)
	}

	recID, err := d.client.GetRecordID(ctx, domID, info.EffectiveFQDN, "TXT", info.Value)
	if err != nil {
		return fmt.Errorf("conohav3: failed to get record ID: %w", err)
	}

	err = d.client.DeleteRecord(ctx, domID, recID)
	if err != nil {
		return fmt.Errorf("conohav3: failed to delete record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
