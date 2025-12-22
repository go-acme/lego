package ispconfigddns

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvServerURL, EnvToken).
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
				EnvServerURL: "https://example.com",
				EnvToken:     "secret",
			},
		},
		{
			desc: "missing server URL",
			envVars: map[string]string{
				EnvServerURL: "",
				EnvToken:     "secret",
			},
			expected: "ispconfig (DDNS module): some credentials information are missing: ISPCONFIG_DDNS_SERVER_URL",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvServerURL: "https://example.com",
				EnvToken:     "",
			},
			expected: "ispconfig (DDNS module): some credentials information are missing: ISPCONFIG_DDNS_TOKEN",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "ispconfig (DDNS module): some credentials information are missing: ISPCONFIG_DDNS_SERVER_URL,ISPCONFIG_DDNS_TOKEN",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		serverURL string
		token     string
		expected  string
	}{
		{
			desc:      "success",
			serverURL: "https://example.com",
			token:     "secret",
		},
		{
			desc:      "missing server URL",
			serverURL: "",
			token:     "secret",
			expected:  "ispconfig (DDNS module): missing server URL",
		},
		{
			desc:      "missing token",
			serverURL: "https://example.com",
			token:     "",
			expected:  "ispconfig (DDNS module): missing token",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ServerURL = test.serverURL
			config.Token = test.token

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		config := NewDefaultConfig()
		config.HTTPClient = server.Client()
		config.Token = "secret"
		config.ServerURL = server.URL

		return NewDNSProviderConfig(config)
	},
		servermock.CheckHeader().
			WithBasicAuth("anonymous", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /ddns/update.php",
			servermock.DumpRequest(),
			servermock.CheckQueryParameter().Strict().
				With("action", "add").
				With("zone", "example.com").
				With("type", "TXT").
				With("record", "_acme-challenge.example.com.").
				With("data", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /ddns/update.php",
			servermock.DumpRequest(),
			servermock.CheckQueryParameter().Strict().
				With("action", "delete").
				With("zone", "example.com").
				With("type", "TXT").
				With("record", "_acme-challenge.example.com.").
				With("data", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
