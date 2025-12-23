// Package euserv implements a DNS provider for solving the DNS-01 challenge using EUserv.
package euserv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/euserv/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "EUSERV_"

	EnvEmail    = envNamespace + "EMAIL"
	EnvPassword = envNamespace + "PASSWORD"
	EnvOrderID  = envNamespace + "ORDER_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Email    string
	Password string
	OrderID  string

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

	identifier *internal.Identifier
	client     *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for EUserv.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvEmail, EnvPassword, EnvOrderID)
	if err != nil {
		return nil, fmt.Errorf("euserv: %w", err)
	}

	config := NewDefaultConfig()
	config.Email = values[EnvEmail]
	config.Password = values[EnvPassword]
	config.OrderID = values[EnvOrderID]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for EUserv.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("euserv: the configuration of the DNS provider is nil")
	}

	identifier, err := internal.NewIdentifier(config.Email, config.Password, config.OrderID)
	if err != nil {
		return nil, fmt.Errorf("euserv: %w", err)
	}

	if config.HTTPClient != nil {
		identifier.HTTPClient = config.HTTPClient
	}

	identifier.HTTPClient = clientdebug.Wrap(identifier.HTTPClient)

	client := internal.NewClient()

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:     config,
		client:     client,
		identifier: identifier,
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	sessionID, err := d.identifier.Login(ctx)
	if err != nil {
		return fmt.Errorf("euserv: login: %w", err)
	}

	ctx = internal.WithContext(ctx, sessionID)

	defer func() { _ = d.identifier.Logout(ctx) }()

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("euserv: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("euserv: %w", err)
	}

	domainID, err := d.findDomainID(ctx, authZone, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("euserv: find domain ID: %w", err)
	}

	srRrequest := internal.SetRecordRequest{
		DomainID:  domainID,
		Subdomain: subDomain,
		Type:      "TXT",
		Content:   info.Value,
		TTL:       d.config.TTL,
	}

	err = d.client.SetRecord(ctx, srRrequest)
	if err != nil {
		return fmt.Errorf("euserv: set record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	sessionID, err := d.identifier.Login(ctx)
	if err != nil {
		return fmt.Errorf("euserv: login: %w", err)
	}

	ctx = internal.WithContext(ctx, sessionID)

	defer func() { _ = d.identifier.Logout(ctx) }()

	recordID, err := d.findRecordID(ctx, info)
	if err != nil {
		return fmt.Errorf("euserv: find record ID: %w", err)
	}

	err = d.client.RemoveRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("euserv: remove record: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findRecordID(ctx context.Context, info dns01.ChallengeInfo) (string, error) {
	grRequest := internal.GetRecordsRequest{
		Type:    "TXT",
		Content: info.Value,
	}

	domains, err := d.client.GetRecords(ctx, grRequest)
	if err != nil {
		return "", fmt.Errorf("get records: %w", err)
	}

	for _, domain := range domains {
		for _, record := range domain.DNSRecords {
			if record.Type.Value == "TXT" && record.Content.Value == info.Value {
				return record.ID.Value, nil
			}
		}
	}

	return "", errors.New("record not found")
}

func (d *DNSProvider) findDomainID(ctx context.Context, authZone, fqdn string) (string, error) {
	grRequest := internal.GetRecordsRequest{
		Keyword: authZone,
	}

	domains, err := d.client.GetRecords(ctx, grRequest)
	if err != nil {
		return "", fmt.Errorf("get records: %w", err)
	}

	for a := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, b := range domains {
			if b.Domain.Value == a {
				return b.ID.Value, nil
			}
		}
	}

	return "", errors.New("domain not found")
}
