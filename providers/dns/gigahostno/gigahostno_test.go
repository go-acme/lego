package gigahostno

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/gigahostno/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvSecret,
).WithDomain(envDomain)

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
				EnvSecret:   "super-secret",
			},
		},
		{
			desc: "missing GIGAHOSTNO_USERNAME",
			envVars: map[string]string{
				EnvPassword: "secret",
			},
			expected: "gigahostno: some credentials information are missing: GIGAHOSTNO_USERNAME",
		},
		{
			desc: "missing GIGAHOSTNO_PASSWORD",
			envVars: map[string]string{
				EnvUsername: "user",
			},
			expected: "gigahostno: some credentials information are missing: GIGAHOSTNO_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "gigahostno: some credentials information are missing: GIGAHOSTNO_USERNAME,GIGAHOSTNO_PASSWORD",
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
		secret   string
		expected string
	}{
		{
			desc:     "success",
			username: "user",
			password: "secret",
			secret:   "super-secret",
		},
		{
			desc:     "missing username",
			password: "secret",
			expected: "gigahostno: credentials missing",
		},
		{
			desc:     "missing password",
			username: "user",
			expected: "gigahostno: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "gigahostno: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
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
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Username = "user"
			config.Password = "secret"
			config.Secret = "JBSWY3DPEHPK3PXP"

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
		Route("POST /authenticate",
			servermock.ResponseFromInternal("authenticate.json")).
		Route("GET /dns/zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken")).
		Route("POST /dns/zones/123/records",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_token_not_expired(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret-token")).
		Route("POST /dns/zones/123/records",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret-token")).
		Build(t)

	provider.token = &internal.Token{
		Token:       "secret-token",
		TokenExpire: 65322892800, // 2040-01-01
		CustomerID:  "123",
	}

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /authenticate",
			servermock.ResponseFromInternal("authenticate.json")).
		Route("GET /dns/zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken")).
		Route("GET /dns/zones/123/records",
			servermock.ResponseFromInternal("zone_records.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken")).
		Route("DELETE /dns/zones/123/records/jkl012",
			servermock.ResponseFromInternal("delete_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secrettoken")).
		Build(t)

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_token_not_expired(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret-token")).
		Route("GET /dns/zones/123/records",
			servermock.ResponseFromInternal("zone_records.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret-token")).
		Route("DELETE /dns/zones/123/records/jkl012",
			servermock.ResponseFromInternal("delete_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret-token")).
		Build(t)

	provider.token = &internal.Token{
		Token:       "secret-token",
		TokenExpire: 65322892800, // 2040-01-01
		CustomerID:  "123",
	}

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
