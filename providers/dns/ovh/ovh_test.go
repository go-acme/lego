package ovh

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEndpoint,
	EnvApplicationKey,
	EnvApplicationSecret,
	EnvConsumerKey,
	EnvClientID,
	EnvClientSecret).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "application key: success",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
			},
		},
		{
			desc: "application key: missing endpoint",
			envVars: map[string]string{
				EnvEndpoint:          "",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
			},
			expected: "ovh: some credentials information are missing: OVH_ENDPOINT",
		},
		{
			desc: "application key: missing invalid endpoint",
			envVars: map[string]string{
				EnvEndpoint:          "foobar",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
			},
			expected: "ovh: new client: unknown endpoint 'foobar', consider checking 'Endpoints' list or using an URL",
		},
		{
			desc: "application key: missing application key",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
			},
			expected: "ovh: some credentials information are missing: OVH_APPLICATION_KEY",
		},
		{
			desc: "application key: missing application secret",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "",
				EnvConsumerKey:       "D",
			},
			expected: "ovh: some credentials information are missing: OVH_APPLICATION_SECRET",
		},
		{
			desc: "application key: missing consumer key",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "",
			},
			expected: "ovh: some credentials information are missing: OVH_CONSUMER_KEY",
		},
		{
			desc: "oauth2: success",
			envVars: map[string]string{
				EnvEndpoint:     "ovh-eu",
				EnvClientID:     "E",
				EnvClientSecret: "F",
			},
		},
		{
			desc: "oauth2: missing client secret",
			envVars: map[string]string{
				EnvEndpoint:     "ovh-eu",
				EnvClientID:     "E",
				EnvClientSecret: "",
			},
			expected: "ovh: some credentials information are missing: OVH_CLIENT_SECRET",
		},
		{
			desc: "oauth2: missing client ID",
			envVars: map[string]string{
				EnvEndpoint:     "ovh-eu",
				EnvClientID:     "",
				EnvClientSecret: "F",
			},
			expected: "ovh: some credentials information are missing: OVH_CLIENT_ID",
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvEndpoint:          "",
				EnvApplicationKey:    "",
				EnvApplicationSecret: "",
				EnvConsumerKey:       "",
				EnvClientID:          "",
				EnvClientSecret:      "",
			},
			expected: "ovh: some credentials information are missing: OVH_ENDPOINT,OVH_APPLICATION_KEY,OVH_APPLICATION_SECRET,OVH_CONSUMER_KEY",
		},
		{
			desc: "mixed auth",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
				EnvClientID:          "E",
				EnvClientSecret:      "F",
			},
			expected: "ovh: can't use both OVH_APPLICATION_KEY and OVH_CLIENT_ID at the same time",
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
		clientID          string
		clientSecret      string
		expected          string
	}{
		{
			desc:              "application key: success",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
		},
		{
			desc:              "application key: missing api endpoint",
			apiEndpoint:       "",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			expected:          "ovh: credentials are missing",
		},
		{
			desc:              "application key: invalid api endpoint",
			apiEndpoint:       "foobar",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			expected:          "ovh: new client: unknown endpoint 'foobar', consider checking 'Endpoints' list or using an URL",
		},
		{
			desc:              "application key: missing application key",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "",
			applicationSecret: "C",
			consumerKey:       "D",
			expected:          "ovh: credentials are missing",
		},
		{
			desc:              "application key: missing application secret",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "",
			consumerKey:       "D",
			expected:          "ovh: credentials are missing",
		},
		{
			desc:              "application key: missing consumer key",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "",
			expected:          "ovh: credentials are missing",
		},
		{
			desc:         "oauth2: success",
			apiEndpoint:  "ovh-eu",
			clientID:     "B",
			clientSecret: "C",
		},
		{
			desc:         "oauth2: missing api endpoint",
			apiEndpoint:  "",
			clientID:     "B",
			clientSecret: "C",
			expected:     "ovh: credentials are missing",
		},
		{
			desc:         "oauth2: invalid api endpoint",
			apiEndpoint:  "foobar",
			clientID:     "B",
			clientSecret: "C",
			expected:     "ovh: new OAuth2 client: unknown endpoint 'foobar', consider checking 'Endpoints' list or using an URL",
		},
		{
			desc:         "oauth2: missing client id",
			apiEndpoint:  "ovh-eu",
			clientID:     "",
			clientSecret: "C",
			expected:     "ovh: credentials are missing",
		},
		{
			desc:         "oauth2: missing client secret",
			apiEndpoint:  "ovh-eu",
			clientID:     "B",
			clientSecret: "",
			expected:     "ovh: credentials are missing",
		},
		{
			desc:     "missing credentials",
			expected: "ovh: credentials are missing",
		},
		{
			desc:              "mixed auth",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			clientID:          "B",
			clientSecret:      "C",
			expected:          "ovh: can't use both authentication systems (ApplicationKey and OAuth2)",
		},
	}

	// The OVH client use the same env vars than lego, so it requires to clean them.
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIEndpoint = test.apiEndpoint
			config.ApplicationKey = test.applicationKey
			config.ApplicationSecret = test.applicationSecret
			config.ConsumerKey = test.consumerKey

			if test.clientID != "" || test.clientSecret != "" {
				config.OAuth2Config = &OAuth2Config{
					ClientID:     test.clientID,
					ClientSecret: test.clientSecret,
				}
			}

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
