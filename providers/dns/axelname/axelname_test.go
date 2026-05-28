package axelname

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvNickname, EnvToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvNickname: "user",
				EnvToken:    "secret",
			},
		},
		{
			desc: "missing nickname",
			envVars: map[string]string{
				EnvNickname: "",
				EnvToken:    "secret",
			},
			expected: "axelname: some credentials information are missing: AXELNAME_NICKNAME",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvNickname: "user",
				EnvToken:    "",
			},
			expected: "axelname: some credentials information are missing: AXELNAME_TOKEN",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "axelname: some credentials information are missing: AXELNAME_NICKNAME,AXELNAME_TOKEN",
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
		token    string
		nickname string
		expected string
	}{
		{
			desc:     "success",
			nickname: "user",
			token:    "secret",
		},
		{
			desc:     "missing nickname",
			expected: "axelname: credentials missing",
		},
		{
			desc:     "missing token",
			expected: "axelname: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "axelname: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Token = test.token
			config.Nickname = test.nickname

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
			config.Nickname = "user"
			config.Token = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns_add",
			servermock.ResponseFromInternal("dns_add.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com").
				With("name", "_acme-challenge").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("nichdl", "user").
				With("token", "secret"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns_list",
			servermock.ResponseFromInternal("dns_list.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com").
				With("nichdl", "user").
				With("token", "secret"),
		).
		Route("GET /dns_delete",
			servermock.ResponseFromInternal("dns_delete.json"),
			servermock.CheckQueryParameter().Strict().
				With("id", "74760").
				With("domain", "example.com").
				With("name", "_acme-challenge").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("nichdl", "user").
				With("token", "secret"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
