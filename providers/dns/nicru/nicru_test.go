package nicru

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const (
	fakeServiceID = "2519234972459cdfa23423adf143324f"
	fakeSecret    = "oo5ahrie0aiPho3Vee4siupoPhahdahCh1thiesohru"
	fakeUsername  = "1234567/NIC-D"
	fakePassword  = "einge8Goo2eBaiXievuj"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword, EnvServiceID, EnvSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvSecret:    fakeSecret,
				EnvUsername:  fakeUsername,
				EnvPassword:  fakePassword,
			},
			expected: "nicru: failed to create oauth2 token: oauth2: \"unauthorized_client\"",
		},
		{
			desc: "missing serviceID",
			envVars: map[string]string{
				EnvSecret:   fakeSecret,
				EnvUsername: fakeUsername,
				EnvPassword: fakePassword,
			},
			expected: "nicru: some credentials information are missing: NICRU_SERVICE_ID",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvUsername:  fakeUsername,
				EnvPassword:  fakePassword,
			},
			expected: "nicru: some credentials information are missing: NICRU_SECRET",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvSecret:    fakeSecret,
				EnvPassword:  fakePassword,
			},
			expected: "nicru: some credentials information are missing: NICRU_USER",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServiceID: fakeServiceID,
				EnvSecret:    fakeSecret,
				EnvUsername:  fakeUsername,
			},
			expected: "nicru: some credentials information are missing: NICRU_PASSWORD",
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
		config   *Config
		expected string
	}{
		{
			desc: "success",
			config: &Config{
				ServiceID: fakeServiceID,
				Secret:    fakeSecret,
				Username:  fakeUsername,
				Password:  fakePassword,
			},
			expected: "nicru: failed to create oauth2 token: oauth2: \"unauthorized_client\"",
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "nicru: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing username",
			config: &Config{
				ServiceID: fakeServiceID,
				Password:  fakePassword,
			},
			expected: "nicru: username is missing in credentials information",
		},
		{
			desc: "missing password",
			config: &Config{
				ServiceID: fakeServiceID,
				Secret:    fakeSecret,
				Username:  fakeUsername,
			},
			expected: "nicru: password is missing in credentials information",
		},
		{
			desc: "missing secret",
			config: &Config{
				ServiceID: fakeServiceID,
				Username:  fakeUsername,
				Password:  fakePassword,
			},
			expected: "nicru: secret is missing in credentials information",
		},
		{
			desc: "missing serviceID",
			config: &Config{
				Secret:   fakeSecret,
				Username: fakeUsername,
				Password: fakePassword,
			},
			expected: "nicru: serviceID is missing in credentials information",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

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
