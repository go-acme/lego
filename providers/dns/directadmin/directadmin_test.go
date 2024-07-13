package directadmin

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const (
	envNamespace = "DIRECTADMIN_"
	envAPIURL    = envNamespace + "API_URL"
	envUsername  = envNamespace + "USERNAME"
	envPassword  = envNamespace + "PASSWORD"
	envDomain    = envNamespace + "DOMAIN"
)

var envTest = tester.NewEnvTest(envAPIURL, envUsername, envPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				envAPIURL:   "https://api.directadmin.com",
				envUsername: "username",
				envPassword: "password",
			},
		},
		{
			desc:     "missing API URL",
			envVars:  map[string]string{},
			expected: "directadmin: some credentials information are missing: DIRECTADMIN_API_URL, DIRECTADMIN_USERNAME, DIRECTADMIN_PASSWORD",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiURL   string
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			apiURL:   "https://api.directadmin.com",
			username: "username",
			password: "password",
		},
		{
			desc:     "missing API URL",
			expected: "directadmin: APIURL is missing",
		},
		{
			desc:     "missing username",
			apiURL:   "https://api.directadmin.com",
			expected: "directadmin: username is missing",
		},
		{
			desc:     "missing password",
			apiURL:   "https://api.directadmin.com",
			username: "username",
			expected: "directadmin: password is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIURL = test.apiURL
			config.Username = test.username
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
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
