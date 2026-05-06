package xinnet

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvSecret, EnvAgentID).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvSecret:  "secret",
				EnvAgentID: "agent12345",
			},
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvSecret:  "",
				EnvAgentID: "agent12345",
			},
			expected: "xinnet: some credentials information are missing: XINNET_SECRET",
		},
		{
			desc: "missing agent ID",
			envVars: map[string]string{
				EnvSecret:  "secret",
				EnvAgentID: "",
			},
			expected: "xinnet: some credentials information are missing: XINNET_AGENT_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "xinnet: some credentials information are missing: XINNET_SECRET,XINNET_AGENT_ID",
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
		secret   string
		agentID  string
		expected string
	}{
		{
			desc:    "success",
			secret:  "secret",
			agentID: "agent12345",
		},
		{
			desc:     "missing secret",
			agentID:  "agent12345",
			expected: "xinnet: credentials missing",
		},
		{
			desc:     "missing agent ID",
			secret:   "secret",
			expected: "xinnet: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "xinnet: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Secret = test.secret
			config.AgentID = test.agentID

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
			config.Secret = "secret"
			config.AgentID = "agent12345"
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
		Route("POST /api/dns/create/",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
			servermock.CheckHeader().
				WithRegexp("Timestamp", `\d{8}T\d{6}Z`).
				WithRegexp("Authorization", `HMAC-SHA256 Access=agent12345, Signature=\w+`),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /api/dns/delete/",
			servermock.ResponseFromInternal("delete_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("delete_record-request.json"),
			servermock.CheckHeader().
				WithRegexp("Timestamp", `\d{8}T\d{6}Z`).
				WithRegexp("Authorization", `HMAC-SHA256 Access=agent12345, Signature=\w+`),
		).
		Build(t)

	provider.recordIDs["abc"] = 165812154

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
