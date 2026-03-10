package excedo

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIURL, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIURL: "https://example.com",
				EnvAPIKey: "secret",
			},
		},
		{
			desc: "missing the API key",
			envVars: map[string]string{
				EnvAPIURL: "https://example.com",
				EnvAPIKey: "",
			},
			expected: "excedo: some credentials information are missing: EXCEDO_API_KEY",
		},
		{
			desc: "missing the API URL",
			envVars: map[string]string{
				EnvAPIURL: "",
				EnvAPIKey: "secret",
			},
			expected: "excedo: some credentials information are missing: EXCEDO_API_URL",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "excedo: some credentials information are missing: EXCEDO_API_URL,EXCEDO_API_KEY",
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
		apiURL   string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiURL: "https://example.com",
			apiKey: "secret",
		},
		{
			desc:     "missing the API key",
			apiURL:   "https://example.com",
			expected: "excedo: credentials missing",
		},
		{
			desc:     "missing the API URL",
			apiKey:   "secret",
			expected: "excedo: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "excedo: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIURL = test.apiURL
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
			config.APIURL = server.URL
			config.APIKey = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /authenticate/login/",
			servermock.ResponseFromInternal("login.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret"),
		).
		Route("POST /dns/addrecord/",
			servermock.ResponseFromInternal("addrecord.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer session-token"),
			servermock.CheckForm().Strict().
				With("content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("domainName", "example.com").
				With("name", "_acme-challenge").
				With("ttl", "60").
				With("type", "TXT"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /authenticate/login/",
			servermock.ResponseFromInternal("login.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret"),
		).
		Route("POST /dns/deleterecord/",
			servermock.ResponseFromInternal("deleterecord.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer session-token"),
		).
		Build(t)

	provider.records["abc"] = 19695822

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
