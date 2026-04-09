package ucloud

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvPrivateKey,
	EnvPublicKey,
	EnvRegion,
	EnvProjectID,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvPrivateKey: "xxx",
				EnvPublicKey:  "yyy",
			},
		},
		{
			desc: "missing private key",
			envVars: map[string]string{
				EnvPrivateKey: "",
				EnvPublicKey:  "yyy",
			},
			expected: "ucloud: some credentials information are missing: UCLOUD_PRIVATE_KEY",
		},
		{
			desc: "missing public key",
			envVars: map[string]string{
				EnvPrivateKey: "xxx",
				EnvPublicKey:  "",
			},
			expected: "ucloud: some credentials information are missing: UCLOUD_PUBLIC_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "ucloud: some credentials information are missing: UCLOUD_PUBLIC_KEY,UCLOUD_PRIVATE_KEY",
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
		desc       string
		privateKey string
		publicKey  string
		expected   string
	}{
		{
			desc:       "success",
			privateKey: "xxx",
			publicKey:  "yyy",
		},
		{
			desc:      "missing private key",
			publicKey: "yyy",
			expected:  "ucloud: credentials missing",
		},
		{
			desc:       "missing public key",
			privateKey: "xxx",
			expected:   "ucloud: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "ucloud: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.PrivateKey = test.privateKey
			config.PublicKey = test.publicKey

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
