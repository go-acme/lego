package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/dns/mgmt/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// dnsProviderPublic implements the challenge.Provider interface for Azure Public Zone DNS.
type dnsProviderPublic struct {
	config     *Config
	authorizer autorest.Authorizer
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *dnsProviderPublic) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *dnsProviderPublic) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}

	rsc := dns.NewRecordSetsClientWithBaseURI(d.config.ResourceManagerEndpoint, d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}

	// Get existing record set
	rset, err := rsc.Get(ctx, d.config.ResourceGroup, zone, subDomain, dns.TXT)
	if err != nil {
		var detailed autorest.DetailedError
		if !errors.As(err, &detailed) || detailed.StatusCode != http.StatusNotFound {
			return fmt.Errorf("azure: %w", err)
		}
	}

	// Construct unique TXT records using map
	uniqRecords := map[string]struct{}{info.Value: {}}
	if rset.RecordSetProperties != nil && rset.TxtRecords != nil {
		for _, txtRecord := range *rset.TxtRecords {
			// Assume Value doesn't contain multiple strings
			values := to.StringSlice(txtRecord.Value)
			if len(values) > 0 {
				uniqRecords[values[0]] = struct{}{}
			}
		}
	}

	var txtRecords []dns.TxtRecord
	for txt := range uniqRecords {
		txtRecords = append(txtRecords, dns.TxtRecord{Value: &[]string{txt}})
	}

	rec := dns.RecordSet{
		Name: &subDomain,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(int64(d.config.TTL)),
			TxtRecords: &txtRecords,
		},
	}

	_, err = rsc.CreateOrUpdate(ctx, d.config.ResourceGroup, zone, subDomain, dns.TXT, rec, "", "")
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *dnsProviderPublic) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}

	rsc := dns.NewRecordSetsClientWithBaseURI(d.config.ResourceManagerEndpoint, d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	_, err = rsc.Delete(ctx, d.config.ResourceGroup, zone, subDomain, dns.TXT, "")
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}
	return nil
}

// Checks that azure has a zone for this domain name.
func (d *dnsProviderPublic) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	if zone := env.GetOrFile(EnvZoneName); zone != "" {
		return zone, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone for FQDN %q: %w", fqdn, err)
	}

	dc := dns.NewZonesClientWithBaseURI(d.config.ResourceManagerEndpoint, d.config.SubscriptionID)
	dc.Authorizer = d.authorizer

	zone, err := dc.Get(ctx, d.config.ResourceGroup, dns01.UnFqdn(authZone))
	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}
