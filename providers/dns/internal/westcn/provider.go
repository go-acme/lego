package westcn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/westcn/internal"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username string
	Password string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client

	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProviderConfig return a DNSProvider instance configured for West.cn/西部数码.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if baseURL != "" {
		client.BaseURL, err = url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	record := internal.Record{
		Domain: dns01.UnFqdn(authZone),
		Host:   subDomain,
		Type:   "TXT",
		Value:  info.Value,
		TTL:    d.config.TTL,
	}

	recordID, err := d.client.AddRecord(context.Background(), record)
	if err != nil {
		return fmt.Errorf("add record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("could not find zone for domain %q: %w", domain, err)
	}

	// gets the record's unique ID
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	if !ok {
		return fmt.Errorf("unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.DeleteRecord(context.Background(), dns01.UnFqdn(authZone), recordID)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	// deletes record ID from map
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
