// Package vinyldns implements a DNS provider for solving the DNS-01 challenge using VinylDNS.
package vinyldns

import (
	"errors"
	"fmt"
	"strings"
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
const vinylDNSSuccessStatus = "Complete"

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKey          string
	SecretKey          string
	Host               string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	Client             *vinyldns.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		AccessKey:          env.GetOrFile(EnvAccessKey),
		SecretKey:          env.GetOrFile(EnvSecretKey),
		Host:               env.GetOrFile(EnvHost),
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

// NewDNSProvider returns a DNSProvider instance configured for the VinylDNS service.
//
// VinylDNS Credentials are automatically detected in the following locations and prioritized in the following order:
// 1. Environment variables: VINYLDNS_ACCESS_KEY, VINYLDNS_SECRET_KEY, VINYLDNS_HOST
// 2. Environment variables of file paths to files containing the value only: VINYLDNS_ACCESS_KEY_FILE, VINYLDNS_SECRET_KEY_FILE, VINYLDNS_HOST_FILE.
//
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderConfig takes a given config ans returns a custom configured DNSProvider instance.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vinyldns: the configuration of the VinylDNS DNS provider is nil")
	}

	if config.Client != nil {
		return &DNSProvider{client: config.Client, config: config}, nil
	}

	cc := vinyldns.ClientConfiguration{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Host:      config.Host,
		UserAgent: "go-acme/lego",
	}

	return &DNSProvider{client: vinyldns.NewClient(cc), config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	existingRecord, err := d.getExistingRecordSet(fqdn)
	if err != nil {
		return err
	}

	record := vinyldns.Record{
		Text: value,
	}

	if existingRecord.ID == "" {
		err = d.createRecordSet(fqdn, []vinyldns.Record{record})
		if err != nil {
			return fmt.Errorf("vinyldns: %w", err)
		}
		return nil
	}

	records := existingRecord.Records

	var found bool
	for _, i := range records {
		if i.Text == value {
			found = true
		}
	}

	if !found {
		records = append(records, record)
		err = d.updateRecordSet(existingRecord, records)
		if err != nil {
			return fmt.Errorf("vinyldns: %w", err)
		}
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	existingRecord, err := d.getExistingRecordSet(fqdn)
	if err != nil {
		return err
	}

	if existingRecord.ID == "" || len(existingRecord.Records) == 0 {
		return nil
	}

	records := []vinyldns.Record{}
	for _, i := range existingRecord.Records {
		if i.Text != value {
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

func (d *DNSProvider) createRecordSet(fqdn string, records []vinyldns.Record) error {
	hostName, domainName, err := d.fqdnSplit(fqdn)
	if err != nil {
		return err
	}
	zone, err := d.getZone(domainName)
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
	return wait.For("vinyldns", d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
		return d.waitForPropagation("CreateRS", resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
	})
}

func (d *DNSProvider) deleteRecordSet(existingRecord *vinyldns.RecordSet) error {
	resp, err := d.client.RecordSetDelete(existingRecord.ZoneID, existingRecord.ID)
	if err != nil {
		return err
	}
	return wait.For("vinyldns", d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
		return d.waitForPropagation("DeleteRS", resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
	})
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
	return wait.For("vinyldns", d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
		return d.waitForPropagation("UpdateRS - "+operation, resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
	})
}

func (d *DNSProvider) getExistingRecordSet(fqdn string) (*vinyldns.RecordSet, error) {
	hostName, domainName, err := d.fqdnSplit(fqdn)
	if err != nil {
		return nil, err
	}
	zone, err := d.getZone(domainName)
	if err != nil {
		return nil, err
	}

	filter := vinyldns.ListFilter{
		NameFilter: hostName,
	}

	recordSetsOutputRaw, err := d.client.RecordSetsListAll(zone.ID, filter)
	if err != nil {
		return nil, err
	}
	recordSetsOutput := []vinyldns.RecordSet{}
	for _, i := range recordSetsOutputRaw {
		if i.Type == "TXT" {
			recordSetsOutput = append(recordSetsOutput, i)
		}
	}

	if recordSetsOutput == nil {
		return nil, nil
	}

	var record *vinyldns.RecordSet

	switch {
	case len(recordSetsOutput) > 1:
		return nil, fmt.Errorf("ambiguous recordset definition of %s", fqdn)
	case len(recordSetsOutput) == 1:
		record = &recordSetsOutput[0]
	default:
		record = &vinyldns.RecordSet{}
	}

	return record, nil
}

func (d *DNSProvider) fqdnSplit(fqdn string) (string, string, error) {
	s := strings.Split(fqdn, ".")
	if len(s) < 3 {
		return "", "", fmt.Errorf("invalid fqdn: %s", fqdn)
	}
	hostName := strings.Join(s[:2], ".")
	domainName := strings.Join(s[2:], ".")
	return hostName, domainName, nil
}

func (d *DNSProvider) getZone(domainName string) (*vinyldns.Zone, error) {
	zone, err := d.client.ZoneByName(domainName)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (d *DNSProvider) waitForPropagation(operation, zoneID, recordsetID, changeID string) (bool, error) {
	changeStatusResp, err := d.client.RecordSetChange(zoneID, recordsetID, changeID)
	if err != nil {
		return false, fmt.Errorf("failed to query change status: %w", err)
	}
	if changeStatusResp.Status == vinylDNSSuccessStatus {
		return true, nil
	}

	return false, fmt.Errorf("waiting for propagation of change to authoritative zone servers; operation: %s, zoneID: %s, recordsetID: %s, changeID: %s", operation, zoneID, recordsetID, changeID)
}
