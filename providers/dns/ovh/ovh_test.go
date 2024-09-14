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
	EnvClientSecret,
	EnvAccessToken).
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
			expected: "ovh: new client: invalid authentication config, both application_key and application_secret must be given",
		},
		{
			desc: "application key: missing application secret",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "",
				EnvConsumerKey:       "D",
			},
			expected: "ovh: new client: invalid authentication config, both application_key and application_secret must be given",
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
			desc: "access token: success",
			envVars: map[string]string{
				EnvEndpoint:    "ovh-eu",
				EnvAccessToken: "G",
			},
		},
		{
			desc: "oauth2: missing client secret",
			envVars: map[string]string{
				EnvEndpoint:     "ovh-eu",
				EnvClientID:     "E",
				EnvClientSecret: "",
			},
			expected: "ovh: new client: invalid oauth2 config, both client_id and client_secret must be given",
		},
		{
			desc: "oauth2: missing client ID",
			envVars: map[string]string{
				EnvEndpoint:     "ovh-eu",
				EnvClientID:     "",
				EnvClientSecret: "F",
			},
			expected: "ovh: new client: invalid oauth2 config, both client_id and client_secret must be given",
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
				EnvAccessToken:       "",
			},
			expected: "ovh: new client: missing authentication information, you need to provide one of the following: application_key/application_secret, client_id/client_secret, or access_token",
		},
		{
			desc: "mixed auth (all)",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
				EnvClientID:          "E",
				EnvClientSecret:      "F",
				EnvAccessToken:       "G",
			},
			expected: "ovh: can't use multiple authentication systems (ApplicationKey, OAuth2, Access Token)",
		},
		{
			desc: "mixed auth (ApplicationKey, OAuth2)",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
				EnvClientID:          "E",
				EnvClientSecret:      "F",
			},
			expected: "ovh: can't use multiple authentication systems (ApplicationKey, OAuth2)",
		},
		{
			desc: "mixed auth (ApplicationKey, Access Token)",
			envVars: map[string]string{
				EnvEndpoint:          "ovh-eu",
				EnvApplicationKey:    "B",
				EnvApplicationSecret: "C",
				EnvConsumerKey:       "D",
				EnvAccessToken:       "G",
			},
			expected: "ovh: can't use multiple authentication systems (ApplicationKey, Access Token)",
		},
		{
			desc: "mixed auth (OAuth2, Access Token)",
			envVars: map[string]string{
				EnvEndpoint:     "ovh-eu",
				EnvClientID:     "E",
				EnvClientSecret: "F",
				EnvAccessToken:  "G",
			},
			expected: "ovh: can't use multiple authentication systems (OAuth2, Access Token)",
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
		accessToken       string
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
			expected:          "ovh: new client: unknown endpoint '', consider checking 'Endpoints' list or using an URL",
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
			expected:          "ovh: new client: invalid authentication config, both application_key and application_secret must be given",
		},
		{
			desc:              "application key: missing application secret",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "",
			consumerKey:       "D",
			expected:          "ovh: new client: invalid authentication config, both application_key and application_secret must be given",
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
			expected:     "ovh: new client: unknown endpoint '', consider checking 'Endpoints' list or using an URL",
		},
		{
			desc:         "oauth2: invalid api endpoint",
			apiEndpoint:  "foobar",
			clientID:     "B",
			clientSecret: "C",
			expected:     "ovh: new client: unknown endpoint 'foobar', consider checking 'Endpoints' list or using an URL",
		},
		{
			desc:         "oauth2: missing client id",
			apiEndpoint:  "ovh-eu",
			clientID:     "",
			clientSecret: "C",
			expected:     "ovh: new client: invalid oauth2 config, both client_id and client_secret must be given",
		},
		{
			desc:         "oauth2: missing client secret",
			apiEndpoint:  "ovh-eu",
			clientID:     "B",
			clientSecret: "",
			expected:     "ovh: new client: invalid oauth2 config, both client_id and client_secret must be given",
		},
		{
			desc:        "access token: success",
			apiEndpoint: "ovh-eu",
			accessToken: "G",
		},
		{
			desc:     "missing credentials",
			expected: "ovh: new client: missing authentication information, you need to provide one of the following: application_key/application_secret, client_id/client_secret, or access_token",
		},
		{
			desc:              "mixed auth (all)",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			clientID:          "B",
			clientSecret:      "C",
			accessToken:       "G",
			expected:          "ovh: can't use multiple authentication systems (ApplicationKey, OAuth2, Access Token)",
		},
		{
			desc:              "mixed auth (ApplicationKey, OAuth2)",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			clientID:          "B",
			clientSecret:      "C",
			expected:          "ovh: can't use multiple authentication systems (ApplicationKey, OAuth2)",
		},
		{
			desc:              "mixed auth (ApplicationKey, Access Token)",
			apiEndpoint:       "ovh-eu",
			applicationKey:    "B",
			applicationSecret: "C",
			consumerKey:       "D",
			accessToken:       "G",
			expected:          "ovh: can't use multiple authentication systems (ApplicationKey, Access Token)",
		},
		{
			desc:         "mixed auth (OAuth2, Access Token)",
			apiEndpoint:  "ovh-eu",
			clientID:     "B",
			clientSecret: "C",
			accessToken:  "G",
			expected:     "ovh: can't use multiple authentication systems (OAuth2, Access Token)",
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
			config.AccessToken = test.accessToken

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
