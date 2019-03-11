package cloudns

import (
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"CLOUDNS_AUTH_ID",
	"CLOUDNS_AUTH_PASSWORD").
	WithDomain("CLOUDNS_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"CLOUDNS_AUTH_ID":       "123",
				"CLOUDNS_AUTH_PASSWORD": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"CLOUDNS_AUTH_ID":       "",
				"CLOUDNS_AUTH_PASSWORD": "",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_ID,CLOUDNS_AUTH_PASSWORD",
		},
		{
			desc: "missing auth-id",
			envVars: map[string]string{
				"CLOUDNS_AUTH_ID":       "",
				"CLOUDNS_AUTH_PASSWORD": "456",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_ID",
		},
		{
			desc: "missing auth-password",
			envVars: map[string]string{
				"CLOUDNS_AUTH_ID":       "123",
				"CLOUDNS_AUTH_PASSWORD": "",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_PASSWORD",
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
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc         string
		authID       string
		authPassword string
		expected     string
	}{
		{
			desc:         "success",
			authID:       "123",
			authPassword: "456",
		},
		{
			desc:     "missing credentials",
			expected: "ClouDNS: credentials missing: authID",
		},
		{
			desc:         "missing auth-id",
			authPassword: "456",
			expected:     "ClouDNS: credentials missing: authID",
		},
		{
			desc:     "missing auth-password",
			authID:   "123",
			expected: "ClouDNS: credentials missing: authPassword",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthID = test.authID
			config.AuthPassword = test.authPassword

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
