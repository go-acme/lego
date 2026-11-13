package nameocean

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "secret",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "nameocean: some credentials information are missing: NAMEOCEAN_API_KEY",
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
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "secret",
		},
		{
			desc:     "missing credentials",
			expected: "nameocean: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
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
			config.APIKey = "secret"
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
			WithAuthorization("Bearer secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /dns/example.com/records",
			servermock.ResponseFromInternal("add_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("add_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]int64{"abc": 1234}, provider.recordIDs)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /dns/example.com/records/TXT/1234",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	provider.recordIDs["abc"] = 1234

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.recordIDs)
}

func Test_roundTTL(t *testing.T) {
	testCases := []struct {
		desc     string
		value    int
		expected string
	}{
		{
			desc:     "negative",
			value:    -5,
			expected: "5m",
		},
		{
			desc:     "zero",
			value:    0,
			expected: "5m",
		},
		{
			desc:     "5 minutes",
			value:    300,
			expected: "5m",
		},
		{
			desc:     "> 5 minutes",
			value:    301,
			expected: "15m",
		},
		{
			desc:     "15 minutes",
			value:    900,
			expected: "15m",
		},
		{
			desc:     "> 15 minutes",
			value:    901,
			expected: "30m",
		},
		{
			desc:     "30 minutes",
			value:    1800,
			expected: "30m",
		},
		{
			desc:     "> 30 minutes",
			value:    1801,
			expected: "1h",
		},
		{
			desc:     "1 hour",
			value:    3600,
			expected: "1h",
		},
		{
			desc:     "> 1 hour",
			value:    3601,
			expected: "1d",
		},
		{
			desc:     "1 day",
			value:    7200,
			expected: "1d",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ttl := roundTTL(test.value)

			assert.Equal(t, test.expected, ttl)
		})
	}
}
