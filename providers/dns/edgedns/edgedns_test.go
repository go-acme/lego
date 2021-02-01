package edgedns

import (
	"os"
	"testing"
	"time"

	configdns "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "TEST_DOMAIN"

var envTest = tester.NewEnvTest(
	EnvHost,
	EnvClientToken,
	EnvClientSecret,
	EnvAccessToken,
	EnvEdgeRc,
	EnvEdgeRcSection,
	"AKAMAI_TEST_HOST",
	"AKAMAI_TEST_CLIENT_TOKEN",
	"AKAMAI_TEST_CLIENT_SECRET",
	"AKAMAI_TEST_ACCESS_TOKEN").
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvHost, EnvClientToken, EnvClientSecret, EnvAccessToken, envDomain)

func TestNewDNSProvider_FromEnv(t *testing.T) {
	testCases := []struct {
		desc           string
		envVars        map[string]string
		expectedConfig *edgegrid.Config
		expectedErr    string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvHost:         "akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx.luna.akamaiapis.net",
				EnvClientToken:  "akab-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				EnvClientSecret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				EnvAccessToken:  "akac-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
			},
			expectedConfig: &edgegrid.Config{
				Host:         "akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx.luna.akamaiapis.net",
				ClientToken:  "akab-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				ClientSecret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				AccessToken:  "akac-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				MaxBody:      EdgeGridMaxBody,
			},
		},
		{
			desc: "with section",
			envVars: map[string]string{
				EnvEdgeRcSection:            "test",
				"AKAMAI_TEST_HOST":          "akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx.luna.akamaiapis.net",
				"AKAMAI_TEST_CLIENT_TOKEN":  "akab-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				"AKAMAI_TEST_CLIENT_SECRET": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"AKAMAI_TEST_ACCESS_TOKEN":  "akac-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
			},
			expectedConfig: &edgegrid.Config{
				Host:         "akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx.luna.akamaiapis.net",
				ClientToken:  "akab-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				ClientSecret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				AccessToken:  "akac-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				MaxBody:      EdgeGridMaxBody,
			},
		},
		{
			desc:        "missing credentials",
			expectedErr: "edgedns: Unable to create instance using environment or .edgerc file",
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				EnvHost:         "",
				EnvClientToken:  "B",
				EnvClientSecret: "C",
				EnvAccessToken:  "D",
			},
			expectedErr: "edgedns: Unable to create instance using environment or .edgerc file",
		},
		{
			desc: "missing client token",
			envVars: map[string]string{
				EnvHost:         "A",
				EnvClientToken:  "",
				EnvClientSecret: "C",
				EnvAccessToken:  "D",
			},
			expectedErr: "edgedns: Fatal missing required environment variables: [AKAMAI_CLIENT_TOKEN]",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvHost:         "A",
				EnvClientToken:  "B",
				EnvClientSecret: "",
				EnvAccessToken:  "D",
			},
			expectedErr: "edgedns: Fatal missing required environment variables: [AKAMAI_CLIENT_SECRET]",
		},
		{
			desc: "missing access token",
			envVars: map[string]string{
				EnvHost:         "A",
				EnvClientToken:  "B",
				EnvClientSecret: "C",
				EnvAccessToken:  "",
			},
			expectedErr: "edgedns: Fatal missing required environment variables: [AKAMAI_ACCESS_TOKEN]",
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)
			os.Setenv(EnvEdgeRc, "/dev/null")

			p, err := NewDNSProvider()

			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			require.NotNil(t, p.config)

			if test.expectedConfig != nil {
				require.Equal(t, *test.expectedConfig, configdns.Config)
			}
		})
	}
}

func TestDNSProvider_findZone(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected string
	}{
		{
			desc:     "Extract root record name",
			domain:   "bar.com",
			expected: "bar.com",
		},
		{
			desc:     "Extract sub record name",
			domain:   "foo.bar.com",
			expected: "bar.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, err := findZone(test.domain)
			require.NoError(t, err)
			require.Equal(t, test.expected, zone)
		})
	}
}

func TestNewDefaultConfig(t *testing.T) {
	defer envTest.RestoreEnv()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected *Config
	}{
		{
			desc: "default configuration",
			expected: &Config{
				TTL:                dns01.DefaultTTL,
				PropagationTimeout: 3 * time.Minute,
				PollingInterval:    15 * time.Second,
			},
		},
		{
			desc: "",
			envVars: map[string]string{
				EnvTTL:                "99",
				EnvPropagationTimeout: "60",
				EnvPollingInterval:    "60",
			},
			expected: &Config{
				TTL:                99,
				PropagationTimeout: 60 * time.Second,
				PollingInterval:    60 * time.Second,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			envTest.ClearEnv()
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}

			config := NewDefaultConfig()

			require.Equal(t, test.expected, config)
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

	// Present Twice to handle create / update
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
