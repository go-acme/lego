// Package constellix implements a DNS provider for solving the DNS-01 challenge using Constellix DNS.
package constellix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/constellix/internal"
	"github.com/hashicorp/go-retryablehttp"
)

// Environment variables names.
const (
	envNamespace = "CONSTELLIX_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvSecretKey = envNamespace + "SECRET_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	SecretKey          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
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

// NewDNSProvider returns a DNSProvider instance configured for Constellix.
// Credentials must be passed in the environment variables:
// CONSTELLIX_API_KEY and CONSTELLIX_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("constellix: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.SecretKey = values[EnvSecretKey]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Constellix.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("constellix: the configuration of the DNS provider is nil")
	}

	if config.SecretKey == "" || config.APIKey == "" {
		return nil, errors.New("constellix: incomplete credentials, missing secret key and/or API key")
	}

	tr, err := internal.NewTokenTransport(config.APIKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("constellix: %w", err)
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.HTTPClient = tr.Wrap(config.HTTPClient)
	retryClient.Backoff = backoff

	client := internal.NewClient(retryClient.StandardClient())

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

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("constellix: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	dom, err := d.client.Domains.GetByName(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("constellix: failed to get domain (%s): %w", authZone, err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("constellix: %w", err)
	}

	records, err := d.client.TxtRecords.Search(ctx, dom.ID, internal.Exact, recordName)
	if err != nil {
		return fmt.Errorf("constellix: failed to search TXT records: %w", err)
	}

	if len(records) > 1 {
		return errors.New("constellix: failed to get TXT records")
	}

	// TXT record entry already existing
	if len(records) == 1 {
		return d.appendRecordValue(ctx, dom, records[0].ID, info.Value)
	}

	err = d.createRecord(ctx, dom, info.EffectiveFQDN, recordName, info.Value)
	if err != nil {
		return fmt.Errorf("constellix: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("constellix: could not find zone for domain %q: %w", domain, err)
	}

	ctx := context.Background()

	dom, err := d.client.Domains.GetByName(ctx, dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("constellix: failed to get domain (%s): %w", authZone, err)
	}

	recordName, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("constellix: %w", err)
	}

	records, err := d.client.TxtRecords.Search(ctx, dom.ID, internal.Exact, recordName)
	if err != nil {
		return fmt.Errorf("constellix: failed to search TXT records: %w", err)
	}

	if len(records) > 1 {
		return errors.New("constellix: failed to get TXT records")
	}

	if len(records) == 0 {
		return nil
	}

	record, err := d.client.TxtRecords.Get(ctx, dom.ID, records[0].ID)
	if err != nil {
		return fmt.Errorf("constellix: failed to get TXT records: %w", err)
	}

	if !containsValue(record, info.Value) {
		return nil
	}

	// only 1 record value, the whole record must be deleted.
	if len(record.Value) == 1 {
		_, err = d.client.TxtRecords.Delete(ctx, dom.ID, record.ID)
		if err != nil {
			return fmt.Errorf("constellix: failed to delete TXT records: %w", err)
		}
		return nil
	}

	err = d.removeRecordValue(ctx, dom, record, info.Value)
	if err != nil {
		return fmt.Errorf("constellix: %w", err)
	}

	return nil
}

func (d *DNSProvider) createRecord(ctx context.Context, dom internal.Domain, fqdn, recordName, value string) error {
	request := internal.RecordRequest{
		Name: recordName,
		TTL:  d.config.TTL,
		RoundRobin: []internal.RecordValue{
			{Value: fmt.Sprintf(`%q`, value)},
		},
	}

	_, err := d.client.TxtRecords.Create(ctx, dom.ID, request)
	if err != nil {
		return fmt.Errorf("failed to create TXT record %s: %w", fqdn, err)
	}

	return nil
}

func (d *DNSProvider) appendRecordValue(ctx context.Context, dom internal.Domain, recordID int64, value string) error {
	record, err := d.client.TxtRecords.Get(ctx, dom.ID, recordID)
	if err != nil {
		return fmt.Errorf("failed to get TXT records: %w", err)
	}

	if containsValue(record, value) {
		return nil
	}

	request := internal.RecordRequest{
		Name:       record.Name,
		TTL:        record.TTL,
		RoundRobin: append(record.RoundRobin, internal.RecordValue{Value: fmt.Sprintf(`%q`, value)}),
	}

	_, err = d.client.TxtRecords.Update(ctx, dom.ID, record.ID, request)
	if err != nil {
		return fmt.Errorf("failed to update TXT records: %w", err)
	}

	return nil
}

func (d *DNSProvider) removeRecordValue(ctx context.Context, dom internal.Domain, record *internal.Record, value string) error {
	request := internal.RecordRequest{
		Name: record.Name,
		TTL:  record.TTL,
	}

	for _, val := range record.Value {
		if val.Value != fmt.Sprintf(`%q`, value) {
			request.RoundRobin = append(request.RoundRobin, val)
		}
	}

	_, err := d.client.TxtRecords.Update(ctx, dom.ID, record.ID, request)
	if err != nil {
		return fmt.Errorf("failed to update TXT records: %w", err)
	}

	return nil
}

func containsValue(record *internal.Record, value string) bool {
	if record == nil {
		return false
	}

	qValue := fmt.Sprintf(`%q`, value)

	return slices.ContainsFunc(record.Value, func(val internal.RecordValue) bool {
		return val.Value == qValue
	})
}

func backoff(minimum, maximum time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil {
		// https://api.dns.constellix.com/v4/docs#section/Using-the-API/Rate-Limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			if s, ok := resp.Header["X-Ratelimit-Reset"]; ok {
				if sleep, err := strconv.ParseInt(s[0], 10, 64); err == nil {
					return time.Second * time.Duration(sleep)
				}
			}
		}
	}

	return retryablehttp.DefaultBackoff(minimum, maximum, attemptNum, resp)
}
