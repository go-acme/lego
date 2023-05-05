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

const (
	envDomain           = envNamespace + "TEST_DOMAIN"
	envTestHost         = envNamespace + "TEST_HOST"
	envTestClientToken  = envNamespace + "TEST_CLIENT_TOKEN"
	envTestClientSecret = envNamespace + "TEST_CLIENT_SECRET"
	envTestAccessToken  = envNamespace + "TEST_ACCESS_TOKEN"
)

var envTest = tester.NewEnvTest(
	EnvHost,
	EnvClientToken,
	EnvClientSecret,
	EnvAccessToken,
	EnvEdgeRc,
	EnvEdgeRcSection,
	envTestHost,
	envTestClientToken,
	envTestClientSecret,
	envTestAccessToken).
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
				MaxBody:      maxBody,
			},
		},
		{
			desc: "with section",
			envVars: map[string]string{
				EnvEdgeRcSection:    "test",
				envTestHost:         "akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx.luna.akamaiapis.net",
				envTestClientToken:  "akab-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				envTestClientSecret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				envTestAccessToken:  "akac-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
			},
			expectedConfig: &edgegrid.Config{
				Host:         "akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx.luna.akamaiapis.net",
				ClientToken:  "akab-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				ClientSecret: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				AccessToken:  "akac-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx",
				MaxBody:      maxBody,
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

			if test.envVars == nil {
				test.envVars = map[string]string{}
			}
			test.envVars[EnvEdgeRc] = "/dev/null"

			envTest.Apply(test.envVars)

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
			domain:   "bar.com.",
			expected: "bar.com",
		},
		{
			desc:     "Extract sub record name",
			domain:   "foo.bar.com.",
			expected: "bar.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, err := getZone(test.domain)
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
				Config: edgegrid.Config{
					MaxBody: maxBody,
				},
			},
		},
		{
			desc: "custom values",
			envVars: map[string]string{
				EnvTTL:                "99",
				EnvPropagationTimeout: "60",
				EnvPollingInterval:    "60",
			},
			expected: &Config{
				TTL:                99,
				PropagationTimeout: 60 * time.Second,
				PollingInterval:    60 * time.Second,
				Config: edgegrid.Config{
					MaxBody: maxBody,
				},
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
