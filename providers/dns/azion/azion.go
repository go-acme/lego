// Package azion implements a DNS provider for solving the DNS-01 challenge using Azion Edge DNS.
package azion

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aziontech/azionapi-go-sdk/idns"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "AZION_"

	EnvPersonalToken = envNamespace + "PERSONAL_TOKEN"
	EnvPageSize      = envNamespace + "PAGE_SIZE"

	EnvTTL                = envNamespace + "TTL"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	PersonalToken string
	PageSize      int

	PollingInterval    time.Duration
	PropagationTimeout time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PageSize:           env.GetOrDefaultInt(EnvPageSize, 50),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *idns.APIClient

	recordIDs   map[string]int32
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Azion.
// Credentials must be passed in the environment variable: AZION_PERSONAL_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPersonalToken)
	if err != nil {
		return nil, fmt.Errorf("azion: %w", err)
	}

	config := NewDefaultConfig()
	config.PersonalToken = values[EnvPersonalToken]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azion.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azion: the configuration of the DNS provider is nil")
	}

	if config.PersonalToken == "" {
		return nil, errors.New("azion: missing credentials")
	}

	clientConfig := idns.NewConfiguration()
	clientConfig.AddDefaultHeader("Accept", "application/json; version=3")
	clientConfig.UserAgent = "lego-dns/azion"

	if config.HTTPClient != nil {
		clientConfig.HTTPClient = config.HTTPClient
	}

	clientConfig.HTTPClient = clientdebug.Wrap(clientConfig.HTTPClient)

	client := idns.NewAPIClient(clientConfig)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]int32),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctxAuth := authContext(context.Background(), d.config.PersonalToken)

	zone, err := d.findZone(ctxAuth, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azion: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := extractSubDomain(info, zone)
	if err != nil {
		return fmt.Errorf("azion: %w", err)
	}

	// Check if a TXT record with the same name already exists
	existingRecord, err := d.findExistingTXTRecord(ctxAuth, zone.GetId(), subDomain)
	if err != nil {
		return fmt.Errorf("azion: check existing records: %w", err)
	}

	record := idns.NewRecordPostOrPut()
	record.SetEntry(subDomain)
	record.SetRecordType("TXT")
	record.SetTtl(int32(d.config.TTL))

	var resp *idns.PostOrPutRecordResponse
	if existingRecord != nil {
		// Update existing record by adding the new value to the existing ones
		record.SetAnswersList(append(existingRecord.GetAnswersList(), info.Value))

		// Use PUT to update the existing record
		resp, _, err = d.client.RecordsAPI.PutZoneRecord(ctxAuth, zone.GetId(), existingRecord.GetRecordId()).RecordPostOrPut(*record).Execute()
		if err != nil {
			return fmt.Errorf("azion: update existing record: %w", err)
		}
	} else {
		// Create a new record
		record.SetAnswersList([]string{info.Value})

		resp, _, err = d.client.RecordsAPI.PostZoneRecord(ctxAuth, zone.GetId()).RecordPostOrPut(*record).Execute()
		if err != nil {
			return fmt.Errorf("azion: create new zone record: %w", err)
		}
	}

	if resp == nil || resp.Results == nil {
		return errors.New("azion: create zone record error")
	}

	results := resp.GetResults()
	d.recordIDsMu.Lock()
	d.recordIDs[token] = results.GetId()
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	ctxAuth := authContext(context.Background(), d.config.PersonalToken)

	zone, err := d.findZone(ctxAuth, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azion: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := extractSubDomain(info, zone)
	if err != nil {
		return fmt.Errorf("azion: %w", err)
	}

	defer func() {
		// Cleans the record ID.
		d.recordIDsMu.Lock()
		delete(d.recordIDs, token)
		d.recordIDsMu.Unlock()
	}()

	existingRecord, err := d.findExistingTXTRecord(ctxAuth, zone.GetId(), subDomain)
	if err != nil {
		return fmt.Errorf("azion: find existing record: %w", err)
	}

	if existingRecord == nil {
		return nil
	}

	currentAnswers := existingRecord.GetAnswersList()

	var updatedAnswers []string
	for _, answer := range currentAnswers {
		if answer != info.Value {
			updatedAnswers = append(updatedAnswers, answer)
		}
	}

	// If no answers remain, delete the entire record
	if len(updatedAnswers) == 0 {
		_, resp, errDelete := d.client.RecordsAPI.DeleteZoneRecord(ctxAuth, zone.GetId(), existingRecord.GetRecordId()).Execute()
		if errDelete != nil {
			// If a record doesn't exist (404), consider cleanup successful
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return nil
			}

			return fmt.Errorf("azion: delete record: %w", errDelete)
		}

		return nil
	}

	// Update the record with remaining answers
	record := idns.NewRecordPostOrPut()
	record.SetEntry(subDomain)
	record.SetRecordType("TXT")
	record.SetAnswersList(updatedAnswers)
	record.SetTtl(existingRecord.GetTtl())

	_, _, err = d.client.RecordsAPI.PutZoneRecord(ctxAuth, zone.GetId(), existingRecord.GetRecordId()).RecordPostOrPut(*record).Execute()
	if err != nil {
		return fmt.Errorf("azion: update record: %w", err)
	}

	return nil
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (*idns.Zone, error) {
	resp, _, err := d.client.ZonesAPI.GetZones(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("get zones: %w", err)
	}

	if resp == nil {
		return nil, errors.New("get zones: no results")
	}

	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		for _, zone := range resp.GetResults() {
			if zone.GetDomain() == domain {
				return &zone, nil
			}
		}
	}

	return nil, fmt.Errorf("zone not found (fqdn: %q)", fqdn)
}

// findExistingTXTRecord searches for an existing TXT record with the given name in the specified zone.
// It handles pagination to search through all pages of results.
func (d *DNSProvider) findExistingTXTRecord(ctx context.Context, zoneID int32, recordName string) (*idns.RecordGet, error) {
	var page int64 = 1

	for {
		resp, _, err := d.client.RecordsAPI.GetZoneRecords(ctx, zoneID).Page(page).PageSize(int64(d.config.PageSize)).Execute()
		if err != nil {
			return nil, fmt.Errorf("get zone records (page %d): %w", page, err)
		}

		if resp == nil {
			return nil, errors.New("get zone records: no results")
		}

		results, ok := resp.GetResultsOk()
		if !ok || results == nil {
			return nil, errors.New("get zone records: empty")
		}

		// Search for existing TXT record with the same name in current page
		for _, record := range results.GetRecords() {
			if record.GetRecordType() == "TXT" && record.GetEntry() == recordName {
				return &record, nil
			}
		}

		// Check if there are more pages to search
		if page >= int64(resp.GetTotalPages()) {
			break
		}

		page++
	}

	// No existing record found in any page
	return nil, nil
}

func authContext(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idns.ContextAPIKeys, map[string]idns.APIKey{
		"tokenAuth": {
			Key:    key,
			Prefix: "Token",
		},
	})
}

func extractSubDomain(info dns01.ChallengeInfo, zone *idns.Zone) (string, error) {
	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.GetName())
	if err != nil {
		return "", err
	}

	if subDomain != "" {
		return subDomain, nil
	}

	return "@", nil
}
