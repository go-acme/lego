package hyperone

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/go-acme/lego/v3/providers/dns/hyperone/internal"
)

const (
	envNamespace = "HYPERONE_"

	envPassportLocation   = envNamespace + "PASSPORT_LOCATION"
	envTTL                = envNamespace + "TTL"
	envAPIUrl             = envNamespace + "API_URL"
	envLocationID         = envNamespace + "LOCATION_ID"
	envPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	envPollingInterval    = envNamespace + "POLLING_INTERVAL"
	envZoneURI            = envNamespace + "ZONE_URI"
)

type Config struct {
	TTL                int
	APIEndpoint        string
	LocationID         string
	PassportLocation   string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	ZoneURI            string
}

type DNSProvider struct {
	client   *Client
	config   *Config
	passport *internal.Passport
}

func NewDefaultConfig() *Config {
	return &Config{
		PassportLocation:   env.GetOrDefaultString(envPassportLocation, internal.GetDefaultPassportLocation()),
		TTL:                env.GetOrDefaultInt(envTTL, dns01.DefaultTTL),
		APIEndpoint:        env.GetOrDefaultString(envAPIUrl, "https://api.hyperone.com/v2"),
		LocationID:         env.GetOrDefaultString(envLocationID, "pl-waw-1"),
		PropagationTimeout: env.GetOrDefaultSecond(envPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(envPollingInterval, dns01.DefaultPollingInterval),
		ZoneURI:            env.GetOrFile(envZoneURI),
	}
}

// NewDNSProvider creates struct for handling DNS challenge.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	if config.PassportLocation == "" {
		return nil, errors.New("You must provide passport location")
	}

	passport, err := internal.LoadPassportFile(config.PassportLocation)
	if err != nil {
		return nil, err
	}
	tokenSigner := &internal.TokenSigner{PrivateKey: passport.PrivateKey, KeyID: passport.CertificateID, Audience: config.APIEndpoint, Issuer: passport.Issuer, Subject: passport.SubjectID}
	client := &Client{Signer: tokenSigner}

	provider := DNSProvider{client: client, config: config, passport: passport}

	return &provider, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneURI, err := d.getHostedZoneURI(fqdn)
	if err != nil {
		return err
	}

	d.client.ZoneFullURI = d.getFullURI(zoneURI)
	recordset, err := d.client.findRecordset("TXT", fqdn)
	if err != nil {
		return err
	}

	if recordset == nil {
		_, err = d.client.createRecordsetWithRecord("TXT", fqdn, value, d.config.TTL)
		return err
	}

	_, err = d.client.setRecord(recordset.ID, value)
	return err
}

// CleanUp removes the TXT record matching the specified parameters
// and recordset if no other records are remaining.
// There is a small possibility that race will cause to delete
// recordset with records for other DNS Challenges.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneURI, err := d.getHostedZoneURI(fqdn)
	if err != nil {
		return err
	}
	d.client.ZoneFullURI = d.getFullURI(zoneURI)

	recordset, err := d.client.findRecordset("TXT", fqdn)
	if err != nil {
		return err
	}
	if recordset == nil {
		return errors.New("Can't find the recordset to remove")
	}

	records, err := d.client.getRecords(recordset.ID)
	if err != nil {
		return err
	}

	if len(records) == 1 {
		if records[0].Content == value {
			return d.client.deleteRecordset(recordset.ID)
		}
		return errors.New("Record with given content not found")
	}

	for _, record := range records {
		if record.Content == value {
			return d.client.deleteRecord(recordset.ID, record.ID)
		}
	}

	return errors.New("Failed to find record with given value")
}

// getHostedZoneURI returns ZoneURI from environment variable (if set)
// or tries to find it in current project.
func (d *DNSProvider) getHostedZoneURI(fqdn string) (string, error) {
	if d.config.ZoneURI != "" {
		return d.config.ZoneURI, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	projectID, err := d.passport.ExtractProjectID()
	if err != nil {
		return "", err
	}

	zoneURI, err := d.client.findZone(authZone, projectID, d.config.LocationID, d.config.APIEndpoint)
	if err != nil {
		return "", fmt.Errorf("Error when finding zoneID:%+v", err)
	}
	return zoneURI, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Takes zoneURI and composes it with APIEndpoint.
func (d *DNSProvider) getFullURI(zoneURI string) string {
	return d.config.APIEndpoint + zoneURI
}
