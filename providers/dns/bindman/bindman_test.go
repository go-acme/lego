package bindman

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvManagerAddress).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvManagerAddress: "http://localhost",
			},
		},
		{
			desc: "missing bindman manager address",
			envVars: map[string]string{
				EnvManagerAddress: "",
			},
			expected: "bindman: some credentials information are missing: BINDMAN_MANAGER_ADDRESS",
		},
		{
			desc: "empty bindman manager address",
			envVars: map[string]string{
				EnvManagerAddress: "  ",
			},
			expected: "bindman: managerAddress parameter must be a non-empty string",
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
		config   *Config
		expected string
	}{
		{
			desc:   "success",
			config: &Config{BaseURL: "http://localhost"},
		},
		{
			desc:     "missing base URL",
			config:   &Config{BaseURL: ""},
			expected: "bindman: bindman manager address missing",
		},
		{
			desc:     "missing base URL",
			config:   &Config{BaseURL: "  "},
			expected: "bindman: managerAddress parameter must be a non-empty string",
		},
		{
			desc:     "missing config",
			expected: "bindman: the configuration of the DNS provider is nil",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

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

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.BaseURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("User-Agent", "bindman-dns-webhook-client"))
}

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		name        string
		mock        *servermock.Builder[*DNSProvider]
		domain      string
		token       string
		keyAuth     string
		expectError bool
	}{
		{
			name: "success when add record function return no error",
			mock: mockBuilder().
				Route("POST /records",
					servermock.Noop().WithStatusCode(http.StatusNoContent),
					servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
				),
			domain:      "example.com",
			keyAuth:     "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			expectError: false,
		},
		{
			name: "error when add record function return an error",
			mock: mockBuilder().
				Route("POST /records",
					servermock.ResponseFromFixture("error.json"),
				),
			domain:      "example.com",
			keyAuth:     "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			expectError: true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			provider := test.mock.Build(t)

			err := provider.Present(test.domain, test.token, test.keyAuth)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	testCases := []struct {
		name        string
		mock        *servermock.Builder[*DNSProvider]
		domain      string
		token       string
		keyAuth     string
		expectError bool
	}{
		{
			name: "success when remove record function return no error",
			mock: mockBuilder().
				Route("DELETE /records/_acme-challenge.example.com./TXT",
					servermock.Noop().WithStatusCode(http.StatusNoContent),
				),
			domain:      "example.com",
			keyAuth:     "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			expectError: false,
		},
		{
			name: "error when remove record function return an error",
			mock: mockBuilder().
				Route("DELETE /records/_acme-challenge.example.com./TXT",
					servermock.ResponseFromFixture("error.json"),
				),
			domain:      "example.com",
			keyAuth:     "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			provider := test.mock.Build(t)

			err := provider.CleanUp(test.domain, test.token, test.keyAuth)
			if test.expectError {
				require.ErrorContains(t, err, "bindman: ERROR (400): bar; ")
			} else {
				require.NoError(t, err)
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

	err = provider.Present(envTest.GetDomain(), "", "123d==")
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
