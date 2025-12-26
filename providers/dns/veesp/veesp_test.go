package veesp

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/veesp/internal"
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
			expected: "veesp: some credentials information are missing: VEESP_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "",
			},
			expected: "veesp: some credentials information are missing: VEESP_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "veesp: some credentials information are missing: VEESP_USERNAME,VEESP_PASSWORD",
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
			expected: "veesp: credentials missing",
		},
		{
			desc:     "missing password",
			username: "user",
			expected: "veesp: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "veesp: credentials missing",
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
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns",
			servermock.ResponseFromInternal("zones.json"),
		).
		Route("POST /service/10/dns/60/records",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromInternal("add_record-request.json"),
		).
		Route("/",
			servermock.DumpRequest()).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /service/10/dns/60",
			servermock.ResponseFromInternal("records.json"),
		).
		Route("DELETE /service/10/dns/60/records/66",
			servermock.Noop(),
		).
		Build(t)

	provider.Zones["abc"] = &internal.Zone{
		DomainID:  "60",
		Name:      "example.com",
		ServiceID: "10",
	}

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
