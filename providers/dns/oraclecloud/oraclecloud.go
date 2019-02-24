package oraclecloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sugi275/oci-env-configprovider/envprovider"
	"github.com/oracle/oci-go-sdk/dns"
	"github.com/xenolf/lego/challenge/dns01"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OracleCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	fqdn = DeleteLastDot(fqdn)
	fmt.Println("fqdn: ", fqdn)
	fmt.Println("value: ", value)
	fmt.Println("domain: ", domain)

	client, err := dns.NewDnsClientWithConfigurationProvider(envprovider.GetEnvConfigProvider())
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	ctx := context.Background()
	compartmentid, err := envprovider.GetCompartmentID()
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	// DNSのレコードを作成するパラメータを生成
	//zn := "test.enc"
	//dn := "_acme-challenge.test.enc"
	txttype := "TXT"
	falseFlg := false
	//rdata := "testdayoooooo!"
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
		CompartmentId:              &compartmentid,
	}

	response, err := client.UpdateDomainRecords(ctx, request)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fmt.Println("helloOracle CleanUp")

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
