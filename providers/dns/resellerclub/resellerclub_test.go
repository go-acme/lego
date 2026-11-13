package resellerclub

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/resellerclub/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvUserID).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "secret",
				EnvUserID: "user123",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey: "",
				EnvUserID: "user123",
			},
			expected: "resellerclub: some credentials information are missing: RESELLERCLUB_API_KEY",
		},
		{
			desc: "missing user ID",
			envVars: map[string]string{
				EnvAPIKey: "secret",
				EnvUserID: "",
			},
			expected: "resellerclub: some credentials information are missing: RESELLERCLUB_USER_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "resellerclub: some credentials information are missing: RESELLERCLUB_API_KEY,RESELLERCLUB_USER_ID",
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
		apiKey   string
		userID   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "secret",
			userID: "user123",
		},
		{
			desc:     "missing API key",
			userID:   "user123",
			expected: "resellerclub: credentials missing",
		},
		{
			desc:     "missing user ID",
			apiKey:   "secret",
			expected: "resellerclub: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "resellerclub: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.UserID = test.userID

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
			config.APIKey = "secret"
			config.UserID = "user123"
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
			WithAuthorization("ApiKey secret").
			With(internal.UserIDHeader, "user123"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /dns/zones/example.com/records",
			servermock.ResponseFromInternal("create_dns_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_dns_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /dns/zones/example.com/records",
			servermock.ResponseFromInternal("delete_dns_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("delete_dns_record-request.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
