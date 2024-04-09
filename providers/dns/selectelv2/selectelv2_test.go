package selectelv2

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsernameOS, EnvPasswordOS, EnvAccount, EnvProjectID).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvPasswordOS: "qwerty",
				EnvAccount:    "1",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
		},
		{
			desc: "No username",
			envVars: map[string]string{
				EnvPasswordOS: "qwerty",
				EnvAccount:    "1",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_USERNAME",
		},
		{
			desc: "No password",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvAccount:    "1",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_PASSWORD",
		},
		{
			desc: "No account",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvPasswordOS: "qwerty",
				EnvProjectID:  "111a11111aaa11aa1a11aaa11111aa1a",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_ACCOUNT_ID",
		},
		{
			desc: "No project",
			envVars: map[string]string{
				EnvUsernameOS: "someName",
				EnvPasswordOS: "qwerty",
				EnvAccount:    "1",
			},
			expected: "selectelv2: some credentials information are missing: SELECTELV2_PROJECT_ID",
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
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
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
