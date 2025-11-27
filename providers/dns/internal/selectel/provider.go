// Package selectel implements a DNS provider for solving the DNS-01 challenge using Selectel Domains API.
package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/selectel/internal"
)

const MinTTL = 60

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client

	// TODO(ldez): remove in v5?
	BaseURL string
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProviderConfig return a DNSProvider instance configured for selectel.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("credentials missing")
	}

	if config.TTL < MinTTL {
		return nil, fmt.Errorf("invalid TTL, TTL (%d) must be greater than %d", config.TTL, MinTTL)
	}

	client := internal.NewClient(config.Token)

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	var err error

	client.BaseURL, err = url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the Timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	domainObj, err := d.client.GetDomainByName(ctx, domain)
	if err != nil {
		return fmt.Errorf("get domain by name: %w", err)
	}

	txtRecord := internal.Record{
		Type:    "TXT",
		TTL:     d.config.TTL,
		Name:    info.EffectiveFQDN,
		Content: info.Value,
	}

	_, err = d.client.AddRecord(ctx, domainObj.ID, txtRecord)
	if err != nil {
		return fmt.Errorf("add record: %w", err)
	}

	return nil
}

// CleanUp removes a TXT record used for DNS-01 challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	recordName := dns01.UnFqdn(info.EffectiveFQDN)

	ctx := context.Background()

	// TODO(ldez) replace domain by FQDN to follow CNAME.
	domainObj, err := d.client.GetDomainByName(ctx, domain)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	records, err := d.client.ListRecords(ctx, domainObj.ID)
	if err != nil {
		return fmt.Errorf("list records: %w", err)
	}

	// Delete records with specific FQDN
	var lastErr error

	for _, record := range records {
		if record.Name == recordName {
			err = d.client.DeleteRecord(ctx, domainObj.ID, record.ID)
			if err != nil {
				lastErr = fmt.Errorf("delete record: %w", err)
			}
		}
	}

	return lastErr
}
