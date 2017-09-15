package alibabacloud

import (
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/dns"
	"github.com/xenolf/lego/acme"
	"os"
	"strings"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dns.Client
}

// NewDNSProvider returns a DNSProvider instance with a configured Alibaba Cloud client.
func NewDNSProvider() (*DNSProvider, error) {
	apiKey := os.Getenv("ALIBABA_CLOUD_API_KEY")
	apiSecret := os.Getenv("ALIBABA_CLOUD_API_SECRET")
	if apiKey == "" {
		return nil, fmt.Errorf("Alibaba Cloud credentials missing")
	}
	if apiSecret == "" {
		return nil, fmt.Errorf("Alibaba Cloud credentials Secret missing")
	}
	return NewDNSProviderCredentials(apiKey, apiSecret)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Alibaba Cloud DNS.
func NewDNSProviderCredentials(apiKey, apiSecret string) (*DNSProvider, error) {
	return &DNSProvider{
		client: dns.NewClient(apiKey, apiSecret),
	}, nil
}

func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zoneDomain, domainErr := c.getHostedZone(domain)

	if domainErr != nil {
		return domainErr
	}

	name := c.extractRecordName(fqdn, zoneDomain)

	_, err := c.client.AddDomainRecord(&dns.AddDomainRecordArgs{
		DomainName: zoneDomain,
		RR:         name,
		Type:       "TXT",
		Value:      value,
		TTL:        600,
	})

	if err != nil {
		return fmt.Errorf("AliBaba Cloud call failed: %v", err)
	}
	return nil
}

func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	recordId, recordIdErr := c.findTxtRecords(domain, fqdn)
	if recordIdErr != nil {
		return recordIdErr
	}

	_, err := c.client.DeleteDomainRecord(&dns.DeleteDomainRecordArgs{
		RecordId: recordId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *DNSProvider) getHostedZone(domain string) (string, error) {
	var allDomains []dns.DomainType
	pagination := common.Pagination{
		PageNumber: 1,
		PageSize:   100,
	}
	args := &dns.DescribeDomainsArgs{}

	for {
		args.Pagination = pagination
		domains, err := c.client.DescribeDomains(args)
		if err != nil {
			return "", err
		}
		allDomains = append(allDomains, domains...)
		if len(domains) < pagination.PageSize {
			break
		}
		pagination.PageNumber += 1
	}

	var hostedDomain dns.DomainType
	for _, d := range allDomains {
		if strings.HasSuffix(domain, d.DomainName) {
			if len(d.DomainName) > len(hostedDomain.DomainName) {
				hostedDomain = d
			}
		}
	}

	if hostedDomain.DomainName == "" {
		return "", fmt.Errorf("No matching AliCloud domain found for domain %s", domain)
	}
	return hostedDomain.DomainName, nil
}

func (c *DNSProvider) findTxtRecords(domain, fqdn string) (string, error) {
	var recordId string
	var allRecords []dns.RecordType

	zoneDomain, err := c.getHostedZone(domain)
	if err != nil {
		return "", err
	}
	recordName := c.extractRecordName(fqdn, zoneDomain)

	pagination := common.Pagination{
		PageNumber: 1,
		PageSize:   100,
	}

	args := &dns.DescribeDomainRecordsArgs{
		DomainName: zoneDomain,
	}

	for {
		args.Pagination = pagination
		response, err := c.client.DescribeDomainRecords(args)
		if err != nil {
			return "", err
		}
		allRecords = append(allRecords, response.DomainRecords.Record...)
		if len(response.DomainRecords.Record) < pagination.PageSize {
			break
		}
		pagination.PageNumber += 1
	}

	for _, record := range allRecords {
		if record.Type == "TXT" && record.RR == recordName && record.DomainName == zoneDomain {
			recordId = record.RecordId
		}
	}

	if recordId == "" {
		return "", fmt.Errorf("No matching AliCloud domain record found for domain %s  and record %s ", domain, recordName)
	}

	return recordId, nil
}

func (c *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
