package cloudru

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvServiceInstanceID,
	EnvKeyID,
	EnvSecret).
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
				EnvServiceInstanceID: "123",
				EnvKeyID:             "user",
				EnvSecret:            "secret",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "cloudru: some credentials information are missing: CLOUDRU_SERVICE_INSTANCE_ID,CLOUDRU_KEY_ID,CLOUDRU_SECRET",
		},
		{
			desc: "missing service instance ID",
			envVars: map[string]string{
				EnvServiceInstanceID: "",
				EnvKeyID:             "user",
				EnvSecret:            "secret",
			},
			expected: "cloudru: some credentials information are missing: CLOUDRU_SERVICE_INSTANCE_ID",
		},
		{
			desc: "missing key ID",
			envVars: map[string]string{
				EnvServiceInstanceID: "123",
				EnvKeyID:             "",
				EnvSecret:            "secret",
			},
			expected: "cloudru: some credentials information are missing: CLOUDRU_KEY_ID",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvServiceInstanceID: "123",
				EnvKeyID:             "user",
				EnvSecret:            "",
			},
			expected: "cloudru: some credentials information are missing: CLOUDRU_SECRET",
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
		desc              string
		serviceInstanceID string
		keyID             string
		secret            string
		expected          string
	}{
		{
			desc:              "success",
			serviceInstanceID: "123",
			keyID:             "user",
			secret:            "secret",
		},
		{
			desc:     "missing credentials",
			expected: "cloudru: some credentials information are missing",
		},
		{
			desc:              "missing service instance ID",
			serviceInstanceID: "",
			keyID:             "user",
			secret:            "secret",
			expected:          "cloudru: some credentials information are missing",
		},
		{
			desc:              "missing key ID",
			serviceInstanceID: "123",
			keyID:             "",
			secret:            "secret",
			expected:          "cloudru: some credentials information are missing",
		},
		{
			desc:              "missing secret",
			serviceInstanceID: "123",
			keyID:             "user",
			secret:            "",
			expected:          "cloudru: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ServiceInstanceID = test.serviceInstanceID
			config.KeyID = test.keyID
			config.Secret = test.secret

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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
