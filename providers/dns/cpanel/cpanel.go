// Package cpanel implements a DNS provider for solving the DNS-01 challenge using CPanel.
package cpanel

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal"
)

// Environment variables names.
const (
	envNamespace = "CPANEL_"

	EnvUsername   = envNamespace + "USERNAME"
	EnvToken      = envNamespace + "TOKEN"
	EnvBaseURL    = envNamespace + "BASE_URL"
	EnvNameserver = envNamespace + "NAMESERVER"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Username           string
	Token              string
	BaseURL            string
	Nameserver         string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config    *Config
	client    *internal.Client
	dnsClient *internal.DNSClient
}

// NewDNSProvider returns a DNSProvider instance configured for CPanel.
// Credentials must be passed in the environment variables:
// CPANEL_USERNAME, CPANEL_TOKEN, CPANEL_BASE_URL, CPANEL_NAMESERVER.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvToken, EnvBaseURL, EnvNameserver)
	if err != nil {
		return nil, fmt.Errorf("cpanel: %w", err)
	}

	config := NewDefaultConfig()
	config.Username = env.GetOrDefaultString(EnvUsername, "gr8")
	config.Token = values[EnvToken]
	config.BaseURL = values[EnvBaseURL]
	config.Nameserver = values[EnvNameserver]

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

	if config.BaseURL == "" || config.Nameserver == "" {
		return nil, errors.New("cpanel: server information are missing")
	}

	client, err := internal.NewClient(config.BaseURL, config.Username, config.Token)
	if err != nil {
		return nil, fmt.Errorf("cpanel: failed to create client: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{
		config:    config,
		client:    client,
		dnsClient: internal.NewDNSClient(10 * time.Second),
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

	soa, err := d.dnsClient.SOACall(strings.TrimPrefix(info.EffectiveFQDN, "_acme-challenge."), d.config.Nameserver)
	if err != nil {
		return fmt.Errorf("cpanel: could not find SOA for domain %q (%s) in %s: %w", domain, info.EffectiveFQDN, d.config.Nameserver, err)
	}

	zoneInfo, err := d.client.FetchZoneInformation(ctx, dns01.UnFqdn(soa.Hdr.Name))
	if err != nil {
		return fmt.Errorf("cpanel: fetch zone information: %w", err)
	}

	valueB64 := base64.StdEncoding.EncodeToString([]byte(info.Value))

	var found bool
	var existingRecord internal.ZoneRecord
	for _, record := range zoneInfo {
		if contains(record.DataB64, valueB64) {
			existingRecord = record
			found = true
			break
		}
	}

	record := internal.Record{
		DName:      info.EffectiveFQDN,
		TTL:        d.config.TTL,
		RecordType: "TXT",
	}

	// New record.
	if !found {
		record.Data = []string{info.Value}

		_, err = d.client.AddRecord(ctx, soa.Serial, soa.Hdr.Name, record)
		if err != nil {
			return fmt.Errorf("cpanel: add record: %w", err)
		}

		return nil
	}

	// Update existing record.
	record.LineIndex = existingRecord.LineIndex

	for _, dataB64 := range existingRecord.DataB64 {
		data, errD := base64.StdEncoding.DecodeString(dataB64)
		if errD != nil {
			return fmt.Errorf("cpanel: decode base64 record value: %w", errD)
		}

		record.Data = append(record.Data, string(data))
	}

	record.Data = append(record.Data, info.Value)

	_, err = d.client.EditRecord(ctx, soa.Serial, soa.Hdr.Name, record)
	if err != nil {
		return fmt.Errorf("cpanel: edit record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	soa, err := d.dnsClient.SOACall(info.EffectiveFQDN, d.config.Nameserver)
	if err != nil {
		return fmt.Errorf("cpanel: could not find SOA for domain %q (%s) in %s: %w", domain, info.EffectiveFQDN, d.config.Nameserver, err)
	}

	zoneInfo, err := d.client.FetchZoneInformation(ctx, dns01.UnFqdn(soa.Hdr.Name))
	if err != nil {
		return fmt.Errorf("cpanel: fetch zone information: %w", err)
	}

	valueB64 := base64.StdEncoding.EncodeToString([]byte(info.Value))

	var found bool
	var existingRecord internal.ZoneRecord
	for _, record := range zoneInfo {
		if contains(record.DataB64, valueB64) {
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
			return fmt.Errorf("cpanel: decode base64 record value: %w", errD)
		}

		newData = append(newData, string(data))
	}

	// Delete record.
	if len(newData) == 0 {
		err = d.client.DeleteRecord(ctx, soa.Serial, soa.Hdr.Name, existingRecord.LineIndex)
		if err != nil {
			return fmt.Errorf("cpanel: delete record: %w", err)
		}

		return nil
	}

	// Remove one value.
	record := internal.Record{
		DName:      info.EffectiveFQDN,
		TTL:        d.config.TTL,
		RecordType: "TXT",
		Data:       newData,
		LineIndex:  existingRecord.LineIndex,
	}

	_, err = d.client.EditRecord(ctx, soa.Serial, soa.Hdr.Name, record)
	if err != nil {
		return fmt.Errorf("cpanel: edit record: %w", err)
	}

	return nil
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
