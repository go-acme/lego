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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// DNSProviderPrivate implements the challenge.Provider interface for Azure Private Zone DNS.
type DNSProviderPrivate struct {
	config                *Config
	credentials           azcore.TokenCredential
	serviceDiscoveryZones map[string]ServiceDiscoveryZone
}

// NewDNSProviderPrivate creates a DNSProviderPrivate structure.
func NewDNSProviderPrivate(config *Config, credentials azcore.TokenCredential) (*DNSProviderPrivate, error) {
	zones, err := discoverDNSZones(context.Background(), config, credentials)
	if err != nil {
		return nil, fmt.Errorf("discover DNS zones: %w", err)
	}

	return &DNSProviderPrivate{
		config:                config,
		credentials:           credentials,
		serviceDiscoveryZones: zones,
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

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	client, err := newPrivateZoneClient(zone, d.credentials, d.config.Environment)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
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
		txtRecords = append(txtRecords, &armprivatedns.TxtRecord{Value: to.SliceOfPtrs(txt)})
	}

	rec := armprivatedns.RecordSet{
		Name: &subDomain,
		Properties: &armprivatedns.RecordSetProperties{
			TTL:        to.Ptr(int64(d.config.TTL)),
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

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	client, err := newPrivateZoneClient(zone, d.credentials, d.config.Environment)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone.Name)
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
func (d *DNSProviderPrivate) getHostedZone(fqdn string) (ServiceDiscoveryZone, error) {
	authZone, err := getZoneName(d.config, fqdn)
	if err != nil {
		return ServiceDiscoveryZone{}, err
	}

	azureZone, exists := d.serviceDiscoveryZones[dns01.UnFqdn(authZone)]
	if !exists {
		return ServiceDiscoveryZone{}, fmt.Errorf("could not find zone (from discovery): %s", authZone)
	}

	return azureZone, nil
}

// privateZoneClient provides Azure client for one DNS zone.
type privateZoneClient struct {
	zone         ServiceDiscoveryZone
	recordClient *armprivatedns.RecordSetsClient
}

// newPrivateZoneClient creates privateZoneClient structure with initialized Azure client.
func newPrivateZoneClient(zone ServiceDiscoveryZone, credential azcore.TokenCredential, environment cloud.Configuration) (*privateZoneClient, error) {
	options := &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: environment,
		},
	}

	recordClient, err := armprivatedns.NewRecordSetsClient(zone.SubscriptionID, credential, options)
	if err != nil {
		return nil, err
	}

	return &privateZoneClient{
		zone:         zone,
		recordClient: recordClient,
	}, nil
}

func (c privateZoneClient) Get(ctx context.Context, subDomain string) (armprivatedns.RecordSetsClientGetResponse, error) {
	return c.recordClient.Get(ctx, c.zone.ResourceGroup, c.zone.Name, armprivatedns.RecordTypeTXT, subDomain, nil)
}

func (c privateZoneClient) CreateOrUpdate(ctx context.Context, subDomain string, rec armprivatedns.RecordSet) (armprivatedns.RecordSetsClientCreateOrUpdateResponse, error) {
	return c.recordClient.CreateOrUpdate(ctx, c.zone.ResourceGroup, c.zone.Name, armprivatedns.RecordTypeTXT, subDomain, rec, nil)
}

func (c privateZoneClient) Delete(ctx context.Context, subDomain string) (armprivatedns.RecordSetsClientDeleteResponse, error) {
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
