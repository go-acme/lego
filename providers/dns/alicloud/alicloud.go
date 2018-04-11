// Package alicloud implements a DNS provider for solving the DNS-01 challenge
// using alicloud DNS.
package alicloud

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/xenolf/lego/acme"
	"fmt"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *alidns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for alicloud.
// Credentials must be passed in the environment variables: ACCESS_KEY_ID and ACCESS_KEY_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	key_id := os.Getenv("ACCESS_KEY_ID")
	key_secret := os.Getenv("ACCESS_KEY_SECRET")
	return NewDNSProviderCredentials(key_id, key_secret)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for alicloud.
func NewDNSProviderCredentials(key_id string, key_secret string) (*DNSProvider, error) {
	if key_id == "" || key_secret == "" {
		return nil, fmt.Errorf("alicloud credentials missing")
	}

	client, err := alidns.NewClientWithAccessKey("cn-beijing", key_id, key_secret)
	if err != nil {
		panic(err)
	}
	return &DNSProvider{
		client: client,
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	recordAttributes := c.newTxtRecord(fqdn, value, ttl)

	// 创建API请求并设置参数
    request := c.client.CreateAddDomainRecordRequest(recordAttributes)
    // 发起请求并处理异常
    response, err := c.client.AddDomainRecord(request)
    if err != nil {
    	// 异常处理
    	panic(err)
    }
	if err != nil {
		return fmt.Errorf("alicloud API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := c.findTxtRecords(domain, fqdn)
	if err != nil {
		return err
	}

	for _, rec := range records {
		_, err := c.client.DeleteDomainRecord(&DeleteDomainRecordRequest{
			RecordId: rec.RequestId
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *DNSProvider) findTxtRecords(domain, fqdn string) ([]alidns.Record, error) {

	var records []alidns.Record
	response, err := c.client.DescribeSubDomainRecords(&alidns.DescribeSubDomainRecordsRequest{
		SubDomain: domain,
		ype:  "TXT",
		})
	if err != nil {
		return records, fmt.Errorf("alidns API call has failed: %v", err)
	}

	recordName := c.extractRecordName(fqdn)

	for _, record := range response.DomainRecords.Record {
		if record.RR == recordName {
			records = append(records, record)
		}
	}

	return records, nil
}

func (c *DNSProvider) newTxtRecord(fqdn, value string, ttl int) *alidns.AddDomainRecordRequest {
	name := c.extractRecordName(fqdn)

	return &alidns.AddDomainRecordRequest{
		DomainName:  fqdn,
		RR: rr,
		Type:  "TXT",
		Value: value,
	}
}

func (c *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}