package sakuracloud

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

type sacloudDNSAPI interface {
	update(id int64, value *sacloud.DNS) (*sacloud.DNS, error)
	find(zoneName string) (*api.SearchDNSResponse, error)
}

type defaultSacloudDNSAPI struct {
	client *api.DNSAPI
}

func (d *defaultSacloudDNSAPI) update(id int64, value *sacloud.DNS) (*sacloud.DNS, error) {
	return d.client.Update(id, value)
}

func (d *defaultSacloudDNSAPI) find(zoneName string) (*api.SearchDNSResponse, error) {
	return d.client.Reset().WithNameLike(zoneName).Find()
}

type sacloudClient struct {
	client sacloudDNSAPI
}

func newSacloudClient(token, secret string) *sacloudClient {
	return &sacloudClient{
		client: &defaultSacloudDNSAPI{
			client: api.NewClient(token, secret, "is1a").GetDNSAPI(),
		},
	}
}

func (d *sacloudClient) AddTXTRecord(fqdn, domain, value string, ttl int) error {
	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	name := d.extractRecordName(fqdn, zone.Name)

	zone.AddRecord(zone.CreateNewRecord(name, "TXT", value, ttl))
	_, err = d.client.update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %v", err)
	}

	return nil
}

func (d *sacloudClient) CleanupTXTRecord(fqdn, domain string) error {
	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	records := d.findTxtRecords(fqdn, zone)

	for _, record := range records {
		var updRecords []sacloud.DNSRecordSet
		for _, r := range zone.Settings.DNS.ResourceRecordSets {
			if !(r.Name == record.Name && r.Type == record.Type && r.RData == record.RData) {
				updRecords = append(updRecords, r)
			}
		}
		zone.Settings.DNS.ResourceRecordSets = updRecords
	}

	_, err = d.client.update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %v", err)
	}
	return nil
}

func (d *sacloudClient) getHostedZone(domain string) (*sacloud.DNS, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return nil, err
	}

	zoneName := dns01.UnFqdn(authZone)

	res, err := d.client.find(zoneName)
	if err != nil {
		if notFound, ok := err.(api.Error); ok && notFound.ResponseCode() == http.StatusNotFound {
			return nil, fmt.Errorf("zone %s not found on SakuraCloud DNS: %v", zoneName, err)
		}
		return nil, fmt.Errorf("API call failed: %v", err)
	}

	for _, zone := range res.CommonServiceDNSItems {
		if zone.Name == zoneName {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone %s not found", zoneName)
}

func (d *sacloudClient) findTxtRecords(fqdn string, zone *sacloud.DNS) []sacloud.DNSRecordSet {
	recordName := d.extractRecordName(fqdn, zone.Name)

	var res []sacloud.DNSRecordSet
	for _, record := range zone.Settings.DNS.ResourceRecordSets {
		if record.Name == recordName && record.Type == "TXT" {
			res = append(res, record)
		}
	}
	return res
}

func (d *sacloudClient) extractRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
