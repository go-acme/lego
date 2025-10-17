// Package anexia implements a DNS provider for solving the DNS-01 challenge using Anexia CloudDNS.
package anexia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	uuid "github.com/satori/go.uuid"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/clouddns"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
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
	client clouddns.API

	recordIDs   map[string]uuid.UUID
	recordIDsMu sync.Mutex
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

	opts := []client.Option{
		client.TokenFromString(config.Token),
	}

	if config.APIURL != "" {
		opts = append(opts, client.BaseURL(config.APIURL))
	}

	if config.HTTPClient != nil {
		opts = append(opts, client.HTTPClient(config.HTTPClient))
	}

	c, err := client.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("anexia: failed to create client: %w", err)
	}

	return &DNSProvider{
		config:    config,
		client:    clouddns.NewAPI(c),
		recordIDs: make(map[string]uuid.UUID),
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

	recordReq := zone.RecordRequest{
		Name:  recordName,
		Type:  "TXT",
		RData: info.Value,
		TTL:   d.config.TTL,
	}

	updatedZone, err := d.client.Zone().NewRecord(ctx, zoneName, recordReq)
	if err != nil {
		return fmt.Errorf("anexia: new record: %w", err)
	}

	// Find the newly created record in the updated zone
	recordID, err := d.findRecordID(ctx, updatedZone, zoneName, recordName, info.Value)
	if err != nil {
		return fmt.Errorf("anexia: find record ID: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

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

	// Get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("anexia: unknown record ID for '%s' '%s'", info.EffectiveFQDN, token)
	}

	err = d.client.Zone().DeleteRecord(ctx, dns01.UnFqdn(authZone), recordID)
	if err != nil {
		return fmt.Errorf("anexia: delete TXT record: %w", err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

// findRecordID attempts to find the record ID from the zone response.
// If the record is not immediately available in the response, it retries by querying the zone.
func (d *DNSProvider) findRecordID(ctx context.Context, updatedZone zone.Zone, zoneName, recordName, rdata string) (uuid.UUID, error) {
	// First, try to find the record in the immediate response
	recordID := findRecordIdentifier(updatedZone, recordName, rdata)
	if recordID != uuid.Nil {
		return recordID, nil
	}

	return backoff.Retry(ctx,
		func() (uuid.UUID, error) {
			currentZone, err := d.client.Zone().Get(ctx, zoneName)
			if err != nil {
				return uuid.Nil, fmt.Errorf("get zone: %w", err)
			}

			recordID := findRecordIdentifier(currentZone, recordName, rdata)
			if recordID == uuid.Nil {
				return uuid.Nil, fmt.Errorf("get record identifier: %w", err)
			}

			return recordID, nil
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(5*time.Second)),
		backoff.WithMaxElapsedTime(300*time.Second),
	)
}

func findRecordIdentifier(z zone.Zone, recordName, rdata string) uuid.UUID {
	if len(z.Revisions) == 0 {
		return uuid.Nil
	}

	// Check the first revision (index 0) which should be the current one

	for _, record := range z.Revisions[0].Records {
		if record.Name != recordName || record.Type != "TXT" {
			continue
		}

		if record.RData == rdata || record.RData == strconv.Quote(rdata) {
			return record.Identifier
		}
	}

	return uuid.Nil
}

func extractRecordName(fqdn, authZone string) (string, error) {
	if dns01.UnFqdn(fqdn) == dns01.UnFqdn(authZone) {
		// "@" for the root domain instead of an empty string.
		return "@", nil
	}

	return dns01.ExtractSubDomain(fqdn, authZone)
}
