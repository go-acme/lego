package tele3

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvKey, EnvSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvKey:    "keyA",
				EnvSecret: "secretA",
			},
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				EnvKey:    "",
				EnvSecret: "secretA",
			},
			expected: "tele3: some credentials information are missing: TELE3_KEY",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvKey:    "keyA",
				EnvSecret: "",
			},
			expected: "tele3: some credentials information are missing: TELE3_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "tele3: some credentials information are missing: TELE3_KEY,TELE3_SECRET",
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
		key      string
		secret   string
		expected string
	}{
		{
			desc:   "success",
			key:    "keyA",
			secret: "secretA",
		},
		{
			desc:     "missing key",
			secret:   "secretA",
			expected: "tele3: credentials missing",
		},
		{
			desc:     "missing secret",
			key:      "keyA",
			expected: "tele3: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "tele3: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Key = test.key
			config.Secret = test.secret

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
			config.Key = "keyA"
			config.Secret = "secretA"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL = server.URL

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /",
			servermock.RawStringResponse("success"),
			servermock.CheckRequestJSONBodyFromInternal("add_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /",
			servermock.RawStringResponse("success"),
			servermock.CheckRequestJSONBodyFromInternal("remove_record-request.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
