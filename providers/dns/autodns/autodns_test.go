package autodns

import (
	"testing"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(envAPIEndpoint, envAPIUser, envAPIPassword)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				envAPIUser:     "123",
				envAPIPassword: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				envAPIUser:     "",
				envAPIPassword: "",
			},
			expected: "autodns: some credentials information are missing: AUTODNS_API_USER,AUTODNS_API_PASSWORD",
		},
		{
			desc: "missing user id",
			envVars: map[string]string{
				envAPIUser:     "",
				envAPIPassword: "456",
			},
			expected: "autodns: some credentials information are missing: AUTODNS_API_USER",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				envAPIUser:     "123",
				envAPIPassword: "",
			},
			expected: "autodns: some credentials information are missing: AUTODNS_API_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "123",
			password: "456",
		},
		{
			desc:     "missing credentials",
			username: "",
			password: "",
			expected: "autodns: missing user",
		},
		{
			desc:     "missing user id",
			username: "",
			password: "456",
			expected: "autodns: missing user",
		},
		{
			desc:     "missing key",
			username: "123",
			password: "",
			expected: "autodns: missing password",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
