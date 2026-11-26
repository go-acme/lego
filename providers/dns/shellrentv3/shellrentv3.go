// Package shellrentv3 implements a DNS provider for solving the DNS-01 challenge using Shellrent v3.
package shellrentv3

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v5/providers/dns/shellrentv3/internal"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Environment variables names.
const (
	envNamespace = "SHELLRENTV3_"

	EnvClientID     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type recordKey struct {
	domainID int
	recordID string
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ClientID     string
	ClientSecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
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

	recordIDs   map[string]recordKey
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Shellrent v3.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvClientID, EnvClientSecret)
	if err != nil {
		return nil, fmt.Errorf("shellrentv3: %w", err)
	}

	config := NewDefaultConfig()
	config.ClientID = values[EnvClientID]
	config.ClientSecret = values[EnvClientSecret]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Shellrent v3.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("shellrentv3: the configuration of the DNS provider is nil")
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, errors.New("shellrentv3: credentials missing")
	}

	return newDNSProviderConfig(config, internal.CreateOAuthConfig("", config.ClientID, config.ClientSecret)), nil
}

func newDNSProviderConfig(config *Config, oauthConfig *clientcredentials.Config) *DNSProvider {
	ctx := context.Background()

	if config.HTTPClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, config.HTTPClient)
	}

	client := internal.NewClient(clientdebug.Wrap(oauthConfig.Client(ctx)))

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]recordKey),
	}
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("shellrentv3: could not find zone for domain %q: %w", domain, err)
	}

	domainID, err := d.findDomainID(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("shellrentv3: find domain ID: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("shellrentv3: %w", err)
	}

	record := internal.Record{
		Type:        "TXT",
		Host:        subDomain,
		Destination: info.Value,
	}

	created, err := d.client.CreateRecord(ctx, domainID, record)
	if err != nil {
		return fmt.Errorf("shellrentv3: create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordKey{domainID: domainID, recordID: created.ID}
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	key, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("shellrentv3: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err := d.client.DeleteRecord(ctx, key.domainID, key.recordID)
	if err != nil {
		return fmt.Errorf("shellrentv3: delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDomainID(ctx context.Context, zone string) (int, error) {
	purchases, err := d.client.ListPurchases(ctx, zone)
	if err != nil {
		return 0, fmt.Errorf("list purchases: %w", err)
	}

	for _, purchase := range purchases {
		if purchase.DomainID == 0 {
			continue
		}

		domain, err := d.client.GetDomain(ctx, purchase.DomainID)
		if err != nil {
			return 0, fmt.Errorf("get domain %d: %w", purchase.DomainID, err)
		}

		if domain == nil {
			continue
		}

		if strings.EqualFold(domain.DomainName.Unicode, zone) || strings.EqualFold(domain.DomainName.Punycode, zone) {
			return domain.DomainID, nil
		}
	}

	return 0, fmt.Errorf("domain %q not found", zone)
}
