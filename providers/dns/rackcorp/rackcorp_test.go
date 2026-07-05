package rackcorp

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIUUID, EnvAPISecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUUID:   "3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36",
				EnvAPISecret: "secret",
			},
		},
		{
			desc: "missing API UUID",
			envVars: map[string]string{
				EnvAPIUUID:   "",
				EnvAPISecret: "secret",
			},
			expected: "rackcorp: some credentials information are missing: RACKCORP_API_UUID",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvAPIUUID:   "3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36",
				EnvAPISecret: "",
			},
			expected: "rackcorp: some credentials information are missing: RACKCORP_API_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "rackcorp: some credentials information are missing: RACKCORP_API_UUID,RACKCORP_API_SECRET",
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
		desc      string
		apiUUID   string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiUUID:   "3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36",
			apiSecret: "secret",
		},
		{
			desc:      "missing API UUID",
			apiSecret: "secret",
			expected:  "rackcorp: credentials missing",
		},
		{
			desc:     "missing API secret",
			apiUUID:  "3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36",
			expected: "rackcorp: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "rackcorp: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUUID = test.apiUUID
			config.APISecret = test.apiSecret

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

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIUUID = "3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36"
			config.APISecret = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/domain",
			servermock.ResponseFromInternal("get_domains.json"),
		).
		Route("POST /dns/records",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /dns/records/456",
			servermock.ResponseFromInternal("delete_record.json"),
		).
		Build(t)

	provider.recordIDs["abc"] = 456

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
