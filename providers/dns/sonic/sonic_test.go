package sonic

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"SONIC_USERID",
	"SONIC_APIKEY").
	WithDomain("SONIC_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"SONIC_USERID": "dummy",
				"SONIC_APIKEY": "dummy",
			},
		},
		{
			desc:     "no userid",
			envVars:  map[string]string{},
			expected: "sonic: some credentials information are missing: SONIC_USERID,SONIC_APIKEY",
		},
		{
			desc: "no apikey",
			envVars: map[string]string{
				"SONIC_USERID": "dummy",
			},
			expected: `sonic: some credentials information are missing: SONIC_APIKEY`,
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
		userID   string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			userID: "dummy",
			apiKey: "dummy",
		},
		{
			desc:     "missing userid",
			expected: "sonic: credentials missing: userID created via https://public-api.sonic.net/dyndns#requesting_an_api_key",
		},
		{
			desc:     "missing apikey",
			userID:   "dummy",
			expected: "sonic: credentials missing: apiKey created via https://public-api.sonic.net/dyndns#requesting_an_api_key",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()

			if test.userID != "" {
				config.UserID = test.userID
			}
			if test.apiKey != "" {
				config.APIKey = test.apiKey
			}

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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}
