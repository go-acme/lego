package hostup

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "secret",
			},
		},
		{
			desc:     "missing API key",
			expected: "hostup: some credentials information are missing: HOSTUP_API_KEY",
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
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "secret",
		},
		{
			desc:     "missing API key",
			expected: "hostup: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey

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

func mockProvider() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIKey = "secret"
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
			WithAuthorization("Bearer secret"))
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockProvider().
		Route("GET /dns/zones",
			servermock.ResponseFromInternal("zones.json")).
		Route("POST /dns/zones/9149/records",
			servermock.ResponseFromInternal("record.json").
				WithStatusCode(http.StatusCreated)).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "token", "123d==")
	require.NoError(t, err)

	provider.recordsMu.Lock()
	ref, ok := provider.records["token"]
	provider.recordsMu.Unlock()

	require.True(t, ok)
	require.Equal(t, "9149", ref.ZoneID)
	require.Equal(t, "drr_06ezwatrgahtygnvpz8cp995y0", ref.RecordID)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockProvider().
		Route("DELETE /dns/zones/9149/records/drr_06ezwatrgahtygnvpz8cp995y0",
			servermock.Noop().WithStatusCode(http.StatusNoContent)).
		Build(t)

	provider.recordsMu.Lock()
	provider.records["token"] = recordRef{ZoneID: "9149", RecordID: "drr_06ezwatrgahtygnvpz8cp995y0"}
	provider.recordsMu.Unlock()

	err := provider.CleanUp(t.Context(), "example.com", "token", "123d==")
	require.NoError(t, err)

	provider.recordsMu.Lock()
	_, ok := provider.records["token"]
	provider.recordsMu.Unlock()

	require.False(t, ok)
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
