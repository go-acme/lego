package dnsla

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIID, EnvAPISecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIID:     "1234",
				EnvAPISecret: "secret",
			},
		},
		{
			desc: "missing API ID",
			envVars: map[string]string{
				EnvAPIID:     "",
				EnvAPISecret: "secret",
			},
			expected: "dnsla: some credentials information are missing: DNSLA_API_ID",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvAPIID:     "1234",
				EnvAPISecret: "",
			},
			expected: "dnsla: some credentials information are missing: DNSLA_API_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "dnsla: some credentials information are missing: DNSLA_API_ID,DNSLA_API_SECRET",
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
		apiID     string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiID:     "123",
			apiSecret: "secret",
		},
		{
			desc:      "missing API ID",
			apiSecret: "secret",
			expected:  "dnsla: credentials missing",
		},
		{
			desc:     "missing API secret",
			apiID:    "123",
			expected: "dnsla: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "dnsla: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIID = test.apiID
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
			config.APIID = "123"
			config.APISecret = "secret"

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
			WithBasicAuth("123", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/domainList",
			servermock.ResponseFromInternal("domains.json"),
			servermock.CheckQueryParameter().
				With("pageIndex", "1").
				With("pageSize", "100")).
		Route("POST /api/record",
			servermock.ResponseFromInternal("add_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("add_record-request.json")).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /api/record",
			servermock.ResponseFromInternal("delete_record.json"),
			servermock.CheckQueryParameter().
				With("id", "85371689655342080")).
		Build(t)

	provider.recordIDs["abc"] = "85371689655342080"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
