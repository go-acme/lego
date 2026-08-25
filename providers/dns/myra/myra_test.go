package myra

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvAPISecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey:    "key",
				EnvAPISecret: "secret",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvAPISecret: "secret",
			},
			expected: "myra: some credentials information are missing: MYRA_API_KEY",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvAPIKey:    "key",
				EnvAPISecret: "",
			},
			expected: "myra: some credentials information are missing: MYRA_API_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "myra: some credentials information are missing: MYRA_API_KEY,MYRA_API_SECRET",
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
		apiKey    string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "key",
			apiSecret: "secret",
		},
		{
			desc:      "missing API key",
			expected:  "myra: missing API credentials",
			apiSecret: "secret",
		},
		{
			desc:     "missing API secret",
			expected: "myra: missing API credentials",
			apiKey:   "key",
		},
		{
			desc:     "missing credentials",
			expected: "myra: missing API credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
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
			config.APIKey = "key"
			config.APISecret = "secret"

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL = server.URL + "/%s"

			return p, nil
		},
		servermock.CheckHeader().
			WithRegexp("Authorization", "MYRA key:.+"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains",
			servermock.ResponseFromFixture("domains.json"),
			servermock.CheckQueryParameter().With("search", "example.com"),
		).
		Route("POST /domain/1/dns-records",
			servermock.ResponseFromFixture("record_create.json"),
			servermock.CheckRequestJSONBodyFromFixture("record_create-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Len(t, provider.domainIDs, 1)
	assert.Len(t, provider.recordIDs, 1)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domain/1/dns-records/2",
			servermock.ResponseFromFixture("record_get.json"),
		).
		Route("DELETE /domain/1/dns-records/2",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("record_delete-request.json"),
		).
		Build(t)

	const tok = "abc"

	provider.domainIDs[tok] = 1
	provider.recordIDs[tok] = 2

	err := provider.CleanUp(t.Context(), "example.com", tok, "123d==")
	require.NoError(t, err)
}
