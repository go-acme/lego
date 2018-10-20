// Package azure implements a DNS provider for solving the DNS-01
// challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-09-01/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

var metadataEndpoint = "http://169.254.169.254"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	// optional if using instance metadata service
	ClientID     string
	ClientSecret string
	TenantID     string

	SubscriptionID string
	ResourceGroup  string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("AZURE_TTL", 60),
		PropagationTimeout: env.GetOrDefaultSecond("AZURE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("AZURE_POLLING_INTERVAL", 2*time.Second),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config     *Config
	authorizer autorest.Authorizer
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials cat be passed in the environment variables: AZURE_CLIENT_ID,
// AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
// If the credentials are _not_ set via the environment, then it will attempt
// to get a bearer token via the instance metadata service.
// see: https://github.com/Azure/go-autorest/blob/v10.14.0/autorest/azure/auth/auth.go#L38-L42
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for azure.
// Deprecated
func NewDNSProviderCredentials(clientID, clientSecret, subscriptionID, tenantID, resourceGroup string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.ClientID = clientID
	config.ClientSecret = clientSecret
	config.SubscriptionID = subscriptionID
	config.TenantID = tenantID
	config.ResourceGroup = resourceGroup

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azure.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azure: the configuration of the DNS provider is nil")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	provider := &DNSProvider{config: config}

	if config.ClientID != "" && config.ClientSecret != "" && config.SubscriptionID != "" && config.TenantID != "" && config.ResourceGroup != "" {
		spt, err := provider.newServicePrincipalToken(azure.PublicCloud.ResourceManagerEndpoint)
		if err != nil {
			return nil, err
		}
		spt.SetSender(config.HTTPClient)
		provider.authorizer = autorest.NewBearerAuthorizer(spt)
	} else {
		var err error
		provider.authorizer, err = auth.NewAuthorizerFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("azure: %v", err)
		}

		// TODO: pass `config.HTTPClient` into authorizer

		if config.SubscriptionID == "" {
			config.SubscriptionID, err = provider.getMetadata("AZURE_SUBSCRIPTION_ID", "subscriptionId")
			if err != nil {
				return nil, fmt.Errorf("azure: %v", err)
			}
		}

		if config.ResourceGroup == "" {
			config.ResourceGroup, err = provider.getMetadata("AZURE_RESOURCE_GROUP", "resourceGroupName")
			if err != nil {
				return nil, fmt.Errorf("azure: %v", err)
			}
		}
	}

	return provider, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	rsc := dns.NewRecordSetsClient(d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(int64(d.config.TTL)),
			TxtRecords: &[]dns.TxtRecord{{Value: &[]string{value}}},
		},
	}

	_, err = rsc.CreateOrUpdate(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, rec, "", "")
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rsc := dns.NewRecordSetsClient(d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	_, err = rsc.Delete(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, "")
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}
	return nil
}

// Checks that azure has a zone for this domain name.
func (d *DNSProvider) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	dc := dns.NewZonesClient(d.config.SubscriptionID)
	dc.Authorizer = d.authorizer

	zone, err := dc.Get(ctx, d.config.ResourceGroup, acme.UnFqdn(authZone))
	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func (d *DNSProvider) newServicePrincipalToken(scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, d.config.TenantID)
	if err != nil {
		return nil, err
	}
	return adal.NewServicePrincipalToken(*oauthConfig, d.config.ClientID, d.config.ClientSecret, scope)
}

// Returns the relative record to the domain
func toRelativeRecord(domain, zone string) string {
	return acme.UnFqdn(strings.TrimSuffix(domain, zone))
}

// Fetches metadata from environment or he instance metadata service
// borrowed from https://github.com/Microsoft/azureimds/blob/master/imdssample.go
func (d *DNSProvider) getMetadata(envVar, field string) (string, error) {

	// first check environment
	if envVar != "" {
		value := env.GetOrDefaultString(envVar, "")
		if value != "" {
			return value, nil
		}
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/metadata/instance/compute/%s", metadataEndpoint, field), nil)
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "text")
	q.Add("api-version", "2017-12-01")
	req.URL.RawQuery = q.Encode()

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	return string(respBody[:]), nil
}
