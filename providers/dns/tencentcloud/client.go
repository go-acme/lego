package tencentcloud

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	errorsdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"golang.org/x/net/idna"
)

func (d *DNSProvider) getHostedZone(domain string) (*dnspod.DomainListItem, error) {
	request := dnspod.NewDescribeDomainListRequest()

	var domains []*dnspod.DomainListItem

	for {
		response, err := d.client.DescribeDomainList(request)
		if err != nil {
			return nil, fmt.Errorf("API call failed: %w", err)
		}

		domains = append(domains, response.Response.DomainList...)

		if uint64(len(domains)) >= *response.Response.DomainCountInfo.AllTotal {
			break
		}

		request.Offset = common.Int64Ptr(int64(len(domains)))
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return nil, err
	}

	var hostedZone *dnspod.DomainListItem
	for _, zone := range domains {
		if *zone.Name == dns01.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone == nil {
		return nil, fmt.Errorf("zone %s not found in dnspod for domain %s", authZone, domain)
	}

	return hostedZone, nil
}

func (d *DNSProvider) findTxtRecords(zone *dnspod.DomainListItem, fqdn string) ([]*dnspod.RecordListItem, error) {
	recordName, err := extractRecordName(fqdn, *zone.Name)
	if err != nil {
		return nil, err
	}

	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = zone.Name
	request.DomainId = zone.DomainId
	request.Subdomain = common.StringPtr(recordName)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")

	response, err := d.client.DescribeRecordList(request)
	if err != nil {
		var sdkError *errorsdk.TencentCloudSDKError
		if errors.As(err, &sdkError) {
			if sdkError.Code == dnspod.RESOURCENOTFOUND_NODATAOFRECORD {
				return nil, nil
			}
		}
		return nil, err
	}

	return response.Response.RecordList, nil
}

func extractRecordName(fqdn, zone string) (string, error) {
	asciiDomain, err := idna.ToASCII(zone)
	if err != nil {
		return "", fmt.Errorf("fail to convert punycode: %w", err)
	}

	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+asciiDomain); idx != -1 {
		return name[:idx], nil
	}
	return name, nil
}
