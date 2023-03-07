// Package vinyldns implements a DNS provider for solving the DNS-01 challenge using VinylDNS.
package vinyldns

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/vinyldns/go-vinyldns/vinyldns"
)

// Environment variables names.
const (
	envNamespace = "VINYLDNS_"

	EnvAccessKey = envNamespace + "ACCESS_KEY"
	EnvSecretKey = envNamespace + "SECRET_KEY"
	EnvHost      = envNamespace + "HOST"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKey          string
	SecretKey          string
	Host               string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 30),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 4*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *vinyldns.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for VinylDNS.
// Credentials must be passed in the environment variables:
// VINYLDNS_ACCESS_KEY, VINYLDNS_SECRET_KEY, VINYLDNS_HOST.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKey, EnvSecretKey, EnvHost)
	if err != nil {
		return nil, fmt.Errorf("vinyldns: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.Host = values[EnvHost]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for VinylDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vinyldns: the configuration of the VinylDNS DNS provider is nil")
	}

	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, errors.New("vinyldns: credentials are missing")
	}

	if config.Host == "" {
		return nil, errors.New("vinyldns: host is missing")
	}

	client := vinyldns.NewClient(vinyldns.ClientConfiguration{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Host:      config.Host,
		UserAgent: "go-acme/lego",
	})

	client.HTTPClient.Timeout = 30 * time.Second

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	existingRecord, err := d.getRecordSet(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	record := vinyldns.Record{Text: info.Value}

	if existingRecord == nil || existingRecord.ID == "" {
		err = d.createRecordSet(info.EffectiveFQDN, []vinyldns.Record{record})
		if err != nil {
			return fmt.Errorf("vinyldns: %w", err)
		}

		return nil
	}

	for _, i := range existingRecord.Records {
		if i.Text == info.Value {
			return nil
		}
	}

	records := existingRecord.Records
	records = append(records, record)

	err = d.updateRecordSet(existingRecord, records)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	existingRecord, err := d.getRecordSet(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	if existingRecord == nil || existingRecord.ID == "" || len(existingRecord.Records) == 0 {
		return nil
	}

	var records []vinyldns.Record
	for _, i := range existingRecord.Records {
		if i.Text != info.Value {
			records = append(records, i)
		}
	}

	if len(records) == 0 {
		err = d.deleteRecordSet(existingRecord)
		if err != nil {
			return fmt.Errorf("vinyldns: %w", err)
		}

		return nil
	}

	err = d.updateRecordSet(existingRecord, records)
	if err != nil {
		return fmt.Errorf("vinyldns: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getRecordSet(fqdn string) (*vinyldns.RecordSet, error) {
	zoneName, hostName, err := splitDomain(fqdn)
	if err != nil {
		return nil, err
	}

	zone, err := d.client.ZoneByName(zoneName)
	if err != nil {
		return nil, err
	}

	allRecordSets, err := d.client.RecordSetsListAll(zone.ID, vinyldns.ListFilter{NameFilter: hostName})
	if err != nil {
		return nil, err
	}

	var recordSets []vinyldns.RecordSet
	for _, i := range allRecordSets {
		if i.Type == "TXT" {
			recordSets = append(recordSets, i)
		}
	}

	switch {
	case len(recordSets) > 1:
		return nil, fmt.Errorf("ambiguous recordset definition of %s", fqdn)
	case len(recordSets) == 1:
		return &recordSets[0], nil
	default:
		return nil, nil
	}
}

func (d *DNSProvider) createRecordSet(fqdn string, records []vinyldns.Record) error {
	zoneName, hostName, err := splitDomain(fqdn)
	if err != nil {
		return err
	}

	zone, err := d.client.ZoneByName(zoneName)
	if err != nil {
		return err
	}

	recordSet := vinyldns.RecordSet{
		Name:    hostName,
		ZoneID:  zone.ID,
		Type:    "TXT",
		TTL:     d.config.TTL,
		Records: records,
	}

	resp, err := d.client.RecordSetCreate(&recordSet)
	if err != nil {
		return err
	}

	return d.waitForChanges("CreateRS", resp)
}

func (d *DNSProvider) updateRecordSet(recordSet *vinyldns.RecordSet, newRecords []vinyldns.Record) error {
	operation := "delete"
	if len(recordSet.Records) < len(newRecords) {
		operation = "add"
	}

	recordSet.Records = newRecords
	recordSet.TTL = d.config.TTL

	resp, err := d.client.RecordSetUpdate(recordSet)
	if err != nil {
		return err
	}

	return d.waitForChanges("UpdateRS - "+operation, resp)
}

func (d *DNSProvider) deleteRecordSet(existingRecord *vinyldns.RecordSet) error {
	resp, err := d.client.RecordSetDelete(existingRecord.ZoneID, existingRecord.ID)
	if err != nil {
		return err
	}

	return d.waitForChanges("DeleteRS", resp)
}

func (d *DNSProvider) waitForChanges(operation string, resp *vinyldns.RecordSetUpdateResponse) error {
	return wait.For("vinyldns", d.config.PropagationTimeout, d.config.PollingInterval,
		func() (bool, error) {
			change, err := d.client.RecordSetChange(resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
			if err != nil {
				return false, fmt.Errorf("failed to query change status: %w", err)
			}

			if change.Status == "Complete" {
				return true, nil
			}

			return false, fmt.Errorf("waiting operation: %s, zoneID: %s, recordsetID: %s, changeID: %s",
				operation, resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
		},
	)
}

// splitDomain splits the hostname from the authoritative zone, and returns both parts.
func splitDomain(fqdn string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", "", err
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return "", "", err
	}

	return zone, subDomain, nil
}
