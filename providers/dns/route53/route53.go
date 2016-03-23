// Package route53 implements a DNS provider for solving the DNS-01 challenge
// using route53 DNS.
package route53

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/route53"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *route53.Route53
}

// NewDNSProvider returns a DNSProvider instance configured for the AWS
// route53 service. The AWS region name must be passed in the environment
// variable AWS_REGION.
func NewDNSProvider() (*DNSProvider, error) {
	regionName := os.Getenv("AWS_REGION")
	return NewDNSProviderCredentials("", "", regionName)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for the AWS route53 service. Authentication
// is done using the passed credentials or, if empty, falling back to the
// custonmary AWS credential mechanisms, including the file referenced by
// $AWS_CREDENTIAL_FILE (defaulting to $HOME/.aws/credentials) optionally
// scoped to $AWS_PROFILE, credentials supplied by the environment variables
// AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY [ + AWS_SECURITY_TOKEN ], and
// finally credentials available via the EC2 instance metadata service.
func NewDNSProviderCredentials(accessKey, secretKey, regionName string) (*DNSProvider, error) {
	region, ok := aws.Regions[regionName]
	if !ok {
		return nil, fmt.Errorf("Invalid AWS region name %s", regionName)
	}

	// use aws.GetAuth, which tries really hard to find credentails:
	//   - uses accessKey and secretKey, if provided
	//   - uses AWS_PROFILE / AWS_CREDENTIAL_FILE, if provided
	//   - uses AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY and optionally AWS_SECURITY_TOKEN, if provided
	//   - uses EC2 instance metadata credentials (http://169.254.169.254/latest/meta-data/…), if available
	//  ...and otherwise returns an error
	auth, err := aws.GetAuth(accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	client := route53.New(auth, region)
	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record using the specified parameters
func (r *DNSProvider) Present(domain *acme.Domain, token, keyAuth string) error {
	fqdn, value, ttl := domain.GetDNS01Record(keyAuth)
	value = `"` + value + `"`
	return r.changeRecord("UPSERT", fqdn, value, ttl)
}

// CleanUp removes the TXT record matching the specified parameters
func (r *DNSProvider) CleanUp(domain *acme.Domain, token, keyAuth string) error {
	fqdn, value, ttl := domain.GetDNS01Record(keyAuth)
	value = `"` + value + `"`
	return r.changeRecord("DELETE", fqdn, value, ttl)
}

func (r *DNSProvider) changeRecord(action, fqdn, value string, ttl int) error {
	hostedZoneID, err := r.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}
	recordSet := newTXTRecordSet(fqdn, value, ttl)
	update := route53.Change{Action: action, Record: recordSet}
	changes := []route53.Change{update}
	req := route53.ChangeResourceRecordSetsRequest{Comment: "Created by Lego", Changes: changes}
	resp, err := r.client.ChangeResourceRecordSets(hostedZoneID, &req)
	if err != nil {
		return err
	}

	return acme.WaitFor(90*time.Second, 5*time.Second, func() (bool, error) {
		status, err := r.client.GetChange(resp.ChangeInfo.ID)
		if err != nil {
			return false, err
		}
		if status == "INSYNC" {
			return true, nil
		}
		return false, nil
	})
}

func (r *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
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
func rateExceeded(err error) bool {
	if strings.Contains(err.Error(), "Throttling") {
		return true
	}
	return false
}
