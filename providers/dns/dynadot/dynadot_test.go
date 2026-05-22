package dynadot

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
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
			expected: "dynadot: some credentials information are missing: DYNADOT_API_KEY",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvAPIKey:    "key",
				EnvAPISecret: "",
			},
			expected: "dynadot: some credentials information are missing: DYNADOT_API_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "dynadot: some credentials information are missing: DYNADOT_API_KEY,DYNADOT_API_SECRET",
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
			apiSecret: "secret",
			expected:  "dynadot: credentials missing",
		},
		{
			desc:     "missing API secret",
			apiKey:   "key",
			expected: "dynadot: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "dynadot: credentials missing",
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
			With("Authorization", "Bearer key").
			WithRegexp("X-Signature", `.+`),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /restful/v2/domains/example.com/records",
			servermock.ResponseFromInternal("success.json"),
			servermock.CheckRequestJSONBodyFromInternal("set_dns-request.json"),
			servermock.CheckHeader().
				With("X-Signature", "StGY3XMuHaR4iZ1vcddPkasNsVuPyoxdG44w29/iYSM="),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /restful/v2/domains/example.com/records",
			servermock.ResponseFromInternal("success.json"),
			servermock.CheckRequestJSONBodyFromInternal("remove_dns-request.json"),
			servermock.CheckHeader().
				With("X-Signature", "dNpJ/HG586+FnDdgeiNQHGRLl2Sdxav6Q0G3IiGBQT0="),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
