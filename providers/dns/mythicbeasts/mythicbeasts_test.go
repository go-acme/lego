package mythicbeasts

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUserName,
	EnvPassword).
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
				EnvUserName: "123",
				EnvPassword: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvUserName: "",
				EnvPassword: "",
			},
			expected: "mythicbeasts: some credentials information are missing: MYTHICBEASTS_CLIENT_ID,MYTHICBEASTS_CLIENT_SECRET",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvUserName: "",
				EnvPassword: "api_password",
			},
			expected: "mythicbeasts: some credentials information are missing: MYTHICBEASTS_CLIENT_ID",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvUserName: "api_username",
				EnvPassword: "",
			},
			expected: "mythicbeasts: some credentials information are missing: MYTHICBEASTS_CLIENT_SECRET",
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
		desc         string
		clientID     string
		clientSecret string
		expected     string
	}{
		{
			desc:         "success",
			clientID:     "api_client_id",
			clientSecret: "api_client_secret",
		},
		{
			desc:     "missing credentials",
			expected: "mythicbeasts: incomplete credentials, missing client ID and/or client secret",
		},
		{
			desc:         "missing username",
			clientID:     "",
			clientSecret: "api_client_secret",
			expected:     "mythicbeasts: incomplete credentials, missing client ID and/or client secret",
		},
		{
			desc:         "missing password",
			clientID:     "api_client_id",
			clientSecret: "",
			expected:     "mythicbeasts: incomplete credentials, missing client ID and/or client secret",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config, err := NewDefaultConfig()
			require.NoError(t, err)

			config.ClientID = test.clientID
			config.ClientSecret = test.clientSecret

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func oauthServer(t *testing.T) string {
	t.Helper()

	return servermock.NewBuilder[string](
		func(server *httptest.Server) (string, error) {
			return server.URL, nil
		}).
		Route("POST /",
			servermock.ResponseFromInternal("token.json"),
			servermock.CheckHeader().
				WithBasicAuth("user", "secret").
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("grant_type", "client_credentials"),
		).
		Build(t)
}

func mockBuilder(tokenURL string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config, err := NewDefaultConfig()
			if err != nil {
				return nil, err
			}

			config.ClientID = "user"
			config.ClientSecret = "secret"

			config.APIEndpoint, _ = url.Parse(server.URL)
			config.AuthAPIEndpoint, _ = url.Parse(tokenURL)

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer xxx"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("POST /zones/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromInternal("post_zoneszonerecords.json"),
			servermock.CheckRequestJSONBodyFromInternal("post_zoneszonerecords-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("DELETE /zones/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromInternal("delete_zoneszonerecords.json"),
			servermock.CheckQueryParameter().Strict().
				With("data", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
