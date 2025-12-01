package gravity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/gravity/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvServerURL,
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
				EnvUsername:  "user",
				EnvPassword:  "secret",
				EnvServerURL: "https://example.org:1234",
			},
		},
		{
			desc: "missing EnvUsername",
			envVars: map[string]string{
				EnvUsername:  "",
				EnvPassword:  "secret",
				EnvServerURL: "https://example.org:1234",
			},
			expected: "gravity: some credentials information are missing: GRAVITY_USERNAME",
		},
		{
			desc: "missing EnvPassword",
			envVars: map[string]string{
				EnvUsername:  "user",
				EnvPassword:  "",
				EnvServerURL: "https://example.org:1234",
			},
			expected: "gravity: some credentials information are missing: GRAVITY_PASSWORD",
		},
		{
			desc: "missing EnvServerURL",
			envVars: map[string]string{
				EnvUsername:  "user",
				EnvPassword:  "secret",
				EnvServerURL: "",
			},
			expected: "gravity: some credentials information are missing: GRAVITY_SERVER_URL",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "gravity: some credentials information are missing: GRAVITY_USERNAME,GRAVITY_PASSWORD,GRAVITY_SERVER_URL",
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
		desc      string
		username  string
		password  string
		serverURL string
		expected  string
	}{
		{
			desc:      "success",
			username:  "user",
			password:  "secret",
			serverURL: "https://example.org:1234",
		},
		{
			desc:      "missing username",
			username:  "",
			password:  "secret",
			serverURL: "https://example.org:1234",
			expected:  "gravity: credentials missing",
		},
		{
			desc:      "missing password",
			username:  "user",
			password:  "",
			serverURL: "https://example.org:1234",
			expected:  "gravity: credentials missing",
		},
		{
			desc:      "missing server URL",
			username:  "user",
			password:  "secret",
			serverURL: "",
			expected:  "gravity: server URL missing",
		},
		{
			desc:     "missing credentials",
			expected: "gravity: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.ServerURL = test.serverURL

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
			config.ServerURL = server.URL

			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api/v1/auth/login",
			servermock.ResponseFromInternal("login.json"),
			servermock.CheckRequestJSONBodyFromInternal("login-request.json")).
		Route("GET /api/v1/dns/",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Query().Get("name") != "example.com." {
					servermock.ResponseFromInternal("zones.json").ServeHTTP(rw, req)
					return
				}

				servermock.ResponseFromInternal("zones_empty.json").ServeHTTP(rw, req)
			}),
		).
		Route("POST /api/v1/dns/zones/records",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckQueryParameter().Strict().
				With("zone", "example.com.").
				WithRegexp("uid", `\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`).
				With("hostname", "_acme-challenge")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /api/v1/dns/zones/records",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckQueryParameter().Strict().
				With("zone", "example.com.").
				With("uid", "123").
				With("type", "TXT").
				With("hostname", "_acme-challenge")).
		Build(t)

	provider.records["abc"] = internal.Record{
		Fqdn:     "example.com.",
		Hostname: "_acme-challenge",
		Type:     "TXT",
		UID:      "123",
	}

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
