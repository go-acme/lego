package zoneedit

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUserID, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUserID:   "user",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing user ID",
			envVars: map[string]string{
				EnvUserID:   "",
				EnvPassword: "secret",
			},
			expected: "zoneedit: some credentials information are missing: ZONEEDIT_USER_ID",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUserID:   "user",
				EnvPassword: "",
			},
			expected: "zoneedit: some credentials information are missing: ZONEEDIT_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "zoneedit: some credentials information are missing: ZONEEDIT_USER_ID,ZONEEDIT_PASSWORD",
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
		password string
		expected string
	}{
		{
			desc:     "success",
			userID:   "user",
			password: "secret",
		},
		{
			desc:     "missing user ID",
			password: "secret",
			expected: "zoneedit: credentials missing",
		},
		{
			desc:     "missing password",
			userID:   "user",
			expected: "zoneedit: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "zoneedit: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.UserID = test.userID
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
