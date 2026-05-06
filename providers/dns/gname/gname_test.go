package gname

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAppID, EnvAppKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAppID:  "app123",
				EnvAppKey: "secret",
			},
		},
		{
			desc: "missing app ID",
			envVars: map[string]string{
				EnvAppID:  "",
				EnvAppKey: "secret",
			},
			expected: "gname: some credentials information are missing: GNAME_APP_ID",
		},
		{
			desc: "missing app key",
			envVars: map[string]string{
				EnvAppID:  "app123",
				EnvAppKey: "",
			},
			expected: "gname: some credentials information are missing: GNAME_APP_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "gname: some credentials information are missing: GNAME_APP_ID,GNAME_APP_KEY",
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
		appID    string
		appKey   string
		expected string
	}{
		{
			desc:   "success",
			appID:  "app123",
			appKey: "secret",
		},
		{
			desc:     "missing app ID",
			appKey:   "secret",
			expected: "gname: credentials missing",
		},
		{
			desc:     "missing app key",
			appID:    "app123",
			expected: "gname: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "gname: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AppID = test.appID
			config.AppKey = test.appKey

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
			config.AppID = "app123"
			config.AppKey = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api/resolution/add",
			servermock.ResponseFromInternal("add_record.json"),
			servermock.CheckForm().Strict().
				With("ym", "example.com").
				With("zj", "_acme-challenge").
				With("lx", "TXT").
				With("ttl", "120").
				With("jlz", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				WithRegexp("gntoken", `[A-Z0-9]{32}`).
				WithRegexp("gntime", `\d+`).
				With("appid", "app123"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api/resolution/delete",
			servermock.ResponseFromInternal("delete_record.json"),
			servermock.CheckForm().Strict().
				With("ym", "example.com").
				With("jxid", "123").
				WithRegexp("gntoken", `[A-Z0-9]{32}`).
				WithRegexp("gntime", `\d+`).
				With("appid", "app123"),
		).
		Build(t)

	provider.recordIDs["abc"] = 123

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
