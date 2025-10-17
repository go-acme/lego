// Package anxcloud implements a DNS provider for solving the DNS-01 challenge using Anexia CloudDNS.
package anxcloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
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
	Token              string
	APIURL             string
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
	api    clouddns.API

	recordIDs   map[string]uuid.UUID
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Anexia CloudDNS.
// Credentials must be passed in the environment variable: ANEXIA_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken)
	if err != nil {
		return nil, fmt.Errorf("anxcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvToken]
	config.APIURL = env.GetOrFile(EnvAPIURL)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Anexia CloudDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("anxcloud: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("anxcloud: incomplete credentials, missing token")
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
		return nil, fmt.Errorf("anxcloud: failed to create client: %w", err)
	}

	api := clouddns.NewAPI(c)

	return &DNSProvider{
		config:    config,
		api:       api,
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
		return fmt.Errorf("anxcloud: could not find zone for domain %q: %w", domain, err)
	}

	zoneName := dns01.UnFqdn(authZone)

	zoneAPI := d.api.Zone()

	// Extract record name (subdomain part)
	recordName := extractRecordName(info.EffectiveFQDN, authZone)

	// Create the TXT record
	recordReq := zone.RecordRequest{
		Name:  recordName,
		Type:  "TXT",
		RData: info.Value,
		TTL:   d.config.TTL,
	}

	updatedZone, err := zoneAPI.NewRecord(ctx, zoneName, recordReq)
	if err != nil {
		return fmt.Errorf("anxcloud: failed to create TXT record: %w", err)
	}

	// Find the newly created record in the updated zone
	recordID, err := d.findRecordID(ctx, updatedZone, zoneName, recordName, info.Value)
	if err != nil {
		return fmt.Errorf("anxcloud: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = recordID
	d.recordIDsMu.Unlock()

	log.Infof("anxcloud: new record for %s, ID %s", domain, recordID)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()

	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("anxcloud: could not find zone for domain %q: %w", domain, err)
	}

	zoneName := dns01.UnFqdn(authZone)
	recordName := extractRecordName(info.EffectiveFQDN, authZone)

	zoneAPI := d.api.Zone()

	// Get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()

	// If we don't have the record ID cached (e.g., Present() failed), try to find it
	if !ok {
		// Get the zone to access current revision
		currentZone, getErr := zoneAPI.Get(ctx, zoneName)
		if getErr != nil {
			return fmt.Errorf("anxcloud: failed to get zone for cleanup: %w", getErr)
		}

		// Check the first revision (index 0) which should be the current one
		if len(currentZone.Revisions) > 0 {
			revision := currentZone.Revisions[0]

			for _, record := range revision.Records {
				if matchTXTRecord(record, recordName, info.Value) {
					recordID = record.Identifier
					ok = true
					break
				}
			}
		}

		if !ok {
			// Don't fail cleanup if record doesn't exist
			return nil
		}
	}

	err = zoneAPI.DeleteRecord(ctx, zoneName, recordID)
	if err != nil {
		return fmt.Errorf("anxcloud: failed to delete TXT record: %w", err)
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
	var recordID uuid.UUID
	if len(updatedZone.Revisions) > 0 {
		for _, record := range updatedZone.Revisions[0].Records {
			if matchTXTRecord(record, recordName, rdata) {
				recordID = record.Identifier
				break
			}
		}
	}

	if recordID != uuid.Nil {
		return recordID, nil
	}

	// If not found, the record might not be immediately available in the response.
	// Query the zone with retries to find the newly created record.
	zoneAPI := d.api.Zone()

	const maxRetries = 60
	retryDelay := 5 * time.Second

	for i := range maxRetries {
		if i > 0 {
			time.Sleep(retryDelay)
		}

		// Get the zone to access current revision
		currentZone, err := zoneAPI.Get(ctx, zoneName)
		if err != nil {
			log.Warnf("anxcloud: failed to get zone (attempt %d/%d): %v", i+1, maxRetries, err)
			continue
		}

		// Check the first revision (index 0) which should be the current one
		if len(currentZone.Revisions) > 0 {
			revision := currentZone.Revisions[0]

			for _, record := range revision.Records {
				if matchTXTRecord(record, recordName, rdata) {
					return record.Identifier, nil
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("could not find created record after %d retries", maxRetries)
}

// matchTXTRecord checks if a record matches our search criteria.
// The Anexia API returns TXT record RData with quotes, so we need to handle both formats.
func matchTXTRecord(record zone.Record, recordName, rdata string) bool {
	if record.Name != recordName || record.Type != "TXT" {
		return false
	}

	// The API may return TXT records with quotes around the value
	// Try exact match first, then try with quotes added
	if record.RData == rdata {
		return true
	}

	// Try with quotes
	quotedRData := "\"" + rdata + "\""
	return record.RData == quotedRData
}

// extractRecordName extracts the record name from the FQDN.
// The Anexia CloudDNS API requires "@" for the root domain instead of an empty string.
func extractRecordName(fqdn, authZone string) string {
	name := dns01.UnFqdn(fqdn)
	zoneName := dns01.UnFqdn(authZone)

	if name == zoneName {
		return "@"
	}

	// Remove the zone suffix from the FQDN to get the record name
	if len(name) > len(zoneName) && name[len(name)-len(zoneName)-1:] == "."+zoneName {
		return name[:len(name)-len(zoneName)-1]
	}

	return name
}
