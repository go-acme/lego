// Package designate implements a DNS provider for solving the DNS-01 challenge using the Designate DNSaaS for Openstack.
package designate

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

// Environment variables names.
const (
	envNamespace = "DESIGNATE_"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	envNamespaceClient = "OS_"

	EnvAuthURL       = envNamespaceClient + "AUTH_URL"
	EnvUsername      = envNamespaceClient + "USERNAME"
	EnvPassword      = envNamespaceClient + "PASSWORD"
	EnvUserID        = envNamespaceClient + "USER_ID"
	EnvAppCredID     = envNamespaceClient + "APPLICATION_CREDENTIAL_ID"
	EnvAppCredName   = envNamespaceClient + "APPLICATION_CREDENTIAL_NAME"
	EnvAppCredSecret = envNamespaceClient + "APPLICATION_CREDENTIAL_SECRET"
	EnvTenantName    = envNamespaceClient + "TENANT_NAME"
	EnvRegionName    = envNamespaceClient + "REGION_NAME"
	EnvProjectID     = envNamespaceClient + "PROJECT_ID"
	EnvCloud         = envNamespaceClient + "CLOUD"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	opts               gophercloud.AuthOptions
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 10),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config       *Config
	client       *gophercloud.ServiceClient
	dnsEntriesMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Designate.
// Credentials must be passed in the environment variables:
// OS_AUTH_URL, OS_USERNAME, OS_PASSWORD, OS_REGION_NAME.
// Or you can specify OS_CLOUD to read the credentials from the according cloud entry.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	val, err := env.Get(EnvCloud)
	if err == nil {
		opts, erro := clientconfig.AuthOptions(&clientconfig.ClientOpts{
			Cloud: val[EnvCloud],
		})

		if erro != nil {
			return nil, fmt.Errorf("designate: %w", erro)
		}

		config.opts = *opts
	} else {
		opts, err := openstack.AuthOptionsFromEnv()
		if err != nil {
			return nil, fmt.Errorf("designate: %w", err)
		}

		config.opts = opts
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Designate.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("designate: the configuration of the DNS provider is nil")
	}

	provider, err := openstack.AuthenticatedClient(config.opts)
	if err != nil {
		return nil, fmt.Errorf("designate: failed to authenticate: %w", err)
	}

	dnsClient, err := openstack.NewDNSV2(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
	if err != nil {
		return nil, fmt.Errorf("designate: failed to get DNS provider: %w", err)
	}

	return &DNSProvider{client: dnsClient, config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("designate: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("designate: couldn't get zone ID in Present: %w", err)
	}

	// use mutex to prevent race condition between creating the record and verifying it
	d.dnsEntriesMu.Lock()
	defer d.dnsEntriesMu.Unlock()

	existingRecord, err := d.getRecord(zoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("designate: %w", err)
	}

	if existingRecord != nil {
		if contains(existingRecord.Records, info.Value) {
			log.Printf("designate: the record already exists: %s", info.Value)
			return nil
		}

		return d.updateRecord(existingRecord, info.Value)
	}

	err = d.createRecord(zoneID, info.EffectiveFQDN, info.Value)
	if err != nil {
		return fmt.Errorf("designate: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("designate: could not find zone for domain %q (%s): %w", domain, info.EffectiveFQDN, err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("designate: couldn't get zone ID in CleanUp: %w", err)
	}

	// use mutex to prevent race condition between getting the record and deleting it
	d.dnsEntriesMu.Lock()
	defer d.dnsEntriesMu.Unlock()

	record, err := d.getRecord(zoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("designate: couldn't get Record ID in CleanUp: %w", err)
	}

	if record == nil {
		// Record is already deleted
		return nil
	}

	err = recordsets.Delete(d.client, zoneID, record.ID).ExtractErr()
	if err != nil {
		return fmt.Errorf("designate: error for %s in CleanUp: %w", info.EffectiveFQDN, err)
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

func (d *DNSProvider) createRecord(zoneID, fqdn, value string) error {
	createOpts := recordsets.CreateOpts{
		Name:        fqdn,
		Type:        "TXT",
		TTL:         d.config.TTL,
		Description: "ACME verification record",
		Records:     []string{value},
	}

	actual, err := recordsets.Create(d.client, zoneID, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("error for %s in Present while creating record: %w", fqdn, err)
	}

	if actual.Name != fqdn || actual.TTL != d.config.TTL {
		return errors.New("the created record doesn't match what we wanted to create")
	}

	return nil
}

func (d *DNSProvider) updateRecord(record *recordsets.RecordSet, value string) error {
	if contains(record.Records, value) {
		log.Printf("skip: the record already exists: %s", value)
		return nil
	}

	values := append([]string{value}, record.Records...)

	updateOpts := recordsets.UpdateOpts{
		Description: &record.Description,
		TTL:         &record.TTL,
		Records:     values,
	}

	result := recordsets.Update(d.client, record.ZoneID, record.ID, updateOpts)
	return result.Err
}

func (d *DNSProvider) getZoneID(wanted string) (string, error) {
	allPages, err := zones.List(d.client, nil).AllPages()
	if err != nil {
		return "", err
	}
	allZones, err := zones.ExtractZones(allPages)
	if err != nil {
		return "", err
	}

	for _, zone := range allZones {
		if zone.Name == wanted {
			return zone.ID, nil
		}
	}
	return "", fmt.Errorf("zone id not found for %s", wanted)
}

func (d *DNSProvider) getRecord(zoneID, wanted string) (*recordsets.RecordSet, error) {
	allPages, err := recordsets.ListByZone(d.client, zoneID, nil).AllPages()
	if err != nil {
		return nil, err
	}
	allRecords, err := recordsets.ExtractRecordSets(allPages)
	if err != nil {
		return nil, err
	}

	for _, record := range allRecords {
		if record.Name == wanted && record.Type == "TXT" {
			return &record, nil
		}
	}

	return nil, nil
}
