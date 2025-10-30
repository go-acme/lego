package brandit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/brandit/internal"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "BRANDIT_"

	EnvAPIKey      = envNamespace + "API_KEY"
	EnvAPIUsername = envNamespace + "API_USERNAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey      string
	APIUsername string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
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

	records   map[string]string
	recordsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for BrandIT.
// Credentials must be passed in the environment variables: BRANDIT_API_KEY, BRANDIT_API_USERNAME.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPIUsername)
	if err != nil {
		return nil, fmt.Errorf("brandit: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APIUsername = values[EnvAPIUsername]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for BrandIT.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("brandit: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIUsername, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("brandit: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:  config,
		client:  client,
		records: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("brandit: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("brandit: %w", err)
	}

	ctx := context.Background()

	record := internal.Record{
		Type:    "TXT",
		Name:    subDomain,
		Content: info.Value,
		TTL:     d.config.TTL,
	}

	// find the account associated with the domain
	account, err := d.client.StatusDomain(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("brandit: status domain: %w", err)
	}

	// Find the next record id
	recordID, err := d.client.ListRecords(ctx, account.Registrar[0], dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("brandit: list records: %w", err)
	}

	result, err := d.client.AddRecord(ctx, dns01.UnFqdn(authZone), account.Registrar[0], strconv.Itoa(recordID.Total[0]), record)
	if err != nil {
		return fmt.Errorf("brandit: add record: %w", err)
	}

	d.recordsMu.Lock()
	d.records[token] = result.Record
	d.recordsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("brandit: could not find zone for domain %q: %w", domain, err)
	}

	// gets the record's unique ID
	d.recordsMu.Lock()
	dnsRecord, ok := d.records[token]
	d.recordsMu.Unlock()

	if !ok {
		return fmt.Errorf("brandit: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	ctx := context.Background()

	// find the account associated with the domain
	account, err := d.client.StatusDomain(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("brandit: status domain: %w", err)
	}

	records, err := d.client.ListRecords(ctx, account.Registrar[0], dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("brandit: list records: %w", err)
	}

	var recordID int

	for i, r := range records.RR {
		if r == dnsRecord {
			recordID = i
		}
	}

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), account.Registrar[0], dnsRecord, strconv.Itoa(recordID))
	if err != nil {
		return fmt.Errorf("brandit: delete record: %w", err)
	}

	// deletes record ID from map
	d.recordsMu.Lock()
	delete(d.records, token)
	d.recordsMu.Unlock()

	return nil
}
