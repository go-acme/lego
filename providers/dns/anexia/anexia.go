// Package anexia implements a DNS provider for solving the DNS-01 challenge using Anexia CloudDNS.
package anexia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/anexia/internal"
)

// Environment variables names.
const (
	envNamespace = "ANEXIA_"

	EnvToken  = envNamespace + "TOKEN"
	EnvAPIURL = envNamespace + "API_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultTTL = 300

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token  string
	APIURL string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
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

// NewDNSProvider returns a DNSProvider instance configured for Anexia CloudDNS.
// Credentials must be passed in the environment variable: ANEXIA_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("anexia: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]
	config.APIURL = env.GetOrFile(EnvAPIURL)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Anexia CloudDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("anexia: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("anexia: incomplete credentials, missing token")
	}

	client, err := internal.NewClient(config.Token)
	if err != nil {
		return nil, fmt.Errorf("anexia: %w", err)
	}

	if config.APIURL != "" {
		var err error
		client.BaseURL, err = url.Parse(config.APIURL)
		if err != nil {
			return nil, fmt.Errorf("anexia: %w", err)
		}
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("anexia: could not find zone for domain %q: %w", domain, err)
	}

	recordName, err := extractRecordName(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("anexia: %w", err)
	}

	zoneName := dns01.UnFqdn(authZone)

	recordReq := internal.Record{
		Name:  recordName,
		Type:  "TXT",
		RData: info.Value,
		TTL:   d.config.TTL,
	}

	// Ignores returned zone, because of UUID unstability.
	// https://github.com/go-acme/lego/pull/2675#issuecomment-3418678194
	_, err = d.client.CreateRecord(ctx, zoneName, recordReq)
	if err != nil {
		return fmt.Errorf("anexia: new record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("anexia: could not find zone for domain %q: %w", domain, err)
	}

	recordName, err := extractRecordName(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("anexia: %w", err)
	}

	recordID, err := d.findRecordID(ctx, dns01.UnFqdn(authZone), recordName, info.Value)
	if err != nil {
		return fmt.Errorf("anexia: %w", err)
	}

	err = d.client.DeleteRecord(ctx, dns01.UnFqdn(authZone), recordID)
	if err != nil {
		return fmt.Errorf("anexia: delete TXT record: %w", err)
	}

	return nil
}

// findRecordID attempts to find the record ID from the zone response.
// If the record is not immediately available in the response, it retries by querying the zone.
func (d *DNSProvider) findRecordID(ctx context.Context, zoneName, recordName, rdata string) (string, error) {
	return backoff.Retry(ctx,
		func() (string, error) {
			currentZone, err := d.client.GetZone(ctx, zoneName)
			if err != nil {
				return "", backoff.Permanent(fmt.Errorf("get zone: %w", err))
			}

			recordID := findRecordIdentifier(currentZone, recordName, rdata)
			if recordID == "" {
				return "", fmt.Errorf("get record identifier: %w", err)
			}

			return recordID, nil
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(5*time.Second)),
		backoff.WithMaxElapsedTime(300*time.Second),
	)
}

func findRecordIdentifier(zone *internal.Zone, recordName, rdata string) string {
	if len(zone.Revisions) == 0 {
		return ""
	}

	// Check the first revision (index 0) which should be the current one

	for _, record := range zone.Revisions[0].Records {
		if record.Name != recordName || record.Type != "TXT" {
			continue
		}

		if record.RData == rdata || record.RData == strconv.Quote(rdata) {
			return record.Identifier
		}
	}

	return ""
}

func extractRecordName(fqdn, authZone string) (string, error) {
	if dns01.UnFqdn(fqdn) == dns01.UnFqdn(authZone) {
		// "@" for the root domain instead of an empty string.
		return "@", nil
	}

	return dns01.ExtractSubDomain(fqdn, authZone)
}
