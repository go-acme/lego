package opusdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/opusdns/opusdns-go-client/models"
	"github.com/opusdns/opusdns-go-client/opusdns"
	"github.com/stretchr/testify/assert"
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
				EnvAPIKey: "opk_test123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "opusdns: some credentials information are missing: OPUSDNS_API_KEY",
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
				APIKey: "opk_test123",
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "opusdns: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing API key",
			config: &Config{
				APIKey: "",
			},
			expected: "opusdns: incomplete credentials, missing API key",
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func setupMockProvider(t *testing.T, handler http.Handler) *DNSProvider {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := opusdns.NewClient(
		opusdns.WithAPIKey("opk_test_secret"),
		opusdns.WithAPIEndpoint(server.URL),
		opusdns.WithMaxRetries(0),
	)
	require.NoError(t, err)

	return &DNSProvider{
		config: &Config{
			APIKey: "opk_test_secret",
			TTL:    60,
		},
		client: client,
		findZoneByFqdn: func(_ context.Context, _ string) (string, error) {
			return "example.com.", nil
		},
	}
}

func TestDNSProvider_Present(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /v1/dns/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-Api-Key"))

		var req models.RecordPatchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		require.Len(t, req.Ops, 1)
		assert.Equal(t, models.RecordOpUpsert, req.Ops[0].Op)
		assert.Equal(t, "_acme-challenge", req.Ops[0].Record.Name)
		assert.Equal(t, models.RRSetTypeTXT, req.Ops[0].Record.Type)
		assert.Equal(t, 60, req.Ops[0].Record.TTL)
		assert.NotEmpty(t, req.Ops[0].Record.RData)

		w.WriteHeader(http.StatusOK)
	})

	provider := setupMockProvider(t, mux)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /v1/dns/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-Api-Key"))

		var req models.RecordPatchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		require.Len(t, req.Ops, 1)
		assert.Equal(t, models.RecordOpRemove, req.Ops[0].Op)
		assert.Equal(t, "_acme-challenge", req.Ops[0].Record.Name)
		assert.Equal(t, models.RRSetTypeTXT, req.Ops[0].Record.Type)
		assert.Equal(t, 60, req.Ops[0].Record.TTL)
		assert.NotEmpty(t, req.Ops[0].Record.RData)

		w.WriteHeader(http.StatusOK)
	})

	provider := setupMockProvider(t, mux)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
