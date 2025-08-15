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
