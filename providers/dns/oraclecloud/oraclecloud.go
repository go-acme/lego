package oraclecloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sugi275/oci-env-configprovider/envprovider"
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
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("OCI_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("OCI_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OCI_POLLING_INTERVAL", dns01.DefaultPollingInterval),
	}

}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dns.DnsClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {

	compartmentid, err := envprovider.GetCompartmentID()
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	config := NewDefaultConfig()
	config.compartmentID = compartmentid

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

	return &DNSProvider{client: &client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	fqdn = DeleteLastDot(fqdn)

	// generate request RecordDetails
	txttype := "TXT"
	falseFlg := false
	ttl := 30

	recordDetails := dns.RecordDetails{
		Domain:      &fqdn,
		Rdata:       &value,
		Rtype:       &txttype,
		Ttl:         &ttl,
		IsProtected: &falseFlg,
	}

	var recordDetailsList []dns.RecordDetails
	recordDetailsList = append(recordDetailsList, recordDetails)

	updateDomainRecordsDetails := dns.UpdateDomainRecordsDetails{
		Items: recordDetailsList,
	}

	request := dns.UpdateDomainRecordsRequest{
		ZoneNameOrId:               &domain,
		Domain:                     &fqdn,
		UpdateDomainRecordsDetails: updateDomainRecordsDetails,
		CompartmentId:              &d.config.compartmentID,
	}

	ctx := context.Background()
	_, err := d.client.UpdateDomainRecords(ctx, request)
	if err != nil {
		panic(err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	fqdn = DeleteLastDot(fqdn)

	request := dns.DeleteDomainRecordsRequest{
		ZoneNameOrId:  &domain,
		Domain:        &fqdn,
		CompartmentId: &d.config.compartmentID,
	}

	ctx := context.Background()
	_, err := d.client.DeleteDomainRecords(ctx, request)
	if err != nil {
		panic(err)
	}

	return nil
}

// DeleteLastDot Delete the last dot.
// error occur if the last dot exist in oci-go-sdk.
func DeleteLastDot(fqdn string) string {
	if strings.HasSuffix(fqdn, ".") {
		fqdn = strings.TrimRight(fqdn, ".")
	}
	return fqdn
}
