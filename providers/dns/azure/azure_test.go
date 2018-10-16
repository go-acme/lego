package azure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/tester"
)

var envTest = tester.NewEnvTest(
	"AZURE_CLIENT_ID",
	"AZURE_CLIENT_SECRET",
	"AZURE_SUBSCRIPTION_ID",
	"AZURE_TENANT_ID",
	"AZURE_RESOURCE_GROUP").
	WithDomain("AZURE_DOMAIN")

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
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

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
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
