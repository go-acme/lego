package loopia

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
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

func TestSplitDomain(t *testing.T) {
	provider := &DNSProvider{
		findZoneByFqdn: func(fqdn string) (string, error) {
			return "example.com.", nil
		},
	}

	testCases := []struct {
		desc      string
		fqdn      string
		subdomain string
		domain    string
	}{
		{
			desc:      "single subdomain",
			fqdn:      "subdomain.example.com",
			subdomain: "subdomain",
			domain:    "example.com",
		},
		{
			desc:      "double subdomain",
			fqdn:      "sub.domain.example.com",
			subdomain: "sub.domain",
			domain:    "example.com",
		},
		{
			desc:      "asterisk subdomain",
			fqdn:      "*.example.com",
			subdomain: "*",
			domain:    "example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			subdomain, domain := provider.splitDomain(test.fqdn)

			assert.Equal(t, test.subdomain, subdomain)
			assert.Equal(t, test.domain, domain)
		})
	}
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc          string
		envVars       map[string]string
		expectedError string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUser:     "user",
				EnvAPIPassword: "secret",
			},
		},
		{
			desc: "missing API user",
			envVars: map[string]string{
				EnvAPIUser:     "",
				EnvAPIPassword: "secret",
			},
			expectedError: "loopia: some credentials information are missing: LOOPIA_API_USER",
		},
		{
			desc: "missing API password",
			envVars: map[string]string{
				EnvAPIUser:     "user",
				EnvAPIPassword: "",
			},
			expectedError: "loopia: some credentials information are missing: LOOPIA_API_PASSWORD",
		},
		{
			desc:          "missing credentials",
			envVars:       map[string]string{},
			expectedError: "loopia: some credentials information are missing: LOOPIA_API_USER,LOOPIA_API_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		config        *Config
		expectedTTL   int
		expectedError string
	}{
		{
			desc: "success",
			config: &Config{
				APIUser:     "user",
				APIPassword: "secret",
				TTL:         3600,
			},
			expectedTTL: 3600,
		},
		{
			desc:          "nil config user",
			expectedError: "loopia: the configuration of the DNS provider is nil",
		},
		{
			desc: "empty user",
			config: &Config{
				APIUser:     "",
				APIPassword: "secret",
				TTL:         3600,
			},
			expectedError: "loopia: credentials missing",
		},
		{
			desc: "empty password",
			config: &Config{
				APIUser:     "user",
				APIPassword: "",
				TTL:         3600,
			},
			expectedTTL:   3600,
			expectedError: "loopia: credentials missing",
		},
		{
			desc: "too low TTL",
			config: &Config{
				APIUser:     "user",
				APIPassword: "secret",
				TTL:         299,
			},
			expectedTTL: 300,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

			if test.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.Equal(t, test.expectedTTL, p.config.TTL)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedError)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
