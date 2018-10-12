package azure

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest              bool
	envTestClientID       string
	envTestClientSecret   string
	envTestSubscriptionID string
	envTestTenantID       string
	envTestResourceGroup  string
	envTestDomain         string
)

func init() {
	envTestClientID = os.Getenv("AZURE_CLIENT_ID")
	envTestClientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	envTestSubscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	envTestTenantID = os.Getenv("AZURE_TENANT_ID")
	envTestResourceGroup = os.Getenv("AZURE_RESOURCE_GROUP")
	envTestDomain = os.Getenv("AZURE_DOMAIN")

	if len(envTestClientID) > 0 && len(envTestClientSecret) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("AZURE_CLIENT_ID", envTestClientID)
	os.Setenv("AZURE_CLIENT_SECRET", envTestClientSecret)
	os.Setenv("AZURE_SUBSCRIPTION_ID", envTestSubscriptionID)
	os.Setenv("AZURE_TENANT_ID", envTestTenantID)
	os.Setenv("AZURE_RESOURCE_GROUP", envTestResourceGroup)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "A",
				"AZURE_CLIENT_SECRET":   "B",
				"AZURE_SUBSCRIPTION_ID": "C",
				"AZURE_TENANT_ID":       "D",
				"AZURE_RESOURCE_GROUP":  "E",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "",
				"AZURE_CLIENT_SECRET":   "",
				"AZURE_SUBSCRIPTION_ID": "",
				"AZURE_TENANT_ID":       "",
				"AZURE_RESOURCE_GROUP":  "",
			},
			expected: "azure: some credentials information are missing: AZURE_CLIENT_ID,AZURE_CLIENT_SECRET,AZURE_SUBSCRIPTION_ID,AZURE_TENANT_ID,AZURE_RESOURCE_GROUP",
		},
		{
			desc: "missing client id",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "",
				"AZURE_CLIENT_SECRET":   "B",
				"AZURE_SUBSCRIPTION_ID": "C",
				"AZURE_TENANT_ID":       "D",
				"AZURE_RESOURCE_GROUP":  "E",
			},
			expected: "azure: some credentials information are missing: AZURE_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "A",
				"AZURE_CLIENT_SECRET":   "",
				"AZURE_SUBSCRIPTION_ID": "C",
				"AZURE_TENANT_ID":       "D",
				"AZURE_RESOURCE_GROUP":  "E",
			},
			expected: "azure: some credentials information are missing: AZURE_CLIENT_SECRET",
		},
		{
			desc: "missing subscription id",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "A",
				"AZURE_CLIENT_SECRET":   "B",
				"AZURE_SUBSCRIPTION_ID": "",
				"AZURE_TENANT_ID":       "D",
				"AZURE_RESOURCE_GROUP":  "E",
			},
			expected: "azure: some credentials information are missing: AZURE_SUBSCRIPTION_ID",
		},
		{
			desc: "missing tenant id",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "A",
				"AZURE_CLIENT_SECRET":   "B",
				"AZURE_SUBSCRIPTION_ID": "C",
				"AZURE_TENANT_ID":       "",
				"AZURE_RESOURCE_GROUP":  "E",
			},
			expected: "azure: some credentials information are missing: AZURE_TENANT_ID",
		},
		{
			desc: "missing resource group",
			envVars: map[string]string{
				"AZURE_CLIENT_ID":       "A",
				"AZURE_CLIENT_SECRET":   "B",
				"AZURE_SUBSCRIPTION_ID": "C",
				"AZURE_TENANT_ID":       "D",
				"AZURE_RESOURCE_GROUP":  "",
			},
			expected: "azure: some credentials information are missing: AZURE_RESOURCE_GROUP",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc           string
		clientID       string
		clientSecret   string
		subscriptionID string
		tenantID       string
		resourceGroup  string
		expected       string
	}{
		{
			desc:           "success",
			clientID:       "A",
			clientSecret:   "B",
			subscriptionID: "C",
			tenantID:       "D",
			resourceGroup:  "E",
		},
		{
			desc:     "missing credentials",
			expected: "azure: some credentials information are missing",
		},
		{
			desc:           "missing client id",
			clientID:       "",
			clientSecret:   "B",
			subscriptionID: "C",
			tenantID:       "D",
			resourceGroup:  "E",
			expected:       "azure: some credentials information are missing",
		},
		{
			desc:           "missing client secret",
			clientID:       "A",
			clientSecret:   "",
			subscriptionID: "C",
			tenantID:       "D",
			resourceGroup:  "E",
			expected:       "azure: some credentials information are missing",
		},
		{
			desc:           "missing subscription id",
			clientID:       "A",
			clientSecret:   "B",
			subscriptionID: "",
			tenantID:       "D",
			resourceGroup:  "E",
			expected:       "azure: some credentials information are missing",
		},
		{
			desc:           "missing tenant id",
			clientID:       "A",
			clientSecret:   "B",
			subscriptionID: "C",
			tenantID:       "",
			resourceGroup:  "E",
			expected:       "azure: some credentials information are missing",
		},
		{
			desc:           "missing resource group",
			clientID:       "A",
			clientSecret:   "B",
			subscriptionID: "C",
			tenantID:       "D",
			resourceGroup:  "",
			expected:       "azure: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("AZURE_CLIENT_ID")
			os.Unsetenv("AZURE_CLIENT_SECRET")
			os.Unsetenv("AZURE_SUBSCRIPTION_ID")
			os.Unsetenv("AZURE_TENANT_ID")
			os.Unsetenv("AZURE_RESOURCE_GROUP")

			config := NewDefaultConfig()
			config.ClientID = test.clientID
			config.ClientSecret = test.clientSecret
			config.SubscriptionID = test.subscriptionID
			config.TenantID = test.tenantID
			config.ResourceGroup = test.resourceGroup

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
