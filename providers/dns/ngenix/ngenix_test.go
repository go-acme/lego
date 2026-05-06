package ngenix

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvToken,
	EnvCustomerID,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername:   "email@example.com",
				EnvToken:      "secret",
				EnvCustomerID: "42",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername:   "",
				EnvToken:      "secret",
				EnvCustomerID: "42",
			},
			expected: "ngenix: some credentials information are missing: NGENIX_USERNAME",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvUsername:   "email@example.com",
				EnvToken:      "",
				EnvCustomerID: "42",
			},
			expected: "ngenix: some credentials information are missing: NGENIX_TOKEN",
		},
		{
			desc: "missing customer ID",
			envVars: map[string]string{
				EnvUsername:   "email@example.com",
				EnvToken:      "secret",
				EnvCustomerID: "",
			},
			expected: "ngenix: some credentials information are missing: NGENIX_CUSTOMER_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "ngenix: some credentials information are missing: NGENIX_USERNAME,NGENIX_TOKEN,NGENIX_CUSTOMER_ID",
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
		username   string
		token      string
		customerID string
		expected   string
	}{
		{
			desc:       "success",
			username:   "email@example.com",
			token:      "secret",
			customerID: "42",
		},
		{
			desc:       "missing username",
			token:      "secret",
			customerID: "42",
			expected:   "ngenix: credentials missing: username",
		},
		{
			desc:       "missing token",
			username:   "email@example.com",
			customerID: "42",
			expected:   "ngenix: credentials missing: token",
		},
		{
			desc:     "missing customer ID",
			username: "email@example.com",
			token:    "secret",
			expected: "ngenix: credentials missing: customerID",
		},
		{
			desc:     "missing credentials",
			expected: "ngenix: credentials missing: username",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Token = test.token
			config.CustomerID = test.customerID

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
			config.Username = "email@example.com"
			config.Token = "secret"
			config.CustomerID = "42"
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
		Route("GET /dns-zone",
			servermock.ResponseFromInternal("list_dns_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("customerId", "42"),
		).
		Route("GET /dns-zone/1",
			servermock.ResponseFromInternal("get_dns_zone.json"),
		).
		Route("PATCH /dns-zone/1",
			servermock.ResponseFromInternal("update_dns_zone.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_dns_zone-request_add.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns-zone",
			servermock.ResponseFromInternal("list_dns_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("customerId", "42"),
		).
		Route("GET /dns-zone/1",
			servermock.ResponseFromInternal("get_dns_zone.json"),
		).
		Route("PATCH /dns-zone/1",
			servermock.ResponseFromInternal("update_dns_zone.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_dns_zone-request_remove.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
