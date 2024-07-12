// Package cpanel implements a DNS provider for solving the DNS-01 challenge using CPanel.
package cpanel

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/cpanel"
	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/whm"
)

// Environment variables names.
const (
	envNamespace = "CPANEL_"

	EnvMode     = envNamespace + "MODE"
	EnvUsername = envNamespace + "USERNAME"
	EnvToken    = envNamespace + "TOKEN"
	EnvBaseURL  = envNamespace + "BASE_URL"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

type apiClient interface {
	FetchZoneInformation(ctx context.Context, domain string) ([]shared.ZoneRecord, error)
	AddRecord(ctx context.Context, serial uint32, domain string, record shared.Record) (*shared.ZoneSerial, error)
	EditRecord(ctx context.Context, serial uint32, domain string, record shared.Record) (*shared.ZoneSerial, error)
	DeleteRecord(ctx context.Context, serial uint32, domain string, lineIndex int) (*shared.ZoneSerial, error)
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Mode               string
	Username           string
	Token              string
	BaseURL            string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		Mode:               env.GetOrDefaultString(EnvMode, "cpanel"),
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client apiClient
}

// NewDNSProvider returns a DNSProvider instance configured for CPanel.
// Credentials must be passed in the environment variables:
// CPANEL_USERNAME, CPANEL_TOKEN, CPANEL_BASE_URL, CPANEL_NAMESERVER.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvUsername, EnvToken, EnvBaseURL)
	if err != nil {
		return nil, fmt.Errorf("cpanel: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = values[EnvUsername]
	config.Token = values[EnvToken]
	config.BaseURL = values[EnvBaseURL]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for CPanel.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("cpanel: the configuration of the DNS provider is nil")
	}

	if config.Username == "" || config.Token == "" {
		return nil, errors.New("cpanel: some credentials information are missing")
	}

	if config.BaseURL == "" {
		return nil, errors.New("cpanel: server information are missing")
	}

	client, err := createClient(config)
	if err != nil {
		return nil, fmt.Errorf("cpanel: create client error: %w", err)
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
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("arvancloud: could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)

	zoneInfo, err := d.client.FetchZoneInformation(ctx, zone)
	if err != nil {
		return fmt.Errorf("cpanel[mode=%s]: fetch zone information: %w", d.config.Mode, err)
	}

	serial, err := getZoneSerial(authZone, zoneInfo)
	if err != nil {
		return fmt.Errorf("cpanel[mode=%s]: get zone serial: %w", d.config.Mode, err)
	}

	valueB64 := base64.StdEncoding.EncodeToString([]byte(info.Value))

	var found bool
	var existingRecord shared.ZoneRecord
	for _, record := range zoneInfo {
		if slices.Contains(record.DataB64, valueB64) {
			existingRecord = record
			found = true
			break
		}
	}

	record := shared.Record{
		DName:      info.EffectiveFQDN,
		TTL:        d.config.TTL,
		RecordType: "TXT",
	}

	// New record.
	if !found {
		record.Data = []string{info.Value}

		_, err = d.client.AddRecord(ctx, serial, zone, record)
		if err != nil {
			return fmt.Errorf("cpanel[mode=%s]: add record: %w", d.config.Mode, err)
		}

		return nil
	}

	// Update existing record.
	record.LineIndex = existingRecord.LineIndex

	for _, dataB64 := range existingRecord.DataB64 {
		data, errD := base64.StdEncoding.DecodeString(dataB64)
		if errD != nil {
			return fmt.Errorf("cpanel[mode=%s]: decode base64 record value: %w", d.config.Mode, errD)
		}

		record.Data = append(record.Data, string(data))
	}

	record.Data = append(record.Data, info.Value)

	_, err = d.client.EditRecord(ctx, serial, zone, record)
	if err != nil {
		return fmt.Errorf("cpanel[mode=%s]: edit record: %w", d.config.Mode, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("arvancloud: could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)

	zoneInfo, err := d.client.FetchZoneInformation(ctx, zone)
	if err != nil {
		return fmt.Errorf("cpanel[mode=%s]: fetch zone information: %w", d.config.Mode, err)
	}

	serial, err := getZoneSerial(authZone, zoneInfo)
	if err != nil {
		return fmt.Errorf("cpanel[mode=%s]: get zone serial: %w", d.config.Mode, err)
	}

	valueB64 := base64.StdEncoding.EncodeToString([]byte(info.Value))

	var found bool
	var existingRecord shared.ZoneRecord
	for _, record := range zoneInfo {
		if slices.Contains(record.DataB64, valueB64) {
			existingRecord = record
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	var newData []string
	for _, dataB64 := range existingRecord.DataB64 {
		if dataB64 == valueB64 {
			continue
		}

		data, errD := base64.StdEncoding.DecodeString(dataB64)
		if errD != nil {
			return fmt.Errorf("cpanel[mode=%s]: decode base64 record value: %w", d.config.Mode, errD)
		}

		newData = append(newData, string(data))
	}

	// Delete record.
	if len(newData) == 0 {
		_, err = d.client.DeleteRecord(ctx, serial, zone, existingRecord.LineIndex)
		if err != nil {
			return fmt.Errorf("cpanel[mode=%s]: delete record: %w", d.config.Mode, err)
		}

		return nil
	}

	// Remove one value.
	record := shared.Record{
		DName:      info.EffectiveFQDN,
		TTL:        d.config.TTL,
		RecordType: "TXT",
		Data:       newData,
		LineIndex:  existingRecord.LineIndex,
	}

	_, err = d.client.EditRecord(ctx, serial, zone, record)
	if err != nil {
		return fmt.Errorf("cpanel[mode=%s]: edit record: %w", d.config.Mode, err)
	}

	return nil
}

func getZoneSerial(zoneFqdn string, zoneInfo []shared.ZoneRecord) (uint32, error) {
	nameB64 := base64.StdEncoding.EncodeToString([]byte(zoneFqdn))

	for _, record := range zoneInfo {
		if record.Type != "record" || record.RecordType != "SOA" || record.DNameB64 != nameB64 {
			continue
		}

		// https://github.com/go-acme/lego/issues/1060#issuecomment-1925572386
		// https://github.com/go-acme/lego/issues/1060#issuecomment-1925581832
		data, err := base64.StdEncoding.DecodeString(record.DataB64[2])
		if err != nil {
			return 0, fmt.Errorf("decode serial DNameB64: %w", err)
		}

		var newSerial uint32
		_, err = fmt.Sscan(string(data), &newSerial)
		if err != nil {
			return 0, fmt.Errorf("decode serial DNameB64, invalid serial value %q: %w", string(data), err)
		}

		return newSerial, nil
	}

	return 0, errors.New("zone serial not found")
}

func createClient(config *Config) (apiClient, error) {
	switch strings.ToLower(config.Mode) {
	case "cpanel":
		client, err := cpanel.NewClient(config.BaseURL, config.Username, config.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create cPanel API client: %w", err)
		}

		if config.HTTPClient != nil {
			client.HTTPClient = config.HTTPClient
		}

		return client, nil

	case "whm":
		client, err := whm.NewClient(config.BaseURL, config.Username, config.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create WHM API client: %w", err)
		}

		if config.HTTPClient != nil {
			client.HTTPClient = config.HTTPClient
		}

		return client, nil

	default:
		return nil, fmt.Errorf("unsupported mode: %q", config.Mode)
	}
}
