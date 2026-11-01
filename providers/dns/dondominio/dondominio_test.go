package dondominio

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

var envTest = tester.NewEnvTest(EnvAPIUser, EnvAPIPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUser:     "user123",
				EnvAPIPassword: "secret",
			},
		},
		{
			desc: "missing API user",
			envVars: map[string]string{
				EnvAPIUser:     "",
				EnvAPIPassword: "secret",
			},
			expected: "dondominio: some credentials information are missing: DONDOMINIO_API_USER",
		},
		{
			desc: "missing API password",
			envVars: map[string]string{
				EnvAPIUser:     "user123",
				EnvAPIPassword: "",
			},
			expected: "dondominio: some credentials information are missing: DONDOMINIO_API_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "dondominio: some credentials information are missing: DONDOMINIO_API_USER,DONDOMINIO_API_PASSWORD",
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
		apiUser     string
		apiPassword string
		expected    string
	}{
		{
			desc:        "success",
			apiUser:     "user123",
			apiPassword: "secret",
		},
		{
			desc:        "missing API user",
			apiPassword: "secret",
			expected:    "dondominio: credentials missing",
		},
		{
			desc:     "missing API user",
			apiUser:  "user123",
			expected: "dondominio: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "dondominio: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUser = test.apiUser
			config.APIPassword = test.apiPassword

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
			config.APIUser = "user123"
			config.APIPassword = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithAccept("application/json").
			WithContentTypeFromURLEncoded(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /service/dnscreate",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckForm().Strict().
				With("name", "_acme-challenge.example.com").
				With("serviceName", "example.com").
				With("ttl", "300").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("apipasswd", "secret").
				With("apiuser", "user123"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"abc": "SRV0000214277-000396072"}, provider.recordIDs)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /service/dnsdelete",
			servermock.ResponseFromInternal("delete_record.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("entityID", "SRV0000214277-000396072").
				With("apipasswd", "secret").
				With("apiuser", "user123"),
		).
		Build(t)

	provider.recordIDs["abc"] = "SRV0000214277-000396072"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.recordIDs)
}
