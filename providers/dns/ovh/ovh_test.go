package ovh

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest                 bool
	envTestAPIEndpoint       string
	envTestApplicationKey    string
	envTestApplicationSecret string
	envTestConsumerKey       string
	envTestDomain            string
)

func init() {
	envTestAPIEndpoint = os.Getenv("OVH_ENDPOINT")
	envTestApplicationKey = os.Getenv("OVH_APPLICATION_KEY")
	envTestApplicationSecret = os.Getenv("OVH_APPLICATION_SECRET")
	envTestConsumerKey = os.Getenv("OVH_CONSUMER_KEY")
	envTestDomain = os.Getenv("OVH_DOMAIN")

	liveTest = len(envTestAPIEndpoint) > 0 && len(envTestApplicationKey) > 0 && len(envTestApplicationSecret) > 0 && len(envTestConsumerKey) > 0
}

func restoreEnv() {
	os.Setenv("OVH_ENDPOINT", envTestAPIEndpoint)
	os.Setenv("OVH_APPLICATION_KEY", envTestApplicationKey)
	os.Setenv("OVH_APPLICATION_SECRET", envTestApplicationSecret)
	os.Setenv("OVH_CONSUMER_KEY", envTestConsumerKey)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "B",
				"OVH_APPLICATION_SECRET": "C",
				"OVH_CONSUMER_KEY":       "D",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "",
				"OVH_APPLICATION_KEY":    "",
				"OVH_APPLICATION_SECRET": "",
				"OVH_CONSUMER_KEY":       "",
			},
			expected: "ovh: some credentials information are missing: OVH_ENDPOINT,OVH_APPLICATION_KEY,OVH_APPLICATION_SECRET,OVH_CONSUMER_KEY",
		},
		{
			desc: "missing endpoint",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "",
				"OVH_APPLICATION_KEY":    "B",
				"OVH_APPLICATION_SECRET": "C",
				"OVH_CONSUMER_KEY":       "D",
			},
			expected: "ovh: some credentials information are missing: OVH_ENDPOINT",
		},
		{
			desc: "missing invalid endpoint",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "foobar",
				"OVH_APPLICATION_KEY":    "B",
				"OVH_APPLICATION_SECRET": "C",
				"OVH_CONSUMER_KEY":       "D",
			},
			expected: "ovh: unknown endpoint 'foobar', consider checking 'Endpoints' list of using an URL",
		},
		{
			desc: "missing application key",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "",
				"OVH_APPLICATION_SECRET": "C",
				"OVH_CONSUMER_KEY":       "D",
			},
			expected: "ovh: some credentials information are missing: OVH_APPLICATION_KEY",
		},
		{
			desc: "missing application secret",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "B",
				"OVH_APPLICATION_SECRET": "",
				"OVH_CONSUMER_KEY":       "D",
			},
			expected: "ovh: some credentials information are missing: OVH_APPLICATION_SECRET",
		},
		{
			desc: "missing consumer key",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "B",
				"OVH_APPLICATION_SECRET": "C",
				"OVH_CONSUMER_KEY":       "",
			},
			expected: "ovh: some credentials information are missing: OVH_CONSUMER_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc              string
		apiEndpoint       string
		applicationKey    string
		applicationSecret string
		consumerKey       string
		expected          string
	}{
		{
			desc:              "success",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
		},
		{
			desc:     "missing credentials",
			expected: "ovh: credentials missing",
		},
		{
			desc:              "missing api endpoint",
			apiEndpoint:       "",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			expected:          "ovh: credentials missing",
		},
		{
			desc:              "missing invalid api endpoint",
			apiEndpoint:       "foobar",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			expected:          "ovh: unknown endpoint 'foobar', consider checking 'Endpoints' list of using an URL",
		},
		{
			desc:              "missing application key",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "",
			applicationSecret: "C",
			consumerKey:       "D",
			expected:          "ovh: credentials missing",
		},
		{
			desc:              "missing application secret",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "",
			consumerKey:       "D",
			expected:          "ovh: credentials missing",
		},
		{
			desc:              "missing consumer key",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "",
			expected:          "ovh: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("OVH_ENDPOINT")
			os.Unsetenv("OVH_APPLICATION_KEY")
			os.Unsetenv("OVH_APPLICATION_SECRET")
			os.Unsetenv("OVH_CONSUMER_KEY")

			config := NewDefaultConfig()
			config.APIEndpoint = test.apiEndpoint
			config.ApplicationKey = test.applicationKey
			config.ApplicationSecret = test.applicationSecret
			config.ConsumerKey = test.consumerKey

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
