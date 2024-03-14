package azuredns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// DNSProviderPrivate implements the challenge.Provider interface for Azure Private Zone DNS.
type DNSProviderPrivate struct {
	config      *Config
	credentials azcore.TokenCredential
}

// NewDNSProviderPrivate creates a DNSProviderPrivate structure.
func NewDNSProviderPrivate(config *Config, credentials azcore.TokenCredential) (*DNSProviderPrivate, error) {
	return &DNSProviderPrivate{
		config:      config,
		credentials: credentials,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProviderPrivate) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProviderPrivate) Present(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZoneID(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	client, err := d.newZoneClient(zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	// Get existing record set
	resp, err := client.Get(ctx, subDomain)
	if err != nil {
		var respErr *azcore.ResponseError
		if !errors.As(err, &respErr) || respErr.StatusCode != http.StatusNotFound {
			return fmt.Errorf("azuredns: %w", err)
		}
	}

	// Construct unique TXT records using map
	uniqRecords := privateUniqueRecords(resp.RecordSet, info.Value)

	var txtRecords []*armprivatedns.TxtRecord
	for txt := range uniqRecords {
		txtRecord := txt
		txtRecords = append(txtRecords, &armprivatedns.TxtRecord{Value: []*string{&txtRecord}})
	}

	rec := armprivatedns.RecordSet{
		Name: &subDomain,
		Properties: &armprivatedns.RecordSetProperties{
			TTL:        pointer(int64(d.config.TTL)),
			TxtRecords: txtRecords,
		},
	}

	_, err = client.CreateOrUpdate(ctx, subDomain, rec)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProviderPrivate) CleanUp(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZoneID(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	client, err := d.newZoneClient(zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	_, err = client.Delete(ctx, subDomain)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	return nil
}

// Checks that azure has a zone for this domain name.
func (d *DNSProviderPrivate) getHostedZoneID(fqdn string) (string, error) {
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

// newZoneClient creates PrivateZoneClient structure with initialized Azure client.
func (d *DNSProviderPrivate) newZoneClient(zoneName string) (*PrivateZoneClient, error) {
	zone, exists := d.config.ServiceDiscoveryZones[zoneName]
	if !exists {
		return nil, fmt.Errorf(`zone %s not found`, zoneName)
	}

	return NewPrivateZoneClient(zone, d.credentials, d.config.Environment)
}

// PrivateZoneClient provides Azure client for one DNS zone.
type PrivateZoneClient struct {
	zone         ServiceDiscoveryZone
	recordClient *armprivatedns.RecordSetsClient
}

func NewPrivateZoneClient(zone ServiceDiscoveryZone, credential azcore.TokenCredential, environment cloud.Configuration) (*PrivateZoneClient, error) {
	options := &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: environment,
		},
	}

	recordClient, err := armprivatedns.NewRecordSetsClient(zone.SubscriptionID, credential, options)
	if err != nil {
		return nil, err
	}

	return &PrivateZoneClient{
		zone:         zone,
		recordClient: recordClient,
	}, nil
}

func (c PrivateZoneClient) Get(ctx context.Context, subDomain string) (armprivatedns.RecordSetsClientGetResponse, error) {
	return c.recordClient.Get(ctx, c.zone.ResourceGroup, c.zone.Name, armprivatedns.RecordTypeTXT, subDomain, nil)
}

func (c PrivateZoneClient) CreateOrUpdate(ctx context.Context, subDomain string, rec armprivatedns.RecordSet) (armprivatedns.RecordSetsClientCreateOrUpdateResponse, error) {
	return c.recordClient.CreateOrUpdate(ctx, c.zone.ResourceGroup, c.zone.Name, armprivatedns.RecordTypeTXT, subDomain, rec, nil)
}

func (c PrivateZoneClient) Delete(ctx context.Context, subDomain string) (armprivatedns.RecordSetsClientDeleteResponse, error) {
	return c.recordClient.Delete(ctx, c.zone.ResourceGroup, c.zone.Name, armprivatedns.RecordTypeTXT, subDomain, nil)
}

func privateUniqueRecords(recordSet armprivatedns.RecordSet, value string) map[string]struct{} {
	uniqRecords := map[string]struct{}{value: {}}
	if recordSet.Properties != nil && recordSet.Properties.TxtRecords != nil {
		for _, txtRecord := range recordSet.Properties.TxtRecords {
			// Assume Value doesn't contain multiple strings
			if len(txtRecord.Value) > 0 {
				uniqRecords[deref(txtRecord.Value[0])] = struct{}{}
			}
		}
	}

	return uniqRecords
}
