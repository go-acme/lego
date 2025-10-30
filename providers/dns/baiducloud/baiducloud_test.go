package baiducloud

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAccessKeyID, EnvSecretAccessKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccessKeyID:     "key",
				EnvSecretAccessKey: "secret",
			},
		},
		{
			desc: "missing access key ID",
			envVars: map[string]string{
				EnvAccessKeyID: "key",
			},
			expected: "baiducloud: some credentials information are missing: BAIDUCLOUD_SECRET_ACCESS_KEY",
		},
		{
			desc: "missing secret access key",
			envVars: map[string]string{
				EnvSecretAccessKey: "secret",
			},
			expected: "baiducloud: some credentials information are missing: BAIDUCLOUD_ACCESS_KEY_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "baiducloud: some credentials information are missing: BAIDUCLOUD_ACCESS_KEY_ID,BAIDUCLOUD_SECRET_ACCESS_KEY",
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
		desc            string
		accessKeyID     string
		secretAccessKey string
		expected        string
	}{
		{
			desc:            "success",
			accessKeyID:     "key",
			secretAccessKey: "secret",
		},
		{
			desc:            "missing access key ID",
			accessKeyID:     "",
			secretAccessKey: "secret",
			expected:        "baiducloud: accessKeyId should not be empty",
		},
		{
			desc:            "missing secret access key",
			accessKeyID:     "key",
			secretAccessKey: "",
			expected:        "baiducloud: secretKey should not be empty",
		},
		{
			desc:     "missing credentials",
			expected: "baiducloud: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKeyID = test.accessKeyID
			config.SecretAccessKey = test.secretAccessKey

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
