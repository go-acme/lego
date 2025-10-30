package webnamesca

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIUser, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUser: "user",
				EnvAPIKey:  "secret",
			},
		},
		{
			desc: "missing EnvAPIUser",
			envVars: map[string]string{
				EnvAPIUser: "",
				EnvAPIKey:  "secret",
			},
			expected: "webnamesca: some credentials information are missing: WEBNAMESCA_API_USER",
		},
		{
			desc: "missing EnvAPIKey",
			envVars: map[string]string{
				EnvAPIUser: "user",
				EnvAPIKey:  "",
			},
			expected: "webnamesca: some credentials information are missing: WEBNAMESCA_API_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "webnamesca: some credentials information are missing: WEBNAMESCA_API_USER,WEBNAMESCA_API_KEY",
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
		apiUser  string
		apiKey   string
		expected string
	}{
		{
			desc:    "success",
			apiUser: "user",
			apiKey:  "secret",
		},
		{
			desc:     "missing apiUser",
			apiKey:   "secret",
			expected: "webnamesca: credentials missing",
		},
		{
			desc:     "missing apiKey",
			apiUser:  "user",
			expected: "webnamesca: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "webnamesca: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUser = test.apiUser
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
			config.APIUser = "user"
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
			With("API-User", "user").
			With("API-Key", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /domains/example.com/add-txt-record",
			servermock.ResponseFromInternal("add_txt_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("hostName", "_acme-challenge.example.com").
				With("txt", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /domains/example.com/delete-txt-record",
			servermock.ResponseFromInternal("delete_txt_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("hostName", "_acme-challenge.example.com").
				With("txt", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")).
		Build(t)

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
