package oraclecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Sugi275/oci-env-configprovider/envprovider"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/dns"
	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	compartmentID      string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("OCI_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("OCI_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OCI_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("OCI_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dns.DnsClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {
	compartmentID, err := envprovider.GetCompartmentID()
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	config := NewDefaultConfig()
	config.compartmentID = compartmentID

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OracleCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("oraclecloud: the configuration of the DNS provider is nil")
	}

	if config.compartmentID == "" {
		return nil, errors.New("oraclecloud: CompartmentID is missing")
	}

	client, err := dns.NewDnsClientWithConfigurationProvider(envprovider.GetEnvConfigProvider())
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: &client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// generate request RecordDetails
	recordDetails := dns.RecordDetails{
		Domain:      common.String(dns01.UnFqdn(fqdn)),
		Rdata:       common.String(value),
		Rtype:       common.String("TXT"),
		Ttl:         common.Int(30),
		IsProtected: common.Bool(false),
	}

	request := dns.UpdateDomainRecordsRequest{
		ZoneNameOrId: common.String(domain),
		Domain:       common.String(dns01.UnFqdn(fqdn)),
		UpdateDomainRecordsDetails: dns.UpdateDomainRecordsDetails{
			Items: []dns.RecordDetails{
				recordDetails,
			},
		},
		CompartmentId: common.String(d.config.compartmentID),
	}

	_, err := d.client.UpdateDomainRecords(context.Background(), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	request := dns.DeleteDomainRecordsRequest{
		ZoneNameOrId:  common.String(domain),
		Domain:        common.String(dns01.UnFqdn(fqdn)),
		CompartmentId: common.String(d.config.compartmentID),
	}

	_, err := d.client.DeleteDomainRecords(context.Background(), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	return nil
}
