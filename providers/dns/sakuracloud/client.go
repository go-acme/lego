package sakuracloud

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
)

// This mutex is required for concurrent updates.
// see: https://github.com/go-acme/lego/pull/850
var mu sync.Mutex // nolint: gochecknoglobals

func (d *DNSProvider) addTXTRecord(fqdn, domain, value string, ttl int) error {
	mu.Lock()
	defer mu.Unlock()

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %w", err)
	}

	name := extractRecordName(fqdn, zone.Name)

	records := append(zone.Records, &iaas.DNSRecord{
		Name:  name,
		Type:  "TXT",
		RData: value,
		TTL:   ttl,
	})
	_, err = d.client.UpdateSettings(context.Background(), zone.ID, &iaas.DNSUpdateSettingsRequest{
		Records:      records,
		SettingsHash: zone.SettingsHash,
	})
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %w", err)
	}

	return nil
}

func (d *DNSProvider) cleanupTXTRecord(fqdn, domain string) error {
	mu.Lock()
	defer mu.Unlock()

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %w", err)
	}

	txtRecords := findTxtRecords(fqdn, zone)
	var updRecords iaas.DNSRecords

	for _, record := range txtRecords {
		for _, r := range zone.Records {
			if !(r.Name == record.Name && r.Type == record.Type && r.RData == record.RData) {
				updRecords = append(updRecords, r)
			}
		}
	}

	_, err = d.client.UpdateSettings(context.Background(), zone.ID, &iaas.DNSUpdateSettingsRequest{
		Records:      updRecords,
		SettingsHash: zone.SettingsHash,
	})
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %w", err)
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (*iaas.DNS, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return nil, err
	}

	zoneName := dns01.UnFqdn(authZone)

	res, err := d.client.Find(context.Background(), &iaas.FindCondition{
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch(zoneName),
		},
	})
	if err != nil {
		if iaas.IsNotFoundError(err) {
			return nil, fmt.Errorf("zone %s not found on SakuraCloud DNS: %w", zoneName, err)
		}

		return nil, fmt.Errorf("API call failed: %w", err)
	}

	for _, zone := range res.DNS {
		if zone.Name == zoneName {
			return zone, nil
		}
	}

	return nil, fmt.Errorf("zone %s not found", zoneName)
}

func findTxtRecords(fqdn string, zone *iaas.DNS) iaas.DNSRecords {
	recordName := extractRecordName(fqdn, zone.Name)

	var res iaas.DNSRecords
	for _, record := range zone.Records {
		if record.Name == recordName && record.Type == "TXT" {
			res = append(res, record)
		}
	}
	return res
}

func extractRecordName(fqdn, zone string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+zone); idx != -1 {
		return name[:idx]
	}
	return name
}
