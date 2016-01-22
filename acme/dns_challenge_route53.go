package acme

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/route53"
)

// DNSProviderRoute53 is an implementation of the DNSProvider interface
type DNSProviderRoute53 struct {
	client *route53.Route53
}

// NewDNSProviderRoute53 returns a DNSProviderRoute53 instance with a configured route53 client.
// Authentication is either done using the passed credentials or - when empty -
// using the environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.
func NewDNSProviderRoute53(awsAccessKey, awsSecretKey, awsRegionName string) (*DNSProviderRoute53, error) {
	region, ok := aws.Regions[awsRegionName]
	if !ok {
		return nil, fmt.Errorf("Invalid AWS region name %s", awsRegionName)
	}

	var auth aws.Auth
	// First try passed in credentials
	if awsAccessKey != "" && awsSecretKey != "" {
		auth = aws.Auth{awsAccessKey, awsSecretKey, ""}
	} else {
		// try getting credentials from environment
		envAuth, err := aws.EnvAuth()
		if err != nil {
			return nil, fmt.Errorf("AWS credentials missing")
		}
		auth = envAuth
	}

	client := route53.New(auth, region)
	return &DNSProviderRoute53{client: client}, nil
}

// CreateTXTRecord creates a TXT record using the specified parameters
func (r *DNSProviderRoute53) CreateTXTRecord(fqdn, value string, ttl int) error {
	return r.changeRecord("UPSERT", fqdn, value, ttl)
}

// RemoveTXTRecord removes the TXT record matching the specified parameters
func (r *DNSProviderRoute53) RemoveTXTRecord(fqdn, value string, ttl int) error {
	return r.changeRecord("DELETE", fqdn, value, ttl)
}

func (r *DNSProviderRoute53) changeRecord(action, fqdn, value string, ttl int) error {
	hostedZoneID, err := r.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}
	recordSet := newTXTRecordSet(fqdn, value, ttl)
	update := route53.Change{action, recordSet}
	changes := []route53.Change{update}
	req := route53.ChangeResourceRecordSetsRequest{Comment: "Created by Lego", Changes: changes}
	_, err = r.client.ChangeResourceRecordSets(hostedZoneID, &req)
	return err
}

func (r *DNSProviderRoute53) getHostedZoneID(fqdn string) (string, error) {
	zones := []route53.HostedZone{}
	zoneResp, err := r.client.ListHostedZones("", 0)
	if err != nil {
		return "", err
	}
	zones = append(zones, zoneResp.HostedZones...)

	for zoneResp.IsTruncated {
		resp, err := r.client.ListHostedZones(zoneResp.Marker, 0)
		if err != nil {
			if rateExceeded(err) {
				time.Sleep(time.Second)
				continue
			}
			return "", err
		}
		zoneResp = resp
		zones = append(zones, zoneResp.HostedZones...)
	}

	var hostedZone route53.HostedZone
	for _, zone := range zones {
		if strings.HasSuffix(fqdn, zone.Name) {
			if len(zone.Name) > len(hostedZone.Name) {
				hostedZone = zone
			}
		}
	}
	if hostedZone.ID == "" {
		return "", fmt.Errorf("No Route53 hosted zone found for domain %s", fqdn)
	}

	return hostedZone.ID, nil
}

func newTXTRecordSet(fqdn, value string, ttl int) route53.ResourceRecordSet {
	return route53.ResourceRecordSet{
		Name:    fqdn,
		Type:    "TXT",
		Records: []string{value},
		TTL:     ttl,
	}
}

// Route53 API has pretty strict rate limits (5req/s globally per account)
// Hence we check if we are being throttled to maybe retry the request
func rateExceeded (err error) bool {
	if strings.Contains(err.Error(), "Throttling") {
		return true
	}
	return false
}
