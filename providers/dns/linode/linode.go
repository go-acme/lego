// Package linode implements a DNS provider for solving the DNS-01 challenge
// using Linode DNS.
package linode

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timewasted/linode/dns"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const (
	dnsMinTTLSecs      = 300
	dnsUpdateFreqMins  = 15
	dnsUpdateFudgeSecs = 120
)

type hostedZoneInfo struct {
	domainID     int
	resourceName string
}

// DNSProvider implements the acme.ChallengeProvider interface.
type DNSProvider struct {
	linode *dns.DNS
}

// NewDNSProvider returns a DNSProvider instance configured for Linode.
// Credentials must be passed in the environment variable: LINODE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("LINODE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("Linode: %v", err)
	}

	return NewDNSProviderCredentials(values["LINODE_API_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Linode.
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	if len(apiKey) == 0 {
		return nil, errors.New("Linode credentials missing")
	}

	return &DNSProvider{
		linode: dns.New(apiKey),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation.  Adjusting here to cope with spikes in propagation times.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	// Since Linode only updates their zone files every X minutes, we need
	// to figure out how many minutes we have to wait until we hit the next
	// interval of X.  We then wait another couple of minutes, just to be
	// safe.  Hopefully at some point during all of this, the record will
	// have propagated throughout Linode's network.
	minsRemaining := dnsUpdateFreqMins - (time.Now().Minute() % dnsUpdateFreqMins)

	timeout = (time.Duration(minsRemaining) * time.Minute) +
		(dnsMinTTLSecs * time.Second) +
		(dnsUpdateFudgeSecs * time.Second)
	interval = 15 * time.Second
	return
}

// Present creates a TXT record using the specified parameters.
func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := p.getHostedZoneInfo(fqdn)
	if err != nil {
		return err
	}

	if _, err = p.linode.CreateDomainResourceTXT(zone.domainID, acme.UnFqdn(fqdn), value, 60); err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := p.getHostedZoneInfo(fqdn)
	if err != nil {
		return err
	}

	// Get all TXT records for the specified domain.
	resources, err := p.linode.GetResourcesByType(zone.domainID, "TXT")
	if err != nil {
		return err
	}

	// Remove the specified resource, if it exists.
	for _, resource := range resources {
		if resource.Name == zone.resourceName && resource.Target == value {
			resp, err := p.linode.DeleteDomainResource(resource.DomainID, resource.ResourceID)
			if err != nil {
				return err
			}
			if resp.ResourceID != resource.ResourceID {
				return errors.New("error deleting resource: resource IDs do not match")
			}
			break
		}
	}

	return nil
}

func (p *DNSProvider) getHostedZoneInfo(fqdn string) (*hostedZoneInfo, error) {
	// Lookup the zone that handles the specified FQDN.
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return nil, err
	}
	resourceName := strings.TrimSuffix(fqdn, "."+authZone)

	// Query the authority zone.
	domain, err := p.linode.GetDomain(acme.UnFqdn(authZone))
	if err != nil {
		return nil, err
	}

	return &hostedZoneInfo{
		domainID:     domain.DomainID,
		resourceName: resourceName,
	}, nil
}
