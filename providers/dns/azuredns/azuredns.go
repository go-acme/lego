// Package azuredns implements a DNS provider for solving the DNS-01 challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azuredns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	EnvOidcToken         = envNamespace + "OIDC_TOKEN"
	EnvOidcTokenFilePath = envNamespace + "OIDC_TOKEN_FILE_PATH"
	EnvOidcRequestURL    = envNamespace + "OIDC_REQUEST_URL"
	EnvOidcRequestToken  = envNamespace + "OIDC_REQUEST_TOKEN"

	EnvAuthMethod     = envNamespace + "AUTH_METHOD"
	EnvAuthMSITimeout = envNamespace + "AUTH_MSI_TIMEOUT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"

	envGihubActions = "ACTIONS_"

	EnvGithubOidcRequestURL   = envGihubActions + "ID_TOKEN_REQUEST_URL"
	EnvGithubOidcRequestToken = envGihubActions + "ID_TOKEN_REQUEST_TOKEN"
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

	OidcToken         string
	OidcTokenFilePath string
	OidcRequestURL    string
	OidcRequestToken  string

	AuthMethod     string
	AuthMSITimeout time.Duration

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
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

	config.OidcToken = env.GetOrFile(EnvOidcToken)
	config.OidcTokenFilePath = env.GetOrFile(EnvOidcTokenFilePath)

	oidcRequest, _ := env.GetWithFallback(
		[]string{EnvOidcRequestURL, EnvGithubOidcRequestURL},
		[]string{EnvOidcRequestToken, EnvGithubOidcRequestToken},
	)

	config.OidcRequestURL = oidcRequest[EnvOidcRequestURL]
	config.OidcRequestToken = oidcRequest[EnvOidcRequestToken]

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
		return azidentity.NewAzureCLICredential(nil)

	case "oidc":
		return getOidcCredentials(config, clientOptions)

	default:
		return azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{ClientOptions: clientOptions})
	}
}

//nolint:gocyclo
func getOidcCredentials(config *Config, clientOptions azcore.ClientOptions) (azcore.TokenCredential, error) {
	if config.TenantID == "" {
		return nil, fmt.Errorf("azuredns: TenantId is missing")
	}

	if config.ClientID == "" {
		return nil, fmt.Errorf("azuredns: ClientId is missing")
	}

	if config.OidcToken == "" && config.OidcTokenFilePath == "" && (config.OidcRequestURL == "" || config.OidcRequestToken == "") {
		return nil, fmt.Errorf("azuredns: OidcToken, OidcTokenFilePath or OidcRequestURL and OidcRequestToken must be set")
	}

	getAssertion := func(ctx context.Context) (string, error) {
		var token string
		if config.OidcToken != "" {
			token = strings.TrimSpace(config.OidcToken)
		}

		if config.OidcTokenFilePath != "" {
			fileTokenRaw, err := os.ReadFile(config.OidcTokenFilePath)
			if err != nil {
				return "", fmt.Errorf("azuredns: error retrieving token file with path %s: %w", config.OidcTokenFilePath, err)
			}

			fileToken := strings.TrimSpace(string(fileTokenRaw))
			if config.OidcToken != fileToken {
				return "", fmt.Errorf("azuredns: token file with path %s does not match token from environment variable", config.OidcTokenFilePath)
			}

			token = fileToken
		}

		if token == "" && config.OidcRequestURL != "" && config.OidcRequestToken != "" {
			return requestToken(config)
		}

		return token, nil
	}

	return azidentity.NewClientAssertionCredential(config.TenantID, config.ClientID, getAssertion,
		&azidentity.ClientAssertionCredentialOptions{ClientOptions: clientOptions})
}

func requestToken(config *Config) (string, error) {
	req, err := http.NewRequest(http.MethodGet, config.OidcRequestURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("azuredns: failed to build OIDC request: %w", err)
	}

	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return "", fmt.Errorf("azuredns: cannot parse OIDC request URL query")
	}

	if query.Get("audience") == "" {
		query.Set("audience", "api://AzureADTokenExchange")
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.OidcRequestToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("azuredns: cannot request OIDC token: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("azuredns: cannot parse OIDC token response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusNoContent {
		return "", fmt.Errorf("azuredns: OIDC token request received HTTP status %d with response: %s", resp.StatusCode, body)
	}

	var returnedToken struct {
		Count *int    `json:"count"`
		Value *string `json:"value"`
	}
	if err := json.Unmarshal(body, &returnedToken); err != nil {
		return "", fmt.Errorf("azuredns: cannot unmarshal OIDC token response: %w", err)
	}

	return *returnedToken.Value, nil
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

func deref[T string | int | int32 | int64](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
