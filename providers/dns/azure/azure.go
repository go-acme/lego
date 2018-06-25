// Package azure implements a DNS provider for solving the DNS-01
// challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-09-01/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	clientID       string
	clientSecret   string
	subscriptionID string
	tenantID       string
	resourceGroup  string
	context        context.Context
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials must be passed in the environment variables: AZURE_CLIENT_ID,
// AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET", "AZURE_SUBSCRIPTION_ID", "AZURE_TENANT_ID", "AZURE_RESOURCE_GROUP")
	if err != nil {
		return nil, fmt.Errorf("Azure: %v", err)
	}

	return NewDNSProviderCredentials(
		values["AZURE_CLIENT_ID"],
		values["AZURE_CLIENT_SECRET"],
		values["AZURE_SUBSCRIPTION_ID"],
		values["AZURE_TENANT_ID"],
		values["AZURE_RESOURCE_GROUP"],
	)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for azure.
func NewDNSProviderCredentials(clientID, clientSecret, subscriptionID, tenantID, resourceGroup string) (*DNSProvider, error) {
	if clientID == "" || clientSecret == "" || subscriptionID == "" || tenantID == "" || resourceGroup == "" {
		return nil, errors.New("Azure: some credentials information are missing")
	}

	return &DNSProvider{
		clientID:       clientID,
		clientSecret:   clientSecret,
		subscriptionID: subscriptionID,
		tenantID:       tenantID,
		resourceGroup:  resourceGroup,
		// TODO: A timeout can be added here for cancellation purposes.
		context: context.Background(),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	rsc := dns.NewRecordSetsClient(d.subscriptionID)
	spt, err := d.newServicePrincipalTokenFromCredentials(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	rsc.Authorizer = autorest.NewBearerAuthorizer(spt)

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(60),
			TxtRecords: &[]dns.TxtRecord{{Value: &[]string{value}}},
		},
	}

	_, err = rsc.CreateOrUpdate(d.context, d.resourceGroup, zone, relative, dns.TXT, rec, "", "")
	return err
}

// Returns the relative record to the domain
func toRelativeRecord(domain, zone string) string {
	return acme.UnFqdn(strings.TrimSuffix(domain, zone))
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rsc := dns.NewRecordSetsClient(d.subscriptionID)
	spt, err := d.newServicePrincipalTokenFromCredentials(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	rsc.Authorizer = autorest.NewBearerAuthorizer(spt)

	_, err = rsc.Delete(d.context, d.resourceGroup, zone, relative, dns.TXT, "")
	return err
}

// Checks that azure has a zone for this domain name.
func (d *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	// Now we want to to Azure and get the zone.
	spt, err := d.newServicePrincipalTokenFromCredentials(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return "", err
	}

	dc := dns.NewZonesClient(d.subscriptionID)
	dc.Authorizer = autorest.NewBearerAuthorizer(spt)

	zone, err := dc.Get(d.context, d.resourceGroup, acme.UnFqdn(authZone))
	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func (d *DNSProvider) newServicePrincipalTokenFromCredentials(scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, d.tenantID)
	if err != nil {
		return nil, err
	}
	return adal.NewServicePrincipalToken(*oauthConfig, d.clientID, d.clientSecret, scope)
}
