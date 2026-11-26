package shellrentv3

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/shellrentv3/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvClientID, EnvClientSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvClientID:     "user",
				EnvClientSecret: "secret",
			},
		},
		{
			desc: "missing client ID",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "secret",
			},
			expected: "shellrentv3: some credentials information are missing: SHELLRENTV3_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvClientID:     "user",
				EnvClientSecret: "",
			},
			expected: "shellrentv3: some credentials information are missing: SHELLRENTV3_CLIENT_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "shellrentv3: some credentials information are missing: SHELLRENTV3_CLIENT_ID,SHELLRENTV3_CLIENT_SECRET",
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
		clientID string
		secret   string
		expected string
	}{
		{
			desc:     "success",
			clientID: "user",
			secret:   "secret",
		},
		{
			desc:     "missing client ID",
			secret:   "secret",
			expected: "shellrentv3: credentials missing",
		},
		{
			desc:     "missing client secret",
			clientID: "user",
			expected: "shellrentv3: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "shellrentv3: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ClientID = test.clientID
			config.ClientSecret = test.secret

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

	err = provider.Present(t.Context(), envTest.GetDomain(), "abc", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "abc", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.ClientID = "user"
			config.ClientSecret = "secret"
			config.HTTPClient = server.Client()

			oAuthConfig := internal.CreateOAuthConfig(server.URL+"/oauth/token", "user", "secret")

			p := newDNSProviderConfig(config, oAuthConfig)

			p.client.BaseURL, _ = url.Parse(server.URL + "/api/v3")

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /oauth/token",
			servermock.ResponseFromInternal("token.json"),
			servermock.CheckHeader().
				WithBasicAuth("user", "secret"),
		).
		Route("GET /api/v3/purchases",
			servermock.ResponseFromInternal("purchases.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer xxx"),
		).
		Route("GET /api/v3/domains/456",
			servermock.ResponseFromInternal("domain.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer xxx"),
		).
		Route("POST /api/v3/domains/456/dns/records",
			servermock.ResponseFromInternal("record-create.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckHeader().
				WithAuthorization("Bearer xxx"),
			servermock.CheckRequestJSONBodyFromInternal("record-create-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /oauth/token",
			servermock.ResponseFromInternal("token.json"),
		).
		Route("DELETE /api/v3/domains/456/dns/records/abc789",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckHeader().
				WithAuthorization("Bearer xxx"),
		).
		Build(t)

	provider.recordIDs["abc"] = recordKey{domainID: 456, recordID: "abc789"}

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
