// Package enum implements a DNS provider for solving the DNS-01 challenge using Enum.
package enum

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/enum/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "ENUM_"

	EnvAPIToken    = envNamespace + "API_TOKEN"
	EnvCredentials = envNamespace + "CREDENTIALS"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const TokenZoneWildcard = "$"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Credentials map[string]string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
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

// NewDNSProvider returns a DNSProvider instance configured for Enum.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	credentials, err := getCredentials()
	if err != nil {
		return nil, fmt.Errorf("enum: %w", err)
	}

	config.Credentials = credentials

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Enum.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("enum: the configuration of the DNS provider is nil")
	}

	if len(config.Credentials) == 0 {
		return nil, errors.New("enum: credentials missing")
	}

	client, err := internal.NewClient()
	if err != nil {
		return nil, fmt.Errorf("enum: %w", err)
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

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("enum: could not find zone for domain %q: %w", domain, err)
	}

	tok, err := getToken(d.config, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("enum: %w", err)
	}

	ctxAuth := internal.WithContext(ctx, tok)

	projectID, err := d.getProjectID(ctxAuth)
	if err != nil {
		return fmt.Errorf("enum: %w", err)
	}

	in := &internal.GetZoneByNameRequest{
		ProjectID: projectID,
		Name:      dns01.UnFqdn(authZone),
	}

	zone, err := d.client.GetZoneByName(ctxAuth, in)
	if err != nil {
		return fmt.Errorf("enum: get zone: %w", err)
	}

	request := &internal.AddRecordSetValueRequest{
		ProjectID: projectID,
		ZoneID:    zone.Zone.ID,
		Name:      dns01.UnFqdn(info.EffectiveFQDN),
		Type:      "TXT",
		TTL:       d.config.TTL,
		Value: &internal.DNSRecordValue{
			Content: info.Value,
		},
	}

	_, err = d.client.AddRecordSetValue(ctxAuth, request)
	if err != nil {
		return fmt.Errorf("enum: add record set value: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("enum: could not find zone for domain %q: %w", domain, err)
	}

	tok, err := getToken(d.config, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("enum: %w", err)
	}

	ctxAuth := internal.WithContext(ctx, tok)

	projectID, err := d.getProjectID(ctxAuth)
	if err != nil {
		return fmt.Errorf("enum: %w", err)
	}

	in := &internal.GetZoneByNameRequest{
		ProjectID: projectID,
		Name:      dns01.UnFqdn(authZone),
	}

	zone, err := d.client.GetZoneByName(ctxAuth, in)
	if err != nil {
		return fmt.Errorf("enum: get zone: %w", err)
	}

	request := &internal.RemoveRecordSetValueRequest{
		ProjectID: projectID,
		ZoneID:    zone.Zone.ID,
		Name:      dns01.UnFqdn(info.EffectiveFQDN),
		Type:      "TXT",
		Content:   info.Value,
	}

	_, err = d.client.RemoveRecordSetValue(ctxAuth, request)
	if err != nil {
		return fmt.Errorf("enum: remove record set value: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getProjectID(ctx context.Context) (string, error) {
	projects, err := d.client.ListProjects(ctx)
	if err != nil {
		return "", fmt.Errorf("list projects: %w", err)
	}

	if len(projects.Projects) != 1 {
		return "", fmt.Errorf("found %d projects, expected 1", len(projects.Projects))
	}

	return projects.Projects[0].ID, nil
}

func getCredentials() (map[string]string, error) {
	values, err := env.Get(EnvAPIToken)
	if err == nil {
		return map[string]string{
			TokenZoneWildcard: values[EnvAPIToken],
		}, nil
	}

	values, err = env.Get(EnvCredentials)
	if err == nil {
		return env.ParsePairs(values[EnvCredentials])
	}

	return nil, fmt.Errorf("some credentials information are missing: %s or %s", EnvAPIToken, EnvCredentials)
}

func getToken(config *Config, authZone string) (string, error) {
	tok, ok := config.Credentials[authZone]
	if ok {
		return tok, nil
	}

	tok, ok = config.Credentials[TokenZoneWildcard]
	if ok {
		return tok, nil
	}

	return "", fmt.Errorf("credentials not found for %q", authZone)
}
