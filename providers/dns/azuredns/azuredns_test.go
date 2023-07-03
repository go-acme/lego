package azuredns

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEnvironment,
	EnvSubscriptionID,
	EnvResourceGroup).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvEnvironment:    "",
				EnvSubscriptionID: "A",
				EnvResourceGroup:  "B",
			},
		},
		{
			desc: "unknown environment",
			envVars: map[string]string{
				EnvEnvironment:    "test",
				EnvSubscriptionID: "A",
				EnvResourceGroup:  "B",
			},
			expected: "azuredns: unknown environment test",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected != "" {
				require.EqualError(t, err, test.expected)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			require.NotNil(t, p.provider)

			assert.IsType(t, p.provider, new(DNSProviderPublic))
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc           string
		subscriptionID string
		resourceGroup  string
		privateZone    bool
		handler        func(w http.ResponseWriter, r *http.Request)
		expected       string
	}{
		{
			desc:           "success (public)",
			subscriptionID: "A",
			resourceGroup:  "B",
			privateZone:    false,
		},
		{
			desc:           "success (private)",
			subscriptionID: "A",
			resourceGroup:  "B",
			privateZone:    true,
		},
		{
			desc:           "SubscriptionID missing",
			subscriptionID: "",
			resourceGroup:  "",
			expected:       "azuredns: SubscriptionID is missing",
		},
		{
			desc:           "ResourceGroup missing",
			subscriptionID: "A",
			resourceGroup:  "",
			expected:       "azuredns: ResourceGroup is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.SubscriptionID = test.subscriptionID
			config.ResourceGroup = test.resourceGroup
			config.PrivateZone = test.privateZone

			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			if test.handler == nil {
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
			} else {
				mux.HandleFunc("/", test.handler)
			}

			p, err := NewDNSProviderConfig(config)

			if test.expected != "" {
				require.EqualError(t, err, test.expected)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			require.NotNil(t, p.provider)

			if test.privateZone {
				assert.IsType(t, p.provider, new(DNSProviderPrivate))
			} else {
				assert.IsType(t, p.provider, new(DNSProviderPublic))
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
