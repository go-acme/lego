package webglobe

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvLogin, EnvToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvLogin: "user",
				EnvToken: "secret",
			},
		},
		{
			desc: "missing login",
			envVars: map[string]string{
				EnvLogin: "",
				EnvToken: "secret",
			},
			expected: "webglobe: some credentials information are missing: WEBGLOBE_LOGIN",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvLogin: "user",
				EnvToken: "",
			},
			expected: "webglobe: some credentials information are missing: WEBGLOBE_TOKEN",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "webglobe: some credentials information are missing: WEBGLOBE_LOGIN,WEBGLOBE_TOKEN",
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
		login    string
		token    string
		expected string
	}{
		{
			desc:  "success",
			login: "user",
			token: "secret",
		},
		{
			desc:     "missing login",
			token:    "secret",
			expected: "webglobe: credentials missing",
		},
		{
			desc:     "missing token",
			login:    "user",
			expected: "webglobe: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "webglobe: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Login = test.login
			config.Token = test.token

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
			config.Login = "user"
			config.Token = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)
			p.identifier.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /auth/login",
			servermock.ResponseFromInternal("login.json"),
			servermock.CheckRequestJSONBodyFromInternal("login-request.json"),
		).
		Route("GET /domains/_acme-challenge.example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNotFound),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken"),
		).
		Route("GET /domains/example.com",
			servermock.ResponseFromInternal("domain_details.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken"),
		).
		Route("POST /12345/dns",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Len(t, provider.zoneIDs, 1)
	assert.Len(t, provider.recordIDs, 1)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /auth/login",
			servermock.ResponseFromInternal("login.json"),
			servermock.CheckRequestJSONBodyFromInternal("login-request.json"),
		).
		Route("DELETE /12345/dns/456789",
			servermock.Noop(),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken"),
		).
		Build(t)

	const token = "abc"

	provider.zoneIDs[token] = 12345
	provider.recordIDs[token] = 456789

	err := provider.CleanUp(t.Context(), "example.com", token, "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.zoneIDs)
	assert.Empty(t, provider.recordIDs)
}
