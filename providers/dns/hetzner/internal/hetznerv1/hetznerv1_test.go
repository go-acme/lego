package hetznerv1

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIToken: "secret",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "hetzner: some credentials information are missing: HETZNER_API_TOKEN",
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
		apiToken string
		expected string
	}{
		{
			desc:     "success",
			apiToken: "secret",
		},
		{
			desc:     "missing credentials",
			expected: "hetzner: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIToken = test.apiToken

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
			config.APIToken = "secret"
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
			WithAuthorization("Bearer secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zones/example.com/rrsets/_acme-challenge/TXT/actions/add_records",
			servermock.ResponseFromFixture("add_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_rrset_records-request.json")).
		Route("GET /actions/1",
			servermock.ResponseFromFixture("get_action_success.json")).
		Build(t)

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err)
}

func TestDNSProvider_Present_error(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zones/example.com/rrsets/_acme-challenge/TXT/actions/add_records",
			servermock.ResponseFromFixture("add_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_rrset_records-request.json")).
		Route("GET /actions/1",
			servermock.ResponseFromFixture("get_action_error.json")).
		Build(t)

	provider.config.PollingInterval = 20 * time.Millisecond
	provider.config.PropagationTimeout = 1 * time.Second

	err := provider.Present("example.com", "", "foobar")
	require.EqualError(t, err, "hetzner: wait (add RRSet records): action 1: error: action_failed: Action failed")
}

func TestDNSProvider_Present_running(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zones/example.com/rrsets/_acme-challenge/TXT/actions/add_records",
			servermock.ResponseFromFixture("add_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_rrset_records-request.json")).
		Route("GET /actions/1",
			servermock.ResponseFromFixture("get_action_running.json")).
		Build(t)

	provider.config.PollingInterval = 20 * time.Millisecond
	provider.config.PropagationTimeout = 1 * time.Second

	err := provider.Present("example.com", "", "foobar")
	require.EqualError(t, err, "hetzner: wait (add RRSet records): action 1 is running")
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zones/example.com/rrsets/_acme-challenge/TXT/actions/remove_records",
			servermock.ResponseFromFixture("remove_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("remove_rrset_records-request.json")).
		Route("GET /actions/1",
			servermock.ResponseFromFixture("get_action_success.json")).
		Build(t)

	err := provider.CleanUp("example.com", "", "foobar")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_error(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zones/example.com/rrsets/_acme-challenge/TXT/actions/remove_records",
			servermock.ResponseFromFixture("remove_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("remove_rrset_records-request.json")).
		Route("GET /actions/1",
			servermock.ResponseFromFixture("get_action_error.json")).
		Build(t)

	provider.config.PollingInterval = 20 * time.Millisecond
	provider.config.PropagationTimeout = 1 * time.Second

	err := provider.CleanUp("example.com", "", "foobar")
	require.EqualError(t, err, "hetzner: wait (remove RRSet records): action 1: error: action_failed: Action failed")
}

func TestDNSProvider_CleanUp_running(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zones/example.com/rrsets/_acme-challenge/TXT/actions/remove_records",
			servermock.ResponseFromFixture("remove_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("remove_rrset_records-request.json")).
		Route("GET /actions/1",
			servermock.ResponseFromFixture("get_action_running.json")).
		Build(t)

	provider.config.PollingInterval = 20 * time.Millisecond
	provider.config.PropagationTimeout = 1 * time.Second

	err := provider.CleanUp("example.com", "", "foobar")
	require.EqualError(t, err, "hetzner: wait (remove RRSet records): action 1 is running")
}
