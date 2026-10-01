package enum

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIToken, EnvCredentials).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success (API token)",
			envVars: map[string]string{
				EnvAPIToken: "secret",
			},
		},
		{
			desc: "success (mapping)",
			envVars: map[string]string{
				EnvCredentials: "example.com:secret",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIToken:    "",
				EnvCredentials: "",
			},
			expected: "enum: some credentials information are missing: ENUM_API_TOKEN or ENUM_CREDENTIALS",
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
		desc        string
		credentials map[string]string
		expected    string
	}{
		{
			desc: "success",
			credentials: map[string]string{
				"example.com": "secret",
			},
		},
		{
			desc:     "missing credentials",
			expected: "enum: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Credentials = test.credentials

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
			config.Credentials = map[string]string{
				TokenZoneWildcard: "secret",
			}
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
		Route("POST /enum.api.v1.ProjectService/ListProjects",
			servermock.ResponseFromInternal("ListProjects.json"),
		).
		Route("POST /enum.api.v1.DnsService/GetZoneByName",
			servermock.ResponseFromInternal("GetZoneByName.json"),
			servermock.CheckRequestJSONBodyFromInternal("GetZoneByName-request.json"),
		).
		Route("POST /enum.api.v1.DnsService/AddRecordSetValue",
			servermock.ResponseFromInternal("AddRecordSetValue.json"),
			servermock.CheckRequestJSONBodyFromInternal("AddRecordSetValue-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /enum.api.v1.ProjectService/ListProjects",
			servermock.ResponseFromInternal("ListProjects.json"),
			servermock.CheckRequestBody(`{}`),
		).
		Route("POST /enum.api.v1.DnsService/GetZoneByName",
			servermock.ResponseFromInternal("GetZoneByName.json"),
			servermock.CheckRequestJSONBodyFromInternal("GetZoneByName-request.json"),
		).
		Route("POST /enum.api.v1.DnsService/RemoveRecordSetValue",
			servermock.ResponseFromInternal("RemoveRecordSetValue.json"),
			servermock.CheckRequestJSONBodyFromInternal("RemoveRecordSetValue-request.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func Test_getToken(t *testing.T) {
	testCases := []struct {
		desc     string
		zone     string
		cfg      *Config
		expected string
	}{
		{
			desc: "explicit zone",
			zone: "example.com",
			cfg: &Config{
				Credentials: map[string]string{
					"example.org":     "secretA",
					"example.com":     "secretB",
					TokenZoneWildcard: "secretC",
				},
			},
			expected: "secretB",
		},
		{
			desc: "wildcard zone",
			zone: TokenZoneWildcard,
			cfg: &Config{
				Credentials: map[string]string{
					"example.org":     "secretA",
					"example.com":     "secretB",
					TokenZoneWildcard: "secretC",
				},
			},
			expected: "secretC",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tok, err := getToken(test.cfg, test.zone)
			require.NoError(t, err)

			assert.Equal(t, test.expected, tok)
		})
	}
}

func Test_getCredentials(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected map[string]string
	}{
		{
			desc: "API token",
			envVars: map[string]string{
				EnvAPIToken:    "secret",
				EnvCredentials: "example.com:secretA,example.org:secretB",
			},
			expected: map[string]string{
				TokenZoneWildcard: "secret",
			},
		},
		{
			desc: "mapping credentials",
			envVars: map[string]string{
				EnvCredentials: "example.com:secretA,example.org:secretB",
			},
			expected: map[string]string{
				"example.com": "secretA",
				"example.org": "secretB",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			credentials, err := getCredentials()
			require.NoError(t, err)

			assert.Equal(t, test.expected, credentials)
		})
	}
}
