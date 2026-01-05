package totaluptime

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

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "secret",
			},
			expected: "totaluptime: some credentials information are missing: TOTALUPTIME_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "",
			},
			expected: "totaluptime: some credentials information are missing: TOTALUPTIME_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "totaluptime: some credentials information are missing: TOTALUPTIME_USERNAME,TOTALUPTIME_PASSWORD",
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
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "user",
			password: "secret",
		},
		{
			desc:     "missing username",
			password: "secret",
			expected: "totaluptime: credentials missing",
		},
		{
			desc:     "missing password",
			username: "user",
			expected: "totaluptime: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "totaluptime: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password

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
			config.Username = "user"
			config.Password = "secret"
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
			WithBasicAuth("user", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /CloudDNS/Domain/ALL",
			servermock.ResponseFromInternal("domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("searchField", "domainName").
				With("searchOper", "eq").
				With("searchString", "example.com"),
		).
		Route("POST /CloudDNS/Domain/d123/TXTRecord",
			servermock.ResponseFromInternal("add_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("add_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"abc": "d123"}, provider.zoneIDs)
	assert.Equal(t, map[string]string{"abc": "r123"}, provider.recordIDs)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /CloudDNS/Domain/d123/TXTRecord/r123",
			servermock.ResponseFromInternal("delete_record.json"),
		).
		Build(t)

	provider.zoneIDs["abc"] = "d123"
	provider.recordIDs["abc"] = "r123"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.zoneIDs)
	assert.Empty(t, provider.recordIDs)
}
