package loopia

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIUser,
	EnvAPIPassword,
	EnvTTL,
	EnvPollingInterval,
	EnvPropagationTimeout,
	EnvHTTPTimeout).
	WithDomain(envDomain)

func TestNewDefaultConfig(t *testing.T) {
	testCases := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name: "default",
			expected: Config{
				TTL:                minTTL,
				PropagationTimeout: 40 * time.Minute,
				PollingInterval:    60 * time.Second,
				HTTPClient: &http.Client{
					Timeout: 60 * time.Second,
				},
			},
		},
		{
			name: "overridden values",
			envVars: map[string]string{
				EnvTTL:                "3600",
				EnvPropagationTimeout: "60",
				EnvPollingInterval:    "120",
				EnvHTTPTimeout:        "120",
			},
			expected: Config{
				TTL:                3600,
				PropagationTimeout: time.Minute,
				PollingInterval:    120 * time.Second,
				HTTPClient: &http.Client{
					Timeout: 120 * time.Second,
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)
			p := NewDefaultConfig()
			assert.Equal(t, test.expected.TTL, p.TTL)
			assert.Equal(t, test.expected.PropagationTimeout, p.PropagationTimeout)
			assert.Equal(t, test.expected.PollingInterval, p.PollingInterval)
			assert.Equal(t, test.expected.HTTPClient.Timeout, p.HTTPClient.Timeout)
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		name             string
		config           *Config
		expectedTTL      int
		expectedErrorMsg string
	}{
		{
			name:             "nil config user",
			config:           nil,
			expectedErrorMsg: "loopia: the configuration of the DNS provider is nil",
		},
		{
			name: "empty user",
			config: &Config{
				APIUser:     "",
				APIPassword: "PASSWORD",
				TTL:         3600,
			},
			expectedErrorMsg: "loopia: credentials missing",
		},

		{
			name: "empty password",
			config: &Config{
				APIUser:     "USER",
				APIPassword: "",
				TTL:         3600,
			},
			expectedTTL:      3600,
			expectedErrorMsg: "loopia: credentials missing",
		},
		{
			name: "to low ttl",
			config: &Config{
				APIUser:     "USER",
				APIPassword: "PASSWORD",
				TTL:         299,
			},
			expectedTTL:      300,
			expectedErrorMsg: "",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)
			if len(test.expectedErrorMsg) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.Equal(t, test.expectedTTL, p.config.TTL)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErrorMsg)
			}
		})
	}
}

func TestSplitDomain(t *testing.T) {
	mockedFindZoneByFqdn := func(fqdn string) (string, error) {
		return "example.com.", nil
	}

	provider := &DNSProvider{
		findZoneByFqdn: mockedFindZoneByFqdn,
	}

	testCases := []struct {
		name      string
		fqdn      string
		subdomain string
		domain    string
	}{
		{
			name:      "single subdomain",
			fqdn:      "subdomain.example.com",
			subdomain: "subdomain",
			domain:    "example.com",
		},
		{
			name:      "double subdomain",
			fqdn:      "sub.domain.example.com",
			subdomain: "sub.domain",
			domain:    "example.com",
		},
		{
			name:      "asterisk subdomain",
			fqdn:      "*.example.com",
			subdomain: "*",
			domain:    "example.com",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			subdomain, domain := provider.splitDomain(test.fqdn)
			assert.Equal(t, test.subdomain, subdomain)
			assert.Equal(t, test.domain, domain)
		})
	}
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		name             string
		envVars          map[string]string
		expectedErrorMsg string
	}{
		{
			name: "success",
			envVars: map[string]string{
				EnvAPIUser:     "API_USER",
				EnvAPIPassword: "API_PASSWORD",
			},
		},
		{
			name:             "missing credentials key",
			envVars:          map[string]string{},
			expectedErrorMsg: "loopia: some credentials information are missing: LOOPIA_API_USER,LOOPIA_API_PASSWORD",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()
			if len(test.expectedErrorMsg) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedErrorMsg)
			}
		})
	}
}

func TestDNSProvider_Timeout(t *testing.T) {
	config := &Config{
		PropagationTimeout: 5 * time.Minute,
		PollingInterval:    120 * time.Second,
	}
	provider := &DNSProvider{
		config: config,
	}
	timeout, polling := provider.Timeout()
	assert.Equal(t, timeout, 5*time.Minute)
	assert.Equal(t, polling, 120*time.Second)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
