package stackpath

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvClientID,
	EnvClientSecret,
	EnvStackID).
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
				EnvClientID:     "test@example.com",
				EnvClientSecret: "123",
				EnvStackID:      "ID",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "",
				EnvStackID:      "",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_ID,STACKPATH_CLIENT_SECRET,STACKPATH_STACK_ID",
		},
		{
			desc: "missing client id",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "123",
				EnvStackID:      "ID",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvClientID:     "test@example.com",
				EnvClientSecret: "",
				EnvStackID:      "ID",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_SECRET",
		},
		{
			desc: "missing stack id",
			envVars: map[string]string{
				EnvClientID:     "test@example.com",
				EnvClientSecret: "123",
				EnvStackID:      "",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_STACK_ID",
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
				assert.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := map[string]struct {
		config      *Config
		expectedErr string
	}{
		"no_config": {
			config:      nil,
			expectedErr: "stackpath: the configuration of the DNS provider is nil",
		},
		"no_client_id": {
			config: &Config{
				ClientSecret: "secret",
				StackID:      "stackID",
			},
			expectedErr: "stackpath: credentials missing",
		},
		"no_client_secret": {
			config: &Config{
				ClientID: "clientID",
				StackID:  "stackID",
			},
			expectedErr: "stackpath: credentials missing",
		},
		"no_stack_id": {
			config: &Config{
				ClientID:     "clientID",
				ClientSecret: "secret",
			},
			expectedErr: "stackpath: stack id missing",
		},
	}

	for desc, test := range testCases {
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			p, err := NewDNSProviderConfig(test.config)
			require.EqualError(t, err, test.expectedErr)
			assert.Nil(t, p)
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
