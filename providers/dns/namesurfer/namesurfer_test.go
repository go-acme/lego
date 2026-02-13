package namesurfer

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAPIEndpoint, EnvAPIKey, EnvAPISecret)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIEndpoint: "https://namesurfer.example.com:8443/API_10/NSService_10/jsonrpc10",
				EnvAPIKey:      "test_key",
				EnvAPISecret:   "test_secret",
			},
		},
		{
			desc: "missing API endpoint",
			envVars: map[string]string{
				EnvAPIEndpoint: "",
				EnvAPIKey:      "test_key",
				EnvAPISecret:   "test_secret",
			},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_API_ENDPOINT",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIEndpoint: "https://namesurfer.example.com:8443/API_10/NSService_10/jsonrpc10",
				EnvAPIKey:      "",
				EnvAPISecret:   "test_secret",
			},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_API_KEY",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvAPIEndpoint: "https://namesurfer.example.com:8443/API_10/NSService_10/jsonrpc10",
				EnvAPIKey:      "test_key",
				EnvAPISecret:   "",
			},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_API_SECRET",
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
				require.NotNil(t, p.config.HTTPClient)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		apiEndpoint string
		apiKey      string
		apiSecret   string
		expected    string
	}{
		{
			desc:        "success",
			apiEndpoint: "https://namesurfer.example.com:8443/API_10/NSService_10/jsonrpc10",
			apiKey:      "test_key",
			apiSecret:   "test_secret",
		},
		{
			desc:     "missing credentials",
			expected: "namesurfer: incomplete credentials",
		},
		{
			desc:        "missing API endpoint",
			apiEndpoint: "",
			apiKey:      "test_key",
			apiSecret:   "test_secret",
			expected:    "namesurfer: incomplete credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIEndpoint = test.apiEndpoint
			config.APIKey = test.apiKey
			config.APISecret = test.apiSecret

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

func TestNewDNSProvider_ConfigNil(t *testing.T) {
	_, err := NewDNSProviderConfig(nil)
	require.EqualError(t, err, "namesurfer: the configuration of the DNS provider is nil")
}

func TestDNSProvider_Timeout(t *testing.T) {
	config := NewDefaultConfig()
	config.APIEndpoint = "https://namesurfer.example.com:8443/API_10/NSService_10/jsonrpc10"
	config.APIKey = "test_key"
	config.APISecret = "test_secret"
	config.PropagationTimeout = 5 * time.Minute
	config.PollingInterval = 30 * time.Second

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	timeout, interval := provider.Timeout()
	assert.Equal(t, 5*time.Minute, timeout)
	assert.Equal(t, 30*time.Second, interval)
}

func TestCalculateDigest(t *testing.T) {
	config := &Config{
		APIKey:      "testkey",
		APISecret:   "testsecret",
		APIEndpoint: "https://test.example.com",
	}

	provider := &DNSProvider{config: config}

	// Test digest calculation with no parts
	digest1 := provider.calculateDigest()
	assert.NotEmpty(t, digest1)
	assert.Len(t, digest1, 64) // SHA256 produces 64 hex characters

	// Test digest calculation with parts
	digest2 := provider.calculateDigest("zone.example.com", "default")
	assert.NotEmpty(t, digest2)
	assert.Len(t, digest2, 64)
	assert.NotEqual(t, digest1, digest2) // Different inputs should produce different digests
}
