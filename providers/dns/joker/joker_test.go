package joker

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvUsername, EnvPassword, EnvMode).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected any
	}{
		{
			desc: "mode DMAPI (default)",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvPassword: "123",
			},
			expected: &dmapiProvider{},
		},
		{
			desc: "mode DMAPI",
			envVars: map[string]string{
				EnvMode:     modeDMAPI,
				EnvUsername: "123",
				EnvPassword: "123",
			},
			expected: &dmapiProvider{},
		},
		{
			desc: "mode SVC",
			envVars: map[string]string{
				EnvMode:     modeSVC,
				EnvUsername: "123",
				EnvPassword: "123",
			},
			expected: &svcProvider{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			fmt.Println(os.Getenv(EnvMode))

			p, err := NewDNSProvider()
			require.NoError(t, err)
			require.NotNil(t, p)

			assert.IsType(t, test.expected, p)
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		mode     string
		expected any
	}{
		{
			desc:     "mode DMAPI (default)",
			expected: &dmapiProvider{},
		},
		{
			desc:     "mode DMAPI",
			mode:     modeDMAPI,
			expected: &dmapiProvider{},
		},
		{
			desc:     "mode SVC",
			mode:     modeSVC,
			expected: &svcProvider{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = "123"
			config.Password = "123"
			config.APIMode = test.mode

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)

			assert.IsType(t, test.expected, p)
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
