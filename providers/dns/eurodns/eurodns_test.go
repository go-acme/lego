package eurodns

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/eurodns/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvApplicationID, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvApplicationID: "abc",
				EnvAPIKey:        "secret",
			},
		},
		{
			desc: "missing application ID",
			envVars: map[string]string{
				EnvApplicationID: "",
				EnvAPIKey:        "secret",
			},
			expected: "eurodns: some credentials information are missing: EURODNS_APP_ID",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvApplicationID: "",
				EnvAPIKey:        "secret",
			},
			expected: "eurodns: some credentials information are missing: EURODNS_APP_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "eurodns: some credentials information are missing: EURODNS_APP_ID,EURODNS_API_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		appID    string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			appID:  "abc",
			apiKey: "secret",
		},
		{
			desc:     "missing application ID",
			expected: "eurodns: credentials missing",
			apiKey:   "secret",
		},
		{
			desc:     "missing API secret",
			expected: "eurodns: credentials missing",
			appID:    "abc",
		},
		{
			desc:     "missing credentials",
			expected: "eurodns: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ApplicationID = test.appID
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIKey = "secret"
			config.ApplicationID = "abc"
			config.HTTPClient = server.Client()

			provider, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			provider.client.BaseURL, _ = url.Parse(server.URL)

			return provider, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(internal.HeaderAppID, "abc").
			With(internal.HeaderAPIKey, "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /example.com",
			servermock.ResponseFromInternal("zone_get.json"),
		).
		Route("POST /example.com/check",
			servermock.ResponseFromInternal("zone_add_validate_ok.json"),
			servermock.CheckRequestJSONBodyFromInternal("zone_add.json"),
		).
		Route("PUT /example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromInternal("zone_add.json"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /example.com",
			servermock.ResponseFromInternal("zone_add.json"),
		).
		Route("POST /example.com/check",
			servermock.ResponseFromInternal("zone_remove.json"),
			servermock.CheckRequestJSONBodyFromInternal("zone_remove.json"),
		).
		Route("PUT /example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromInternal("zone_remove.json"),
		).
		Build(t)

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
