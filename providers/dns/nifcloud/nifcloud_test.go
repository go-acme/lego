package nifcloud

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest         bool
	envTestAccessKey string
	envTestSecretKey string
	envTestDomain    string
)

func init() {
	envTestAccessKey = os.Getenv("NIFCLOUD_ACCESS_KEY_ID")
	envTestSecretKey = os.Getenv("NIFCLOUD_SECRET_ACCESS_KEY")
	envTestDomain = os.Getenv("NIFCLOUD_DOMAIN")

	if len(envTestAccessKey) > 0 && len(envTestSecretKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("NIFCLOUD_ACCESS_KEY_ID", envTestAccessKey)
	os.Setenv("NIFCLOUD_SECRET_ACCESS_KEY", envTestSecretKey)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"NIFCLOUD_ACCESS_KEY_ID":     "123",
				"NIFCLOUD_SECRET_ACCESS_KEY": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"NIFCLOUD_ACCESS_KEY_ID":     "",
				"NIFCLOUD_SECRET_ACCESS_KEY": "",
			},
			expected: "nifcloud: some credentials information are missing: NIFCLOUD_ACCESS_KEY_ID,NIFCLOUD_SECRET_ACCESS_KEY",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				"NIFCLOUD_ACCESS_KEY_ID":     "",
				"NIFCLOUD_SECRET_ACCESS_KEY": "456",
			},
			expected: "nifcloud: some credentials information are missing: NIFCLOUD_ACCESS_KEY_ID",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				"NIFCLOUD_ACCESS_KEY_ID":     "123",
				"NIFCLOUD_SECRET_ACCESS_KEY": "",
			},
			expected: "nifcloud: some credentials information are missing: NIFCLOUD_SECRET_ACCESS_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

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
		desc      string
		accessKey string
		secretKey string
		expected  string
	}{
		{
			desc:      "success",
			accessKey: "123",
			secretKey: "456",
		},
		{
			desc:     "missing credentials",
			expected: "nifcloud: credentials missing",
		},
		{
			desc:      "missing api key",
			secretKey: "456",
			expected:  "nifcloud: credentials missing",
		},
		{
			desc:      "missing secret key",
			accessKey: "123",
			expected:  "nifcloud: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("NIFCLOUD_ACCESS_KEY_ID")
			os.Unsetenv("NIFCLOUD_SECRET_ACCESS_KEY")

			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey

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
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
