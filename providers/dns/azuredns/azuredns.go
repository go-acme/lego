// Package azuredns implements a DNS provider for solving the DNS-01 challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azuredns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-acme/lego/v4/challenge"
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

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	EnvUseEnvVars = envNamespace + "USE_ENV_VARS"
	EnvUseWli     = envNamespace + "USE_WLI"
	EnvUseMsi     = envNamespace + "USE_MSI"
	EnvUseCli     = envNamespace + "USE_CLI"
	EnvMsiTimeout = envNamespace + "MSI_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	SubscriptionID string
	ResourceGroup  string
	PrivateZone    bool

	Environment cloud.Configuration

	// optional if using default Azure credentials
	ClientID     string
	ClientSecret string
	TenantID     string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int

	UseEnvVars bool
	UseWli     bool
	UseMsi     bool
	UseCli     bool
	MsiTimeout time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 60),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 2*time.Second),
		Environment:        cloud.AzurePublic,
		UseEnvVars:         env.GetOrDefaultBool(EnvUseEnvVars, false),
		UseWli:             env.GetOrDefaultBool(EnvUseWli, false),
		UseMsi:             env.GetOrDefaultBool(EnvUseMsi, false),
		UseCli:             env.GetOrDefaultBool(EnvUseCli, false),
		MsiTimeout:         env.GetOrDefaultSecond(EnvMsiTimeout, 2*time.Second),
	}
}

// timeoutWrapper wraps a ManagedIdentityCredential to add a timeout.
type timeoutWrapper struct {
	cred    azcore.TokenCredential
	timeout time.Duration
}

// timeoutWrapper GetToken implements the azcore.TokenCredential interface.
func (w *timeoutWrapper) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	var tk azcore.AccessToken
	var err error
	if w.timeout > 0 {
		c, cancel := context.WithTimeout(ctx, w.timeout)
		defer cancel()
		tk, err = w.cred.GetToken(c, opts)
		if ce := c.Err(); errors.Is(ce, context.DeadlineExceeded) {
			err = azidentity.NewCredentialUnavailableError("managed identity timed out")
		} else {
			w.timeout = 0
		}
	} else {
		tk, err = w.cred.GetToken(ctx, opts)
	}
	return tk, err
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

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azure.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azuredns: the configuration of the DNS provider is nil")
	}

	credentials, err := getCredentials(config)
	if err != nil {
		return nil, errors.New("azuredns: Unable to retrieve valid credentials")
	}

	if config.SubscriptionID == "" {
		return nil, errors.New("azuredns: SubscriptionID is missing")
	}

	if config.ResourceGroup == "" {
		return nil, errors.New("azuredns: ResourceGroup is missing")
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

//gocyclo:ignore
func getCredentials(config *Config) (azcore.TokenCredential, error) {
	var creds []azcore.TokenCredential
	clientOptions := azcore.ClientOptions{Cloud: config.Environment}

	var err error
	var cred azcore.TokenCredential
	if config.UseEnvVars {
		if config.ClientID != "" && config.ClientSecret != "" && config.TenantID != "" {
			cred, err = azidentity.NewClientSecretCredential(config.TenantID, config.ClientID, config.ClientSecret,
				&azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions})
			if err == nil {
				creds = append(creds, cred)
			}
		} else {
			cred, err = azidentity.NewEnvironmentCredential(&azidentity.EnvironmentCredentialOptions{ClientOptions: clientOptions})
			if err == nil {
				creds = append(creds, cred)
			}
		}
	}

	if config.UseWli {
		cred, err = azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{ClientOptions: clientOptions})
		if err == nil {
			creds = append(creds, cred)
		}
	}

	if config.UseMsi {
		cred, err = azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{ClientOptions: clientOptions})
		if err == nil {
			creds = append(creds, &timeoutWrapper{cred, time.Second})
		}
	}

	if config.UseCli {
		cred, err = azidentity.NewAzureCLICredential(nil)
		if err == nil {
			creds = append(creds, cred)
		}
	}

	if len(creds) != 0 {
		return azidentity.NewChainedTokenCredential(creds, nil)
	}

	return azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{ClientOptions: clientOptions})
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

func deref[T string | int | int32 | int64](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
