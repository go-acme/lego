package azure

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	azureLiveTest       bool
	azureClientID       string
	azureClientSecret   string
	azureSubscriptionID string
	azureTenantID       string
	azureResourceGroup  string
	azureDomain         string
)

func init() {
	azureClientID = os.Getenv("AZURE_CLIENT_ID")
	azureClientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	azureSubscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	azureTenantID = os.Getenv("AZURE_TENANT_ID")
	azureResourceGroup = os.Getenv("AZURE_RESOURCE_GROUP")
	azureDomain = os.Getenv("AZURE_DOMAIN")
	if len(azureClientID) > 0 && len(azureClientSecret) > 0 {
		azureLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("AZURE_CLIENT_ID", azureClientID)
	os.Setenv("AZURE_SUBSCRIPTION_ID", azureSubscriptionID)
}

func TestNewDNSProviderValid(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test (requires credentials)")
	}

	defer restoreEnv()
	os.Setenv("AZURE_CLIENT_ID", "")

	config := NewDefaultConfig()
	config.ClientID = azureClientID
	config.ClientSecret = azureClientSecret
	config.SubscriptionID = azureSubscriptionID
	config.TenantID = azureTenantID
	config.ResourceGroup = azureResourceGroup

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test (requires credentials)")
	}

	defer restoreEnv()
	os.Setenv("AZURE_CLIENT_ID", "other")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AZURE_SUBSCRIPTION_ID", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "azure: some credentials information are missing: AZURE_CLIENT_ID,AZURE_CLIENT_SECRET,AZURE_SUBSCRIPTION_ID,AZURE_TENANT_ID,AZURE_RESOURCE_GROUP")
}

func TestLiveAzurePresent(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.ClientID = azureClientID
	config.ClientSecret = azureClientSecret
	config.SubscriptionID = azureSubscriptionID
	config.TenantID = azureTenantID
	config.ResourceGroup = azureResourceGroup

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(azureDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveAzureCleanUp(t *testing.T) {
	if !azureLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.ClientID = azureClientID
	config.ClientSecret = azureClientSecret
	config.SubscriptionID = azureSubscriptionID
	config.TenantID = azureTenantID
	config.ResourceGroup = azureResourceGroup

	provider, err := NewDNSProviderConfig(config)

	time.Sleep(time.Second * 1)

	assert.NoError(t, err)

	err = provider.CleanUp(azureDomain, "", "123d==")
	assert.NoError(t, err)
}
