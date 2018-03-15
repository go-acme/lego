// Package azure implements a DNS provider for solving the DNS-01
// challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/dns/mgmt/dns"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/xenolf/lego/acme"
)

// TODO: replace context.TODO() with context.WithTimeout()

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	subscriptionID          string
	resourceGroup           string
	clientCredentialsConfig auth.ClientCredentialsConfig
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials must be passed in the environment variables: AZURE_CLIENT_ID,
// AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
func NewDNSProvider() (*DNSProvider, error) {
	clientID, ok := os.LookupEnv("AZURE_CLIENT_ID")
	if !ok {
		return nil, errors.New("AZURE_CLIENT_ID was not set")
	}
	clientSecret, ok := os.LookupEnv("AZURE_CLIENT_SECRET")
	if !ok {
		return nil, errors.New("AZURE_CLIENT_SECRET was not set")
	}
	subscriptionID, ok := os.LookupEnv("AZURE_SUBSCRIPTION_ID")
	if !ok {
		return nil, errors.New("AZURE_SUBSCRIPTION_ID was not set")
	}
	tenantID, ok := os.LookupEnv("AZURE_TENANT_ID")
	if !ok {
		return nil, errors.New("AZURE_TENANT_ID was not set")
	}
	resourceGroup, ok := os.LookupEnv("AZURE_RESOURCE_GROUP")
	if !ok {
		return nil, errors.New("AZURE_RESOURCE_GROUP was not set")
	}
	return NewDNSProviderCredentials(clientID, clientSecret, subscriptionID, tenantID, resourceGroup)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for azure.
func NewDNSProviderCredentials(clientID, clientSecret, subscriptionID, tenantID, resourceGroup string) (*DNSProvider, error) {
	if clientID == "" || clientSecret == "" || subscriptionID == "" || tenantID == "" || resourceGroup == "" {
		return nil, fmt.Errorf("Azure configuration missing")
	}

	return &DNSProvider{
		subscriptionID:          subscriptionID,
		resourceGroup:           resourceGroup,
		clientCredentialsConfig: auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (c *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	rsc := dns.NewRecordSetsClient(c.subscriptionID)
	a, err := c.clientCredentialsConfig.Authorizer()
	if err != nil {
		return err
	}
	rsc.Authorizer = a

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(60),
			TxtRecords: &[]dns.TxtRecord{dns.TxtRecord{Value: &[]string{value}}},
		},
	}
	_, err = rsc.CreateOrUpdate(context.TODO(), c.resourceGroup, zone, relative, dns.TXT, rec, "", "")

	if err != nil {
		return err
	}

	return nil
}

// Returns the relative record to the domain
func toRelativeRecord(domain, zone string) string {
	return acme.UnFqdn(strings.TrimSuffix(domain, zone))
}

// CleanUp removes the TXT record matching the specified parameters
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rsc := dns.NewRecordSetsClient(c.subscriptionID)
	a, err := c.clientCredentialsConfig.Authorizer()
	if err != nil {
		return err
	}
	rsc.Authorizer = a

	_, err = rsc.Delete(context.TODO(), c.resourceGroup, zone, relative, dns.TXT, "")
	if err != nil {
		return err
	}

	return nil
}

// Checks that azure has a zone for this domain name.
func (c *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	// Now we want to to Azure and get the zone.
	dc := dns.NewZonesClient(c.subscriptionID)
	a, err := c.clientCredentialsConfig.Authorizer()
	if err != nil {
		return "", err
	}
	dc.Authorizer = a

	zone, err := dc.Get(context.TODO(), c.resourceGroup, acme.UnFqdn(authZone))

	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}
