// Package azuredns implements a DNS provider for solving the DNS-01 challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azuredns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "AZURE_"

	EnvEnvironment    = envNamespace + "ENVIRONMENT"
	EnvSubscriptionID = envNamespace + "SUBSCRIPTION_ID"
	EnvResourceGroup  = envNamespace + "RESOURCE_GROUP"
	EnvZoneName       = envNamespace + "ZONE_NAME"
	EnvPrivateZone    = envNamespace + "PRIVATE_ZONE"

	EnvTenantID     = envNamespace + "TENANT_ID"
	EnvClientID     = envNamespace + "CLIENT_ID"
	EnvClientSecret = envNamespace + "CLIENT_SECRET"

	EnvOIDCToken         = envNamespace + "OIDC_TOKEN"
	EnvOIDCTokenFilePath = envNamespace + "OIDC_TOKEN_FILE_PATH"
	EnvOIDCRequestURL    = envNamespace + "OIDC_REQUEST_URL"
	EnvOIDCRequestToken  = envNamespace + "OIDC_REQUEST_TOKEN"

	EnvAuthMethod     = envNamespace + "AUTH_METHOD"
	EnvAuthMSITimeout = envNamespace + "AUTH_MSI_TIMEOUT"

	EnvServiceDiscoveryFilter = envNamespace + "SERVICEDISCOVERY_FILTER"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	EnvGitHubOIDCRequestURL   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	EnvGitHubOIDCRequestToken = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	ZoneName string

	SubscriptionID string
	ResourceGroup  string
	PrivateZone    bool

	Environment cloud.Configuration

	// optional if using default Azure credentials
	ClientID     string
	ClientSecret string
	TenantID     string

	OIDCToken         string
	OIDCTokenFilePath string
	OIDCRequestURL    string
	OIDCRequestToken  string

	AuthMethod     string
	AuthMSITimeout time.Duration

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client

	ServiceDiscoveryFilter string
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		ZoneName:           env.GetOrFile(EnvZoneName),
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		Environment:        cloud.AzurePublic,
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	provider challenge.ProviderTimeout
}

// NewDNSProvider returns a DNSProvider instance configured for azuredns.
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	environmentName := env.GetOrFile(EnvEnvironment)
	if environmentName != "" {
		switch environmentName {
		case "china":
			config.Environment = cloud.AzureChina
		case "public":
			config.Environment = cloud.AzurePublic
		case "usgovernment":
			config.Environment = cloud.AzureGovernment
		default:
			return nil, fmt.Errorf("azuredns: unknown environment %s", environmentName)
		}
	} else {
		config.Environment = cloud.AzurePublic
	}

	config.SubscriptionID = env.GetOrFile(EnvSubscriptionID)
	config.ResourceGroup = env.GetOrFile(EnvResourceGroup)
	config.PrivateZone = env.GetOrDefaultBool(EnvPrivateZone, false)

	config.ClientID = env.GetOrFile(EnvClientID)
	config.ClientSecret = env.GetOrFile(EnvClientSecret)
	config.TenantID = env.GetOrFile(EnvTenantID)

	config.OIDCToken = env.GetOrFile(EnvOIDCToken)
	config.OIDCTokenFilePath = env.GetOrFile(EnvOIDCTokenFilePath)

	config.ServiceDiscoveryFilter = env.GetOrFile(EnvServiceDiscoveryFilter)

	oidcValues, _ := env.GetWithFallback(
		[]string{EnvOIDCRequestURL, EnvGitHubOIDCRequestURL},
		[]string{EnvOIDCRequestToken, EnvGitHubOIDCRequestToken},
	)

	config.OIDCRequestURL = oidcValues[EnvOIDCRequestURL]
	config.OIDCRequestToken = oidcValues[EnvOIDCRequestToken]

	config.AuthMethod = env.GetOrFile(EnvAuthMethod)
	config.AuthMSITimeout = env.GetOrDefaultSecond(EnvAuthMSITimeout, 2*time.Second)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azure.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azuredns: the configuration of the DNS provider is nil")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	}

	credentials, err := getCredentials(config)
	if err != nil {
		return nil, fmt.Errorf("azuredns: Unable to retrieve valid credentials: %w", err)
	}

	var dnsProvider challenge.ProviderTimeout
	if config.PrivateZone {
		dnsProvider, err = NewDNSProviderPrivate(config, credentials)
		if err != nil {
			return nil, fmt.Errorf("azuredns: %w", err)
		}
	} else {
		dnsProvider, err = NewDNSProviderPublic(config, credentials)
		if err != nil {
			return nil, fmt.Errorf("azuredns: %w", err)
		}
	}

	return &DNSProvider{provider: dnsProvider}, nil
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

func getCredentials(config *Config) (azcore.TokenCredential, error) {
	clientOptions := azcore.ClientOptions{Cloud: config.Environment}

	switch strings.ToLower(config.AuthMethod) {
	case "env":
		if config.ClientID != "" && config.ClientSecret != "" && config.TenantID != "" {
			return azidentity.NewClientSecretCredential(config.TenantID, config.ClientID, config.ClientSecret,
				&azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions})
		}

		return azidentity.NewEnvironmentCredential(&azidentity.EnvironmentCredentialOptions{ClientOptions: clientOptions})

	case "wli":
		return azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{ClientOptions: clientOptions})

	case "msi":
		cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{ClientOptions: clientOptions})
		if err != nil {
			return nil, err
		}

		return &timeoutTokenCredential{cred: cred, timeout: config.AuthMSITimeout}, nil

	case "cli":
		var credOptions *azidentity.AzureCLICredentialOptions
		if config.TenantID != "" {
			credOptions = &azidentity.AzureCLICredentialOptions{TenantID: config.TenantID}
		}
		return azidentity.NewAzureCLICredential(credOptions)

	case "oidc":
		err := checkOIDCConfig(config)
		if err != nil {
			return nil, err
		}

		return azidentity.NewClientAssertionCredential(config.TenantID, config.ClientID, getOIDCAssertion(config), &azidentity.ClientAssertionCredentialOptions{ClientOptions: clientOptions})

	default:
		return azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{ClientOptions: clientOptions})
	}
}

// timeoutTokenCredential wraps a TokenCredential to add a timeout.
type timeoutTokenCredential struct {
	cred    azcore.TokenCredential
	timeout time.Duration
}

// GetToken implements the azcore.TokenCredential interface.
func (w *timeoutTokenCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if w.timeout <= 0 {
		return w.cred.GetToken(ctx, opts)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	tk, err := w.cred.GetToken(ctxTimeout, opts)
	if ce := ctxTimeout.Err(); errors.Is(ce, context.DeadlineExceeded) {
		return tk, azidentity.NewCredentialUnavailableError("managed identity timed out")
	}

	w.timeout = 0

	return tk, err
}

func getZoneName(config *Config, fqdn string) (string, error) {
	if config.ZoneName != "" {
		return config.ZoneName, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone for %s: %w", fqdn, err)
	}

	if authZone == "" {
		return "", errors.New("empty zone name")
	}

	return authZone, nil
}
