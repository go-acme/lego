package sakuracloud

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
)

// This mutex is required for concurrent updates.
// see: https://github.com/go-acme/lego/pull/850
var mu sync.Mutex

func (d *DNSProvider) addTXTRecord(ctx context.Context, fqdn, value string, ttl int) error {
	mu.Lock()
	defer mu.Unlock()

	zone, err := d.getHostedZone(ctx, fqdn)
	if err != nil {
		return err
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone.Name)
	if err != nil {
		return err
	}

	records := append(zone.Records, &iaas.DNSRecord{
		Name:  subDomain,
		Type:  "TXT",
		RData: value,
		TTL:   ttl,
	})

	_, err = d.client.UpdateSettings(ctx, zone.ID, &iaas.DNSUpdateSettingsRequest{
		Records:      records,
		SettingsHash: zone.SettingsHash,
	})
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	return nil
}

func (d *DNSProvider) cleanupTXTRecord(ctx context.Context, fqdn, value string) error {
	mu.Lock()
	defer mu.Unlock()

	zone, err := d.getHostedZone(ctx, fqdn)
	if err != nil {
		return err
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone.Name)
	if err != nil {
		return err
	}

	var updRecords iaas.DNSRecords

	for _, r := range zone.Records {
		if !(r.Name == subDomain && r.Type == "TXT" && r.RData == value) { //nolint:staticcheck // Clearer without De Morgan's law.
			updRecords = append(updRecords, r)
		}
	}

	settings := &iaas.DNSUpdateSettingsRequest{
		Records:      updRecords,
		SettingsHash: zone.SettingsHash,
	}

	_, err = d.client.UpdateSettings(ctx, zone.ID, settings)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	return nil
}

func (d *DNSProvider) getHostedZone(ctx context.Context, domain string) (*iaas.DNS, error) {
	authZone, err := dns01.FindZoneByFqdn(domain)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	zoneName := dns01.UnFqdn(authZone)

	conditions := &iaas.FindCondition{
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch(zoneName),
		},
	}

	res, err := d.client.Find(ctx, conditions)
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
