package azuredns

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

const (
	authMethodEnv      = "env"
	authMethodWLI      = "wli"
	authMethodMSI      = "msi"
	authMethodCLI      = "cli"
	authMethodOIDC     = "oidc"
	authMethodPipeline = "pipeline"
)

//nolint:gocyclo // The complexity is related to the number of possible configurations.
func getCredentials(config *Config) (azcore.TokenCredential, error) {
	clientOptions := azcore.ClientOptions{Cloud: config.Environment}

	switch strings.ToLower(config.AuthMethod) {
	case authMethodEnv:
		if config.ClientID != "" && config.ClientSecret != "" && config.TenantID != "" {
			return azidentity.NewClientSecretCredential(config.TenantID, config.ClientID, config.ClientSecret,
				&azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions})
		}

		return azidentity.NewEnvironmentCredential(&azidentity.EnvironmentCredentialOptions{ClientOptions: clientOptions})

	case authMethodWLI:
		return azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{ClientOptions: clientOptions})

	case authMethodMSI:
		cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{ClientOptions: clientOptions})
		if err != nil {
			return nil, err
		}

		return &timeoutTokenCredential{cred: cred, timeout: config.AuthMSITimeout}, nil

	case authMethodCLI:
		var credOptions *azidentity.AzureCLICredentialOptions
		if config.TenantID != "" {
			credOptions = &azidentity.AzureCLICredentialOptions{TenantID: config.TenantID}
		}

		return azidentity.NewAzureCLICredential(credOptions)

	case authMethodOIDC:
		err := checkOIDCConfig(config)
		if err != nil {
			return nil, err
		}

		return azidentity.NewClientAssertionCredential(config.TenantID, config.ClientID, getOIDCAssertion(config), &azidentity.ClientAssertionCredentialOptions{ClientOptions: clientOptions})

	case authMethodPipeline:
		err := checkPipelineConfig(config)
		if err != nil {
			return nil, err
		}

		// Uses the env var `SYSTEM_OIDCREQUESTURI`,
		// but the constant is not exported,
		// and there is no way to set it programmatically.
		// https://github.com/Azure/azure-sdk-for-go/blob/aae2fb75ffccafc669db72bebc3c1a66332f48d7/sdk/azidentity/azure_pipelines_credential.go#L22
		// https://github.com/Azure/azure-sdk-for-go/blob/aae2fb75ffccafc669db72bebc3c1a66332f48d7/sdk/azidentity/azure_pipelines_credential.go#L79

		return azidentity.NewAzurePipelinesCredential(config.TenantID, config.ClientID, config.ServiceConnectionID, config.SystemAccessToken, &azidentity.AzurePipelinesCredentialOptions{ClientOptions: clientOptions})

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

func checkPipelineConfig(config *Config) error {
	if config.ServiceConnectionID == "" {
		return errors.New("azuredns: ServiceConnectionID is missing")
	}

	if config.SystemAccessToken == "" {
		return errors.New("azuredns: SystemAccessToken is missing")
	}

	return nil
}
