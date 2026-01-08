package todaynic

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAuthUserID, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAuthUserID: "user123",
				EnvAPIKey:     "secret",
			},
		},
		{
			desc: "missing user ID",
			envVars: map[string]string{
				EnvAuthUserID: "",
				EnvAPIKey:     "secret",
			},
			expected: "todaynic: some credentials information are missing: TODAYNIC_AUTH_USER_ID",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAuthUserID: "user123",
				EnvAPIKey:     "",
			},
			expected: "todaynic: some credentials information are missing: TODAYNIC_API_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "todaynic: some credentials information are missing: TODAYNIC_AUTH_USER_ID,TODAYNIC_API_KEY",
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
		desc       string
		authUserID string
		apiKey     string
		expected   string
	}{
		{
			desc:       "success",
			authUserID: "user123",
			apiKey:     "secret",
		},
		{
			desc:     "missing user ID",
			apiKey:   "secret",
			expected: "todaynic: credentials missing",
		},
		{
			desc:       "missing API key",
			authUserID: "user123",
			expected:   "todaynic: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "todaynic: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthUserID = test.authUserID
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
			config.AuthUserID = "user123"
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
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/dns/add-domain-record.json",
			servermock.ResponseFromInternal("add_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("Domain", "example.com").
				With("Host", "_acme-challenge").
				With("Type", "TXT").
				With("Value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("Ttl", "600").
				With("auth-userid", "user123").
				With("api-key", "secret"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/dns/delete-domain-record.json",
			servermock.ResponseFromInternal("add_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("Id", "123").
				With("auth-userid", "user123").
				With("api-key", "secret"),
		).
		Build(t)

	provider.recordIDs["abc"] = 123

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
