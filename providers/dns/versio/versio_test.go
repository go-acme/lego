package versio

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDomain = "example.com"

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword, EnvEndpoint).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "me@example.com",
				EnvPassword: "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvPassword: "me@example.com",
			},
			expected: "versio: some credentials information are missing: VERSIO_USERNAME",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				EnvUsername: "TOKEN",
			},
			expected: "versio: some credentials information are missing: VERSIO_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "versio: some credentials information are missing: VERSIO_USERNAME,VERSIO_PASSWORD",
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
		desc     string
		config   *Config
		expected string
	}{
		{
			desc: "success",
			config: &Config{
				Username: "me@example.com",
				Password: "PW",
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "versio: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing username",
			config: &Config{
				Password: "PW",
			},
			expected: "versio: the versio username is missing",
		},
		{
			desc: "missing password",
			config: &Config{
				Username: "UN",
			},
			expected: "versio: the versio password is missing",
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

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "Success",
			builder: mockBuilder().
				Route("GET /domains/example.com",
					servermock.ResponseFromFixture("token.json"),
					servermock.CheckQueryParameter().Strict().
						With("show_dns_records", "true")).
				Route("POST /domains/example.com/update",
					servermock.ResponseFromFixture("token.json")),
		},
		{
			desc: "FailToFindZone",
			builder: mockBuilder().
				Route("GET /domains/example.com",
					servermock.ResponseFromFixture("error_failToFindZone.json").
						WithStatusCode(http.StatusUnauthorized)),
			expectedError: `versio: [status code: 401] 401: ObjectDoesNotExist|Domain not found`,
		},
		{
			desc: "FailToCreateTXT",
			builder: mockBuilder().
				Route("GET /domains/example.com",
					servermock.ResponseFromFixture("token.json"),
					servermock.CheckQueryParameter().Strict().
						With("show_dns_records", "true")).
				Route("POST /domains/example.com/update",
					servermock.ResponseFromFixture("error_failToCreateTXT.json").
						WithStatusCode(http.StatusBadRequest)),
			expectedError: `versio: [status code: 400] 400: ProcessError|DNS record invalid type _acme-challenge.example.eu. TST`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			provider := test.builder.Build(t)

			err := provider.Present(testDomain, "token", "keyAuth")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "Success",
			builder: mockBuilder().
				Route("GET /domains/example.com",
					servermock.ResponseFromFixture("token.json"),
					servermock.CheckQueryParameter().Strict().
						With("show_dns_records", "true")).
				Route("POST /domains/example.com/update",
					servermock.ResponseFromFixture("token.json")),
		},
		{
			desc: "FailToFindZone",
			builder: mockBuilder().
				Route("GET /domains/example.com",
					servermock.ResponseFromFixture("error_failToFindZone.json").
						WithStatusCode(http.StatusUnauthorized)),
			expectedError: `versio: [status code: 401] 401: ObjectDoesNotExist|Domain not found`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			provider := test.builder.Build(t)

			err := provider.CleanUp(testDomain, "token", "keyAuth")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		envTest.Apply(map[string]string{
			EnvUsername: "me@example.com",
			EnvPassword: "secret",
			EnvEndpoint: server.URL,
		})

		return NewDNSProvider()
	})
}
