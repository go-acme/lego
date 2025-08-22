package hostinger

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIToken: "secret",
			},
		},
		{
			desc: "missing API token",
			envVars: map[string]string{
				EnvAPIToken: "",
			},
			expected: "hostinger: some credentials information are missing: HOSTINGER_API_TOKEN",
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
		apiToken string
		expected string
	}{
		{
			desc:     "success",
			apiToken: "secret",
		},
		{
			desc:     "missing API token",
			expected: "hostinger: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIToken = test.apiToken

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

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIToken = "secret"

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.HTTPClient = server.Client()
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
		Route("GET /api/dns/v1/zones/example.com",
			servermock.ResponseFromInternal("get_dns_records.json")).
		Route("PUT /api/dns/v1/zones/example.com",
			servermock.ResponseFromInternal("update_dns_records.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_dns_records-request.json")).
		Build(t)

	err := provider.Present("example.com", "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/dns/v1/zones/example.com",
			servermock.ResponseFromInternal("get_dns_records_acme.json")).
		Route("PUT /api/dns/v1/zones/example.com",
			servermock.ResponseFromInternal("update_dns_records.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_dns_records_base-request.json")).
		Build(t)

	err := provider.CleanUp("example.com", "", "123d==")
	require.NoError(t, err)
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
