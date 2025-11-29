// Package rimuhosting implements a DNS provider for solving the DNS-01 challenge using RimuHosting DNS.
package rimuhosting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
	"github.com/go-acme/lego/v4/providers/dns/internal/rimuhosting/internal"
)

const DefaultTTL = 3600

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProviderConfig return a DNSProvider instance configured for RimuHosting.
func NewDNSProviderConfig(config *Config, baseURL string) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("incomplete credentials, missing API key")
	}

	client := internal.NewClient(config.APIKey)

	if baseURL != "" {
		client.BaseURL = baseURL
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctx := context.Background()

	records, err := d.client.FindTXTRecords(ctx, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("failed to find record(s) for %s: %w", domain, err)
	}

	actions := []internal.ActionParameter{
		internal.NewAddRecordAction(dns01.UnFqdn(info.EffectiveFQDN), info.Value, d.config.TTL),
	}

	for _, record := range records {
		actions = append(actions, internal.NewAddRecordAction(record.Name, record.Content, d.config.TTL))
	}

	_, err = d.client.DoActions(ctx, actions...)
	if err != nil {
		return fmt.Errorf("failed to add record(s) for %s: %w", domain, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	action := internal.NewDeleteRecordAction(dns01.UnFqdn(info.EffectiveFQDN), info.Value)

	_, err := d.client.DoActions(context.Background(), action)
	if err != nil {
		return fmt.Errorf("failed to delete record for %s: %w", domain, err)
	}

	return nil
}
