package beget

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
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
				EnvUsername: "123",
				EnvPassword: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "",
			},
			expected: "beget: some credentials information are missing: BEGET_USERNAME,BEGET_PASSWORD",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "456",
			},
			expected: "beget: some credentials information are missing: BEGET_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvPassword: "",
			},
			expected: "beget: some credentials information are missing: BEGET_PASSWORD",
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
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "123",
			password: "456",
		},
		{
			desc:     "missing credentials",
			username: "",
			password: "",
			expected: "beget: incomplete credentials, missing username and/or password",
		},
		{
			desc:     "missing username",
			username: "",
			password: "456",
			expected: "beget: incomplete credentials, missing username and/or password",
		},
		{
			desc:     "missing password",
			username: "123",
			password: "",
			expected: "beget: incomplete credentials, missing username and/or password",
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
	assert.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
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
		servermock.CheckQueryParameter().
			With("login", "user").
			With("passwd", "secret").
			With("input_format", "json").
			With("output_format", "json"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/getData",
			servermock.ResponseFromInternal("getData-real.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"_acme-challenge.example.com"}`),
		).
		Route("GET /dns/changeRecords",
			servermock.ResponseFromInternal("changeRecords-doc.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"_acme-challenge.example.com","records":{"TXT":[{"txtdata":"v=spf1 redirect=beget.com","ttl":300},{"value":"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY","priority":10,"ttl":300}]}}`),
		).
		Build(t)

	err := provider.Present("example.com", "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/getData",
			servermock.ResponseFromInternal("getData.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"_acme-challenge.example.com"}`),
		).
		Route("GET /dns/changeRecords",
			servermock.ResponseFromInternal("changeRecords-doc.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"_acme-challenge.example.com","records":{"TXT":[{"txtdata":"foo","ttl":300}]}}`),
		).
		Build(t)

	err := provider.CleanUp("example.com", "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_empty(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/getData",
			servermock.ResponseFromInternal("getData_empty.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"_acme-challenge.example.com"}`),
		).
		Route("/",
			servermock.Noop().WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	err := provider.CleanUp("example.com", "", "123d==")
	require.NoError(t, err)
}
