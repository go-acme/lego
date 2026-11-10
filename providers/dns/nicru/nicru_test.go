package nicru

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/nicru/internal"
	"github.com/stretchr/testify/require"
)

const (
	fakeServiceID = "2519234972459cdfa23423adf143324f"
	fakeSecret    = "oo5ahrie0aiPho3Vee4siupoPhahdahCh1thiesohru"
	fakeUsername  = "1234567/NIC-D"
	fakePassword  = "einge8Goo2eBaiXievuj"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword, EnvServiceID, EnvSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvSecret:    fakeSecret,
				EnvUsername:  fakeUsername,
				EnvPassword:  fakePassword,
			},
		},
		{
			desc: "missing serviceID",
			envVars: map[string]string{
				EnvSecret:   fakeSecret,
				EnvUsername: fakeUsername,
				EnvPassword: fakePassword,
			},
			expected: "nicru: some credentials information are missing: NICRU_SERVICE_ID",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvUsername:  fakeUsername,
				EnvPassword:  fakePassword,
			},
			expected: "nicru: some credentials information are missing: NICRU_SECRET",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvSecret:    fakeSecret,
				EnvPassword:  fakePassword,
			},
			expected: "nicru: some credentials information are missing: NICRU_USER",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvSecret:    fakeSecret,
				EnvUsername:  fakeUsername,
			},
			expected: "nicru: some credentials information are missing: NICRU_PASSWORD",
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
				ServiceID: fakeServiceID,
				Secret:    fakeSecret,
				Username:  fakeUsername,
				Password:  fakePassword,
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "nicru: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing username",
			config: &Config{
				ServiceID: fakeServiceID,
				Password:  fakePassword,
			},
			expected: "nicru: username is missing in credentials information",
		},
		{
			desc: "missing password",
			config: &Config{
				ServiceID: fakeServiceID,
				Secret:    fakeSecret,
				Username:  fakeUsername,
			},
			expected: "nicru: password is missing in credentials information",
		},
		{
			desc: "missing secret",
			config: &Config{
				ServiceID: fakeServiceID,
				Username:  fakeUsername,
				Password:  fakePassword,
			},
			expected: "nicru: secret is missing in credentials information",
		},
		{
			desc: "missing serviceID",
			config: &Config{
				Secret:   fakeSecret,
				Username: fakeUsername,
				Password: fakePassword,
			},
			expected: "nicru: serviceID is missing in credentials information",
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

func oauthServer(t *testing.T) string {
	t.Helper()

	return servermock.NewBuilder[string](
		func(server *httptest.Server) (string, error) {
			return server.URL, nil
		}).
		Route("POST /",
			// NOTE(ldez): I don't know if the real response is JSON or x-www-form-urlencoded.
			servermock.ResponseFromInternal("auth.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("client_id", "client123").
				With("client_secret", "secret123").
				With("grant_type", "password").
				With("password", "pwd123").
				With("scope", ".+:/dns-master/.+").
				With("username", "user123"),
		).
		Build(t)
}

func mockBuilder(tokenURL string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.ServiceID = "client123"
			config.Secret = "secret123"
			config.Username = "user123"
			config.Password = "pwd123"

			authConfig := internal.CreateOAuthConfig(tokenURL, config.ServiceID, config.Secret)

			p, err := newDNSProviderConfig(config, authConfig)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithAuthorization("Bearer at123").
			WithAccept("text/xml"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("GET /zones",
			servermock.ResponseFromInternal("zones_all_GET.xml"),
		).
		Route("GET /services/myservice/zones/example.com/records",
			servermock.ResponseFromInternal("records_GET.xml"),
		).
		Route("PUT /services/myservice/zones/example.com/records",
			servermock.ResponseFromInternal("records_PUT.xml"),
			servermock.CheckHeader().
				WithContentType("text/xml"),
			servermock.CheckRequestBodyFromInternal("records_PUT_real-request.xml").
				IgnoreWhitespace(),
		).
		Route("POST /services/myservice/zones/example.com/commit",
			servermock.ResponseFromInternal("commit_POST.xml"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("GET /zones",
			servermock.ResponseFromInternal("zones_all_GET.xml"),
		).
		Route("GET /services/myservice/zones/example.com/records",
			servermock.ResponseFromInternal("records_GET_real.xml"),
		).
		Route("DELETE /services/myservice/zones/example.com/records/123456",
			servermock.ResponseFromInternal("record_DELETE.xml"),
		).
		Route("POST /services/myservice/zones/example.com/commit",
			servermock.ResponseFromInternal("commit_POST.xml"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
