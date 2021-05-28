package infoblox

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	EnvUser,
	EnvHost,
	EnvPassword,
).
	WithDomain("INFOBLOX_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success email, password, host",
			envVars: map[string]string{
				EnvUser:     "test@example.com",
				EnvPassword: "123",
				EnvHost:     "infoblox.yourorg.org",
			},
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				EnvUser:     "",
				EnvPassword: "",
			},
			expected: "infoblox new dns provider could not get config from env: infoblox build config from env could not find value for " + EnvHost,
		},
		{
			desc: "missing user",
			envVars: map[string]string{
				EnvUser:     "",
				EnvPassword: "",
				EnvHost:     "infoblox.yourorg.org",
			},
			expected: "infoblox new dns provider could not get config from env: infoblox build config from env could not find value for " + EnvUser,
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUser:     "user",
				EnvPassword: "",
				EnvHost:     "infoblox.yourorg.org",
			},
			expected: "infoblox new dns provider could not get config from env: infoblox build config from env could not find value for " + EnvPassword,
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
				assert.NotNil(t, p.Config)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
