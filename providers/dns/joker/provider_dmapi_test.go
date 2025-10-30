package joker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newDmapiProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success API key",
			envVars: map[string]string{
				EnvAPIKey: "123",
			},
		},
		{
			desc: "success username password",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvPassword: "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvUsername: "",
				EnvPassword: "",
			},
			expected: "joker: some credentials information are missing: JOKER_USERNAME,JOKER_PASSWORD or some credentials information are missing: JOKER_API_KEY",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvUsername: "123",
				EnvPassword: "",
			},
			expected: "joker: some credentials information are missing: JOKER_PASSWORD or some credentials information are missing: JOKER_API_KEY",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvUsername: "",
				EnvPassword: "123",
			},
			expected: "joker: some credentials information are missing: JOKER_USERNAME or some credentials information are missing: JOKER_API_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := newDmapiProvider()

			if test.expected != "" {
				require.EqualError(t, err, test.expected)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
			}
		})
	}
}

func Test_newDmapiProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		username string
		password string
		expected string
	}{
		{
			desc:   "success api key",
			apiKey: "123",
		},
		{
			desc:     "success username and password",
			username: "123",
			password: "123",
		},
		{
			desc:     "missing credentials",
			expected: "joker: credentials missing",
		},
		{
			desc:     "missing credentials: username",
			expected: "joker: credentials missing",
			username: "123",
		},
		{
			desc:     "missing credentials: password",
			expected: "joker: credentials missing",
			password: "123",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.Username = test.username
			config.Password = test.password

			p, err := newDmapiProviderConfig(config)

			if test.expected != "" {
				require.EqualError(t, err, test.expected)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
			}
		})
	}
}
