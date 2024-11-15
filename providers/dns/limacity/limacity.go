// Package limacity implements a DNS provider for solving the DNS-01 challenge using Lima-City DNS.
package limacity

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
	"github.com/go-acme/lego/v4/providers/dns/limacity/internal"
	"github.com/miekg/dns"
)

// Environment variables names.
const (
	envNamespace = "LIMACITY_"

	EnvAPIKey = envNamespace + "API_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
	EnvSequenceInterval   = envNamespace + "SEQUENCE_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	SequenceInterval   time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 8*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 80*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond(EnvSequenceInterval, 90*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	domainIDs   map[string]int
	domainIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Lima-City DNS.
// LIMACITY_API_KEY must be passed in the environment variables.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("limacity: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Lima-City DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("limacity: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("limacity: APIKey is missing")
	}

	client := internal.NewClient(config.APIKey)

	return &DNSProvider{
		config:    config,
		client:    client,
		domainIDs: make(map[string]int),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	domains, err := d.client.GetDomains(context.Background())
	if err != nil {
		return fmt.Errorf("limacity: get domains: %w", err)
	}

	dom, err := findDomain(domains, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("limacity: find domain: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, dom.UnicodeFqdn)
	if err != nil {
		return fmt.Errorf("limacity: %w", err)
	}

	record := internal.Record{
		Name:    subDomain,
		Content: info.Value,
		TTL:     d.config.TTL,
		Type:    "TXT",
	}

	err = d.client.AddRecord(context.Background(), dom.ID, record)
	if err != nil {
		return fmt.Errorf("limacity: add record: %w", err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[token] = dom.ID
	d.domainIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	// gets the domain's unique ID
	d.domainIDsMu.Lock()
	domainID, ok := d.domainIDs[token]
	d.domainIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("limacity: unknown domain ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	records, err := d.client.GetRecords(context.Background(), domainID)
	if err != nil {
		return fmt.Errorf("limacity: get records: %w", err)
	}

	var recordID int
	for _, record := range records {
		if record.Type == "TXT" && record.Content == strconv.Quote(info.Value) {
			recordID = record.ID
			break
		}
	}

	if recordID == 0 {
		return errors.New("limacity: TXT record not found")
	}

	err = d.client.DeleteRecord(context.Background(), domainID, recordID)
	if err != nil {
		return fmt.Errorf("limacity: delete record (domain ID=%d, record ID=%d): %w", domainID, recordID, err)
	}

	return nil
}

func findDomain(domains []internal.Domain, fqdn string) (internal.Domain, error) {
	labelIndexes := dns.Split(fqdn)

	for _, index := range labelIndexes {
		f := fqdn[index:]
		domain := dns01.UnFqdn(f)

		for _, dom := range domains {
			if dom.UnicodeFqdn == domain || dom.UnicodeFqdn == f {
				return dom, nil
			}
		}
	}

	return internal.Domain{}, fmt.Errorf("domain %s not found", fqdn)
}
