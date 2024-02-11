package azuredns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// DNSProviderPrivate implements the challenge.Provider interface for Azure Private Zone DNS.
type DNSProviderPrivate struct {
	config       *Config
	zoneClient   *armprivatedns.PrivateZonesClient
	recordClient *armprivatedns.RecordSetsClient
}

// NewDNSProviderPrivate creates a DNSProviderPrivate structure with initialized Azure clients.
func NewDNSProviderPrivate(config *Config, credentials azcore.TokenCredential) (*DNSProviderPrivate, error) {
	options := arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: config.Environment,
		},
	}

	zoneClient, err := armprivatedns.NewPrivateZonesClient(config.SubscriptionID, credentials, &options)
	if err != nil {
		return nil, err
	}

	recordClient, err := armprivatedns.NewRecordSetsClient(config.SubscriptionID, credentials, &options)
	if err != nil {
		return nil, err
	}

	return &DNSProviderPrivate{
		config:       config,
		zoneClient:   zoneClient,
		recordClient: recordClient,
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

	zone, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zone)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	// Get existing record set
	rset, err := d.recordClient.Get(ctx, d.config.ResourceGroup, zone, armprivatedns.RecordTypeTXT, subDomain, nil)
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

	var txtRecords []*armprivatedns.TxtRecord
	for txt := range uniqRecords {
		txtRecord := txt
		txtRecords = append(txtRecords, &armprivatedns.TxtRecord{Value: []*string{&txtRecord}})
	}

	ttlInt64 := int64(d.config.TTL)
	rec := armprivatedns.RecordSet{
		Name: &subDomain,
		Properties: &armprivatedns.RecordSetProperties{
			TTL:        &ttlInt64,
			TxtRecords: txtRecords,
		},
	}

	_, err = d.recordClient.CreateOrUpdate(ctx, d.config.ResourceGroup, zone, armprivatedns.RecordTypeTXT, subDomain, rec, nil)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProviderPrivate) CleanUp(domain, _, keyAuth string) error {
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

	_, err = d.recordClient.Delete(ctx, d.config.ResourceGroup, zone, armprivatedns.RecordTypeTXT, subDomain, nil)
	if err != nil {
		return fmt.Errorf("azuredns: %w", err)
	}

	return nil
}

// Checks that azure has a zone for this domain name.
func (d *DNSProviderPrivate) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	if zone := env.GetOrFile(EnvZoneName); zone != "" {
		return zone, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone: %w", err)
	}

	zone, err := d.zoneClient.Get(ctx, d.config.ResourceGroup, dns01.UnFqdn(authZone), nil)
	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return dns01.UnFqdn(deref(zone.Name)), nil
}
