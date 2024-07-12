// Package oraclecloud implements a DNS provider for solving the DNS-01 challenge using Oracle Cloud DNS.
package oraclecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/dns"
)

// Environment variables names.
const (
	envNamespace = "OCI_"

	EnvCompartmentOCID   = envNamespace + "COMPARTMENT_OCID"
	envPrivKey           = envNamespace + "PRIVKEY"
	EnvPrivKeyFile       = envPrivKey + "_FILE"
	EnvPrivKeyPass       = envPrivKey + "_PASS"
	EnvTenancyOCID       = envNamespace + "TENANCY_OCID"
	EnvUserOCID          = envNamespace + "USER_OCID"
	EnvPubKeyFingerprint = envNamespace + "PUBKEY_FINGERPRINT"
	EnvRegion            = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	CompartmentID      string
	OCIConfigProvider  common.ConfigurationProvider
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 60*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *dns.DnsClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(envPrivKey, EnvTenancyOCID, EnvUserOCID, EnvPubKeyFingerprint, EnvRegion, EnvCompartmentOCID)
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %w", err)
	}

	config := NewDefaultConfig()
	config.CompartmentID = values[EnvCompartmentOCID]
	config.OCIConfigProvider = newConfigProvider(values)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OracleCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("oraclecloud: the configuration of the DNS provider is nil")
	}

	if config.CompartmentID == "" {
		return nil, errors.New("oraclecloud: CompartmentID is missing")
	}

	if config.OCIConfigProvider == nil {
		return nil, errors.New("oraclecloud: OCIConfigProvider is missing")
	}

	client, err := dns.NewDnsClientWithConfigurationProvider(config.OCIConfigProvider)
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: &client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneNameOrID, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("oraclecloud: could not find zone for domain %q: %w", domain, err)
	}

	// generate request to dns.PatchDomainRecordsRequest
	recordOperation := dns.RecordOperation{
		Domain:      common.String(dns01.UnFqdn(info.EffectiveFQDN)),
		Rdata:       common.String(info.Value),
		Rtype:       common.String("TXT"),
		Ttl:         common.Int(d.config.TTL),
		IsProtected: common.Bool(false),
	}

	request := dns.PatchDomainRecordsRequest{
		CompartmentId: common.String(d.config.CompartmentID),
		ZoneNameOrId:  common.String(zoneNameOrID),
		Domain:        common.String(dns01.UnFqdn(info.EffectiveFQDN)),
		PatchDomainRecordsDetails: dns.PatchDomainRecordsDetails{
			Items: []dns.RecordOperation{recordOperation},
		},
	}

	_, err = d.client.PatchDomainRecords(context.Background(), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneNameOrID, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("oraclecloud: could not find zone for domain %q: %w", domain, err)
	}

	// search to TXT record's hash to delete
	getRequest := dns.GetDomainRecordsRequest{
		ZoneNameOrId:  common.String(zoneNameOrID),
		Domain:        common.String(dns01.UnFqdn(info.EffectiveFQDN)),
		CompartmentId: common.String(d.config.CompartmentID),
		Rtype:         common.String("TXT"),
	}

	ctx := context.Background()

	domainRecords, err := d.client.GetDomainRecords(ctx, getRequest)
	if err != nil {
		return fmt.Errorf("oraclecloud: %w", err)
	}

	if *domainRecords.OpcTotalItems == 0 {
		return errors.New("oraclecloud: no record to clean up")
	}

	var deleteHash *string
	for _, record := range domainRecords.RecordCollection.Items {
		if record.Rdata != nil && *record.Rdata == `"`+info.Value+`"` {
			deleteHash = record.RecordHash
			break
		}
	}

	if deleteHash == nil {
		return errors.New("oraclecloud: no record to clean up")
	}

	recordOperation := dns.RecordOperation{
		RecordHash: deleteHash,
		Operation:  dns.RecordOperationOperationRemove,
	}

	patchRequest := dns.PatchDomainRecordsRequest{
		ZoneNameOrId: common.String(zoneNameOrID),
		Domain:       common.String(dns01.UnFqdn(info.EffectiveFQDN)),
		PatchDomainRecordsDetails: dns.PatchDomainRecordsDetails{
			Items: []dns.RecordOperation{recordOperation},
		},
		CompartmentId: common.String(d.config.CompartmentID),
	}

	_, err = d.client.PatchDomainRecords(ctx, patchRequest)
	if err != nil {
		return fmt.Errorf("oraclecloud: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
