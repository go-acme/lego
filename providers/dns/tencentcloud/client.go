package tencentcloud

import (
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type domainData struct {
	domain    string
	subDomain string
}

func getDomainData(fqdn string) (*domainData, error) {
	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, err
	}

	return &domainData{
		domain:    dns01.UnFqdn(zone),
		subDomain: dns01.UnFqdn(strings.TrimSuffix(fqdn, zone)),
	}, nil
}

func (d *DNSProvider) createRecordData(domainData *domainData, value string) error {
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domainData.domain)
	request.SubDomain = common.StringPtr(domainData.subDomain)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.TTL = common.Uint64Ptr(uint64(d.config.TTL))

	_, err := d.client.CreateRecord(request)
	if err != nil {
		return err
	}

	return nil
}

func (d *DNSProvider) listRecordData(domainData *domainData) ([]*dnspod.RecordListItem, error) {
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domainData.domain)
	request.Subdomain = common.StringPtr(domainData.subDomain)
	request.RecordType = common.StringPtr("TXT")

	response, err := d.client.DescribeRecordList(request)
	if err != nil {
		return nil, err
	}

	return response.Response.RecordList, nil
}

func (d *DNSProvider) deleteRecordData(domainData *domainData, item *dnspod.RecordListItem) error {
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domainData.domain)
	request.RecordId = item.RecordId

	_, err := d.client.DeleteRecord(request)
	if err != nil {
		return err
	}

	return nil
}
