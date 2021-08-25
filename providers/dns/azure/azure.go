// Package azure implements a DNS provider for solving the DNS-01 challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-09-01/dns"
	"github.com/Azure/go-autorest/autorest"
	aazure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const defaultMetadataEndpoint = "http://169.254.169.254"

// Environment variables names.
const (
	envNamespace = "AZURE_"

	EnvEnvironment      = envNamespace + "ENVIRONMENT"
	EnvMetadataEndpoint = envNamespace + "METADATA_ENDPOINT"
	EnvSubscriptionID   = envNamespace + "SUBSCRIPTION_ID"
	EnvResourceGroup    = envNamespace + "RESOURCE_GROUP"
	EnvTenantID         = envNamespace + "TENANT_ID"
	EnvClientID         = envNamespace + "CLIENT_ID"
	EnvClientSecret     = envNamespace + "CLIENT_SECRET"
	EnvZoneName         = envNamespace + "ZONE_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	// optional if using instance metadata service
	ClientID     string
	ClientSecret string
	TenantID     string

	SubscriptionID string
	ResourceGroup  string

	MetadataEndpoint        string
	ResourceManagerEndpoint string
	ActiveDirectoryEndpoint string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                     env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout:      env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:         env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		MetadataEndpoint:        env.GetOrFile(EnvMetadataEndpoint),
		ResourceManagerEndpoint: aazure.PublicCloud.ResourceManagerEndpoint,
		ActiveDirectoryEndpoint: aazure.PublicCloud.ActiveDirectoryEndpoint,
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config     *Config
	authorizer autorest.Authorizer
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials can be passed in the environment variables:
// AZURE_ENVIRONMENT, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET,
// AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
// If the credentials are _not_ set via the environment,
// then it will attempt to get a bearer token via the instance metadata service.
// see: https://github.com/Azure/go-autorest/blob/v10.14.0/autorest/azure/auth/auth.go#L38-L42
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	environmentName := env.GetOrFile(EnvEnvironment)
	if environmentName != "" {
		var environment aazure.Environment
		switch environmentName {
		case "china":
			environment = aazure.ChinaCloud
		case "german":
			environment = aazure.GermanCloud
		case "public":
			environment = aazure.PublicCloud
		case "usgovernment":
			environment = aazure.USGovernmentCloud
		default:
			return nil, fmt.Errorf("azure: unknown environment %s", environmentName)
		}

		config.ResourceManagerEndpoint = environment.ResourceManagerEndpoint
		config.ActiveDirectoryEndpoint = environment.ActiveDirectoryEndpoint
	}

	config.SubscriptionID = env.GetOrFile(EnvSubscriptionID)
	config.ResourceGroup = env.GetOrFile(EnvResourceGroup)
	config.ClientSecret = env.GetOrFile(EnvClientSecret)
	config.ClientID = env.GetOrFile(EnvClientID)
	config.TenantID = env.GetOrFile(EnvTenantID)

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

	authorizer, err := getAuthorizer(config)
	if err != nil {
		return nil, err
	}

	if config.SubscriptionID == "" {
		subsID, err := getMetadata(config, "subscriptionId")
		if err != nil {
			return nil, fmt.Errorf("azure: %w", err)
		}

		if subsID == "" {
			return nil, errors.New("azure: SubscriptionID is missing")
		}
		config.SubscriptionID = subsID
	}

	if config.ResourceGroup == "" {
		resGroup, err := getMetadata(config, "resourceGroupName")
		if err != nil {
			return nil, fmt.Errorf("azure: %w", err)
		}

		if resGroup == "" {
			return nil, errors.New("azure: ResourceGroup is missing")
		}
		config.ResourceGroup = resGroup
	}

	return &DNSProvider{config: config, authorizer: authorizer}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}

	rsc := dns.NewRecordSetsClientWithBaseURI(d.config.ResourceManagerEndpoint, d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	relative := toRelativeRecord(fqdn, dns01.ToFqdn(zone))

	// Get existing record set
	rset, err := rsc.Get(ctx, d.config.ResourceGroup, zone, relative, dns.TXT)
	if err != nil {
		var detailed autorest.DetailedError
		if !errors.As(err, &detailed) || detailed.StatusCode != http.StatusNotFound {
			return fmt.Errorf("azure: %w", err)
		}
	}

	// Construct unique TXT records using map
	uniqRecords := map[string]struct{}{value: {}}
	if rset.RecordSetProperties != nil && rset.TxtRecords != nil {
		for _, txtRecord := range *rset.TxtRecords {
			// Assume Value doesn't contain multiple strings
			if txtRecord.Value != nil && len(*txtRecord.Value) > 0 {
				uniqRecords[(*txtRecord.Value)[0]] = struct{}{}
			}
		}
	}

	var txtRecords []dns.TxtRecord
	for txt := range uniqRecords {
		txtRecords = append(txtRecords, dns.TxtRecord{Value: &[]string{txt}})
	}

	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(int64(d.config.TTL)),
			TxtRecords: &txtRecords,
		},
	}

	_, err = rsc.CreateOrUpdate(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, rec, "", "")
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}

	relative := toRelativeRecord(fqdn, dns01.ToFqdn(zone))
	rsc := dns.NewRecordSetsClientWithBaseURI(d.config.ResourceManagerEndpoint, d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	_, err = rsc.Delete(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, "")
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}
	return nil
}

// Checks that azure has a zone for this domain name.
func (d *DNSProvider) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	if zone := env.GetOrFile(EnvZoneName); zone != "" {
		return zone, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
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

// Returns the relative record to the domain.
func toRelativeRecord(domain, zone string) string {
	return dns01.UnFqdn(strings.TrimSuffix(domain, zone))
}

func getAuthorizer(config *Config) (autorest.Authorizer, error) {
	if config.ClientID != "" && config.ClientSecret != "" && config.TenantID != "" {
		credentialsConfig := auth.ClientCredentialsConfig{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			TenantID:     config.TenantID,
			Resource:     config.ResourceManagerEndpoint,
			AADEndpoint:  config.ActiveDirectoryEndpoint,
		}

		spToken, err := credentialsConfig.ServicePrincipalToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get oauth token from client credentials: %w", err)
		}

		spToken.SetSender(config.HTTPClient)

		return autorest.NewBearerAuthorizer(spToken), nil
	}

	return auth.NewAuthorizerFromEnvironment()
}

// Fetches metadata from environment or he instance metadata service.
// borrowed from https://github.com/Microsoft/azureimds/blob/master/imdssample.go
func getMetadata(config *Config, field string) (string, error) {
	metadataEndpoint := config.MetadataEndpoint
	if metadataEndpoint == "" {
		metadataEndpoint = defaultMetadataEndpoint
	}

	resource := fmt.Sprintf("%s/metadata/instance/compute/%s", metadataEndpoint, field)
	req, err := http.NewRequest(http.MethodGet, resource, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "text")
	q.Add("api-version", "2017-12-01")
	req.URL.RawQuery = q.Encode()

	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
