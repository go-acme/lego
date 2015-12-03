package acme

import (
	"fmt"
	"os"
	"strings"

	"github.com/crackcomm/cloudflare"
	"golang.org/x/net/context"
)

// DNSProviderCloudFlare is an implementation of the DNSProvider interface
type DNSProviderCloudFlare struct {
	client *cloudflare.Client
	ctx    context.Context
}

// NewDNSProviderCloudFlare returns a DNSProviderCloudFlare instance with a configured cloudflare client.
// Authentication is either done using the passed credentials or - when empty - using the environment
// variables CLOUDFLARE_EMAIL and CLOUDFLARE_API_KEY.
func NewDNSProviderCloudFlare(cloudflareEmail, cloudflareKey string) (*DNSProviderCloudFlare, error) {
	if cloudflareEmail == "" || cloudflareKey == "" {
		cloudflareEmail, cloudflareKey = envAuth()
		if cloudflareEmail == "" || cloudflareKey == "" {
			return nil, fmt.Errorf("CloudFlare credentials missing")
		}
	}

	c := &DNSProviderCloudFlare{
		client: cloudflare.New(&cloudflare.Options{cloudflareEmail, cloudflareKey}),
		ctx:    context.Background(),
	}

	return c, nil
}

// CreateTXTRecord creates a TXT record using the specified parameters
func (c *DNSProviderCloudFlare) CreateTXTRecord(fqdn, value string, ttl int) error {
	zoneID, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	record := newTxtRecord(zoneID, fqdn, value, ttl)
	err = c.client.Records.Create(c.ctx, record)
	if err != nil {
		return fmt.Errorf("CloudFlare API call failed: %v", err)
	}

	return nil
}

// RemoveTXTRecord removes the TXT record matching the specified parameters
func (c *DNSProviderCloudFlare) RemoveTXTRecord(fqdn, value string, ttl int) error {
	records, err := c.findTxtRecords(fqdn)
	if err != nil {
		return err
	}

	for _, rec := range records {
		err := c.client.Records.Delete(c.ctx, rec.ZoneID, rec.ID)
		if err != nil {
			return fmt.Errorf("CloudFlare API call has failed: %v", err)
		}
	}

	return nil
}

func (c *DNSProviderCloudFlare) findTxtRecords(fqdn string) ([]*cloudflare.Record, error) {
	zoneID, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return nil, err
	}

	var records []*cloudflare.Record
	result, err := c.client.Records.List(c.ctx, zoneID)
	if err != nil {
		return records, fmt.Errorf("CloudFlare API call has failed: %v", err)
	}

	name := unFqdn(fqdn)
	for _, rec := range result {
		if rec.Name == name && rec.Type == "TXT" {
			records = append(records, rec)
		}
	}

	return records, nil
}

func (c *DNSProviderCloudFlare) getHostedZoneID(fqdn string) (string, error) {
	zones, err := c.client.Zones.List(c.ctx)
	if err != nil {
		return "", fmt.Errorf("CloudFlare API call failed: %v", err)
	}

	var hostedZone cloudflare.Zone
	for _, zone := range zones {
		name := toFqdn(zone.Name)
		if strings.HasSuffix(fqdn, name) {
			if len(zone.Name) > len(hostedZone.Name) {
				hostedZone = *zone
			}
		}
	}
	if hostedZone.ID == "" {
		return "", fmt.Errorf("No matching CloudFlare zone found for domain %s", fqdn)
	}

	return hostedZone.ID, nil
}

func newTxtRecord(zoneID, fqdn, value string, ttl int) *cloudflare.Record {
	name := unFqdn(fqdn)
	return &cloudflare.Record{
		Type:    "TXT",
		Name:    name,
		Content: value,
		TTL:     sanitizeTTL(ttl),
		ZoneID:  zoneID,
	}
}

func toFqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	}
	return name + "."
}

func unFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}

// TTL must be between 120 and 86400 seconds
func sanitizeTTL(ttl int) int {
	if ttl < 120 {
		ttl = 120
	} else if ttl > 86400 {
		ttl = 86400
	}
	return ttl
}

func envAuth() (email, apiKey string) {
	email = os.Getenv("CLOUDFLARE_EMAIL")
	apiKey = os.Getenv("CLOUDFLARE_API_KEY")
	if len(email) == 0 || len(apiKey) == 0 {
		return "", ""
	}
	return
}
