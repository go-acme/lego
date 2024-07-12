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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// DNSProviderPublic implements the challenge.Provider interface for Azure Public Zone DNS.
type DNSProviderPublic struct {
	config                *Config
	credentials           azcore.TokenCredential
	serviceDiscoveryZones map[string]ServiceDiscoveryZone
}

// NewDNSProviderPublic creates a DNSProviderPublic structure.
func NewDNSProviderPublic(config *Config, credentials azcore.TokenCredential) (*DNSProviderPublic, error) {
	zones, err := discoverDNSZones(context.Background(), config, credentials)
	if err != nil {
		return nil, fmt.Errorf("discover DNS zones: %w", err)
	}

	return &DNSProviderPublic{
		config:                config,
		credentials:           credentials,
		serviceDiscoveryZones: zones,
	}, nil
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

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	client, err := newPublicZoneClient(zone, d.credentials, d.config.Environment)
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

	uniqRecords := publicUniqueRecords(resp.RecordSet, info.Value)

	var txtRecords []*armdns.TxtRecord
	for txt := range uniqRecords {
		txtRecords = append(txtRecords, &armdns.TxtRecord{Value: to.SliceOfPtrs(txt)})
	}

	rec := armdns.RecordSet{
		Name: &subDomain,
		Properties: &armdns.RecordSetProperties{
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
func (d *DNSProviderPublic) CleanUp(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	client, err := newPublicZoneClient(zone, d.credentials, d.config.Environment)
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
func (d *DNSProviderPublic) getHostedZone(fqdn string) (ServiceDiscoveryZone, error) {
	authZone, err := getAuthZone(fqdn)
	if err != nil {
		return ServiceDiscoveryZone{}, err
	}

	azureZone, exists := d.serviceDiscoveryZones[dns01.UnFqdn(authZone)]
	if !exists {
		return ServiceDiscoveryZone{}, fmt.Errorf("could not find zone (from discovery): %s", authZone)
	}

	return azureZone, nil
}

type publicZoneClient struct {
	zone         ServiceDiscoveryZone
	recordClient *armdns.RecordSetsClient
}

// newPublicZoneClient creates publicZoneClient structure with initialized Azure client.
func newPublicZoneClient(zone ServiceDiscoveryZone, credential azcore.TokenCredential, environment cloud.Configuration) (*publicZoneClient, error) {
	options := &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: environment,
		},
	}

	recordClient, err := armdns.NewRecordSetsClient(zone.SubscriptionID, credential, options)
	if err != nil {
		return nil, err
	}

	return &publicZoneClient{
		zone:         zone,
		recordClient: recordClient,
	}, nil
}

func (c publicZoneClient) Get(ctx context.Context, subDomain string) (armdns.RecordSetsClientGetResponse, error) {
	return c.recordClient.Get(ctx, c.zone.ResourceGroup, c.zone.Name, subDomain, armdns.RecordTypeTXT, nil)
}

func (c publicZoneClient) CreateOrUpdate(ctx context.Context, subDomain string, rec armdns.RecordSet) (armdns.RecordSetsClientCreateOrUpdateResponse, error) {
	return c.recordClient.CreateOrUpdate(ctx, c.zone.ResourceGroup, c.zone.Name, subDomain, armdns.RecordTypeTXT, rec, nil)
}

func (c publicZoneClient) Delete(ctx context.Context, subDomain string) (armdns.RecordSetsClientDeleteResponse, error) {
	return c.recordClient.Delete(ctx, c.zone.ResourceGroup, c.zone.Name, subDomain, armdns.RecordTypeTXT, nil)
}

func publicUniqueRecords(recordSet armdns.RecordSet, value string) map[string]struct{} {
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
