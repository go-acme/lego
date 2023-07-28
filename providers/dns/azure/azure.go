// Package azure implements a DNS provider for solving the DNS-01 challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/go-autorest/autorest"
	aazure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
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
	EnvPrivateZone      = envNamespace + "PRIVATE_ZONE"

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
	PrivateZone    bool

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
	provider challenge.ProviderTimeout
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials can be passed in the environment variables:
// AZURE_ENVIRONMENT, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET,
// AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
// If the credentials are _not_ set via the environment,
// then it will attempt to get a bearer token via the instance metadata service.
// see: https://github.com/Azure/go-autorest/blob/v10.14.0/autorest/azure/auth/auth.go#L38-L42
// Deprecated: use azuredns instead.
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
	config.PrivateZone = env.GetOrDefaultBool(EnvPrivateZone, false)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azure.
// Deprecated: use azuredns instead.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azure: the configuration of the DNS provider is nil")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 5 * time.Second}
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

	if config.PrivateZone {
		return &DNSProvider{provider: &dnsProviderPrivate{config: config, authorizer: authorizer}}, nil
	}

	return &DNSProvider{provider: &dnsProviderPublic{config: config, authorizer: authorizer}}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.provider.Timeout()
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	return d.provider.Present(domain, token, keyAuth)
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.provider.CleanUp(domain, token, keyAuth)
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

// Fetches metadata from environment or the instance metadata service.
// borrowed from https://github.com/Microsoft/azureimds/blob/master/imdssample.go
func getMetadata(config *Config, field string) (string, error) {
	metadataEndpoint := config.MetadataEndpoint
	if metadataEndpoint == "" {
		metadataEndpoint = defaultMetadataEndpoint
	}

	endpoint, err := url.JoinPath(metadataEndpoint, "metadata", "instance", "compute", field)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
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
		return "", errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	return string(raw), nil
}
