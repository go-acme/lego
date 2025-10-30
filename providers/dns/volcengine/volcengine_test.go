package volcengine

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAccessKey,
	EnvSecretKey,
	EnvRegion,
	EnvHost,
	EnvScheme).
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
				EnvAccessKey: "access",
				EnvSecretKey: "secret",
			},
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				EnvSecretKey: "secret",
			},
			expected: "volcengine: some credentials information are missing: VOLC_ACCESSKEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAccessKey: "access",
			},
			expected: "volcengine: some credentials information are missing: VOLC_SECRETKEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "volcengine: some credentials information are missing: VOLC_ACCESSKEY,VOLC_SECRETKEY",
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
		desc      string
		expected  string
		accessKey string
		secretKey string
	}{
		{
			desc:      "success",
			accessKey: "access",
			secretKey: "secret",
		},
		{
			desc:      "missing access key",
			secretKey: "secret",
			expected:  "volcengine: missing credentials",
		},
		{
			desc:      "missing secret key",
			accessKey: "access",
			expected:  "volcengine: missing credentials",
		},
		{
			desc:     "missing credentials",
			expected: "volcengine: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey

			p, err := NewDNSProviderConfig(config)

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
