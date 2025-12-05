package syse

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvCredentials).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvCredentials: "example.org:123",
			},
		},
		{
			desc: "success multiple domains",
			envVars: map[string]string{
				EnvCredentials: "example.org:123,example.com:456,example.net:789",
			},
		},
		{
			desc: "invalid credentials",
			envVars: map[string]string{
				EnvCredentials: ",",
			},
			expected: `syse: credentials: incorrect pair: `,
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvCredentials: "example.org:",
			},
			expected: `syse: missing password: "example.org:"`,
		},
		{
			desc: "missing domain",
			envVars: map[string]string{
				EnvCredentials: ":123",
			},
			expected: `syse: missing domain: ":123"`,
		},
		{
			desc: "invalid credentials, partial",
			envVars: map[string]string{
				EnvCredentials: "example.org:123,example.net",
			},
			expected: "syse: credentials: incorrect pair: example.net",
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvCredentials: "",
			},
			expected: "syse: some credentials information are missing: SYSE_CREDENTIALS",
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
		creds    map[string]string
		expected string
	}{
		{
			desc:  "success",
			creds: map[string]string{"example.org": "123"},
		},
		{
			desc: "success multiple domains",
			creds: map[string]string{
				"example.org": "123",
				"example.com": "456",
				"example.net": "789",
			},
		},
		{
			desc:     "missing credentials",
			expected: "syse: missing credentials",
		},
		{
			desc:     "missing domain",
			creds:    map[string]string{"": "123"},
			expected: `syse: missing domain: ":123"`,
		},
		{
			desc:     "missing password",
			creds:    map[string]string{"example.org": ""},
			expected: `syse: missing password: "example.org:"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Credentials = test.creds

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
			config.Credentials = map[string]string{
				"example.org": "secret",
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
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("/", servermock.DumpRequest()).
		Route("POST /dns/example.com",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /dns/example.com/1234",
			servermock.Noop()).
		Build(t)

	provider.recordIDs["abc"] = "1234"

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
