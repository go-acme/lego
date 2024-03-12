package azuredns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// DNSProviderPublic implements the challenge.Provider interface for Azure Public Zone DNS.
type DNSProviderPublic struct {
	config      *Config
	credentials azcore.TokenCredential
}

// DnsProviderPublicZone provides Azure client for one DNS zone
type DnsProviderPublicZone struct {
	Name           string
	ResourceGroup  string
	SubscriptionID string

	RecordClient *armdns.RecordSetsClient
}

// NewDNSProviderPublic creates a DNSProviderPublic structure
func NewDNSProviderPublic(config *Config, credentials azcore.TokenCredential) (*DNSProviderPublic, error) {
	return &DNSProviderPublic{
		config:      config,
		credentials: credentials,
	}, nil
}

// zoneClient creates DnsProviderPublicZone structure with initialised Azure client.
func (d *DNSProviderPublic) zoneClient(zoneName string) (*DnsProviderPublicZone, error) {
	options := arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: d.config.Environment,
		},
	}

	if zone, exists := d.config.ServiceDiscoveryZones[zoneName]; exists {
		recordClient, err := armdns.NewRecordSetsClient(zone.SubscriptionID, d.credentials, &options)
		if err != nil {
			panic(err)
		}

		return &DnsProviderPublicZone{
			RecordClient:   recordClient,
			Name:           zone.Name,
			ResourceGroup:  zone.ResourceGroup,
			SubscriptionID: zone.SubscriptionID,
		}, nil
	} else {
		return nil, fmt.Errorf(`zone %s not found`, zoneName)
	}
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProviderPublic) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProviderPublic) Present(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	zoneClient, err := d.zoneClient(zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	// Get existing record set
	rset, err := zoneClient.RecordClient.Get(ctx, zoneClient.ResourceGroup, zoneClient.Name, subDomain, armdns.RecordTypeTXT, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if !errors.As(err, &respErr) || respErr.StatusCode != http.StatusNotFound {
			return fmt.Errorf("azuredns: %w", err)
		}
	}

	// Construct unique TXT records using map
	uniqRecords := map[string]struct{}{info.Value: {}}
	if rset.RecordSet.Properties != nil && rset.RecordSet.Properties.TxtRecords != nil {
		for _, txtRecord := range rset.RecordSet.Properties.TxtRecords {
			// Assume Value doesn't contain multiple strings
			if len(txtRecord.Value) > 0 {
				uniqRecords[deref(txtRecord.Value[0])] = struct{}{}
			}
		}
	}

	var txtRecords []*armdns.TxtRecord
	for txt := range uniqRecords {
		txtRecord := txt
		txtRecords = append(txtRecords, &armdns.TxtRecord{Value: []*string{&txtRecord}})
	}

	ttlInt64 := int64(d.config.TTL)
	rec := armdns.RecordSet{
		Name: &subDomain,
		Properties: &armdns.RecordSetProperties{
			TTL:        &ttlInt64,
			TxtRecords: txtRecords,
		},
	}

	_, err = zoneClient.RecordClient.CreateOrUpdate(ctx, zoneClient.ResourceGroup, zoneClient.Name, subDomain, armdns.RecordTypeTXT, rec, nil)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProviderPublic) CleanUp(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	zoneClient, err := d.zoneClient(zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	_, err = zoneClient.RecordClient.Delete(ctx, zoneClient.ResourceGroup, zoneClient.Name, subDomain, armdns.RecordTypeTXT, nil)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	return nil
}

// Checks that azure has a zone for this domain name.
func (d *DNSProviderPublic) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	if zone := env.GetOrFile(EnvZoneName); zone != "" {
		return zone, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone: %w", err)
	}

	if azureZone, exists := d.config.ServiceDiscoveryZones[dns01.UnFqdn(authZone)]; exists {
		return dns01.UnFqdn(azureZone.Name), nil
	}

	return "", fmt.Errorf(`could not find zone: %s`, authZone)
}
