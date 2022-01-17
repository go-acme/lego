package azure

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
	EnvClientID,
	EnvClientSecret,
	EnvSubscriptionID,
	EnvTenantID,
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
				EnvClientID:       "A",
				EnvClientSecret:   "B",
				EnvTenantID:       "C",
				EnvSubscriptionID: "D",
				EnvResourceGroup:  "E",
			},
		},
		{
			desc: "missing client ID",
			envVars: map[string]string{
				EnvClientID:       "",
				EnvClientSecret:   "B",
				EnvTenantID:       "C",
				EnvSubscriptionID: "D",
				EnvResourceGroup:  "E",
			},
			expected: "failed to get SPT from client credentials: parameter 'clientID' cannot be empty",
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

			assert.IsType(t, p.provider, new(dnsProviderPublic))
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
		privateZone    bool
		handler        func(w http.ResponseWriter, r *http.Request)
		expected       string
	}{
		{
			desc:           "success (public)",
			clientID:       "A",
			clientSecret:   "B",
			tenantID:       "C",
			subscriptionID: "D",
			resourceGroup:  "E",
			privateZone:    false,
		},
		{
			desc:           "success (private)",
			clientID:       "A",
			clientSecret:   "B",
			tenantID:       "C",
			subscriptionID: "D",
			resourceGroup:  "E",
			privateZone:    true,
		},
		{
			desc:           "SubscriptionID missing",
			clientID:       "A",
			clientSecret:   "B",
			tenantID:       "C",
			subscriptionID: "",
			resourceGroup:  "",
			expected:       "azure: SubscriptionID is missing",
		},
		{
			desc:           "ResourceGroup missing",
			clientID:       "A",
			clientSecret:   "B",
			tenantID:       "C",
			subscriptionID: "D",
			resourceGroup:  "",
			expected:       "azure: ResourceGroup is missing",
		},
		{
			desc:           "use metadata",
			clientID:       "A",
			clientSecret:   "B",
			tenantID:       "C",
			subscriptionID: "",
			resourceGroup:  "",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, err := w.Write([]byte("foo"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
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
			config.PrivateZone = test.privateZone

			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			if test.handler == nil {
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
			} else {
				mux.HandleFunc("/", test.handler)
			}
			config.MetadataEndpoint = server.URL

			p, err := NewDNSProviderConfig(config)

			if test.expected != "" {
				require.EqualError(t, err, test.expected)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			require.NotNil(t, p.provider)

			if test.privateZone {
				assert.IsType(t, p.provider, new(dnsProviderPrivate))
			} else {
				assert.IsType(t, p.provider, new(dnsProviderPublic))
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
