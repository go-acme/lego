package desec

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvToken: "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvToken: "",
			},
			expected: "desec: some credentials information are missing: DESEC_TOKEN",
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
		expected string
		token    string
	}{
		{
			desc:  "success",
			token: "api_key",
		},
		{
			desc:     "missing credentials",
			expected: "desec: incomplete credentials, missing token",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Token = test.token

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Token = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL = server.URL

			return p, nil
		},
		servermock.CheckHeader().WithAuthorization("Token secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/",
			servermock.ResponseFromFixture("domains_responsible.json"),
			servermock.CheckQueryParameter().Strict().
				With("owns_qname", "_acme-challenge.example.com"),
		).
		Route("GET /domains/example.com/rrsets/_acme-challenge/TXT/",
			servermock.ResponseFromFixture("records_get.json"),
		).
		Route("PATCH /domains/example.com/rrsets/_acme-challenge/TXT/",
			servermock.ResponseFromFixture("records_update.json"),
			servermock.CheckRequestJSONBodyFromFixture("records_update-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/",
			servermock.ResponseFromFixture("domains_responsible.json"),
			servermock.CheckQueryParameter().Strict().
				With("owns_qname", "_acme-challenge.example.com"),
		).
		Route("GET /domains/example.com/rrsets/_acme-challenge/TXT/",
			servermock.ResponseFromFixture("records_get.json"),
		).
		Route("PATCH /domains/example.com/rrsets/_acme-challenge/TXT/",
			servermock.ResponseFromFixture("records_update.json"),
			servermock.CheckRequestJSONBodyFromFixture("records_update-request_remove.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
