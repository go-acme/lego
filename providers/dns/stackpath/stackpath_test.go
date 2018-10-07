package stackpath

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

var (
	stackpathLiveTest     bool
	stackpathClientID     string
	stackpathClientSecret string
	stackpathStackID      string
	stackpathDomain       string
)

func init() {
	stackpathClientID = os.Getenv("STACKPATH_CLIENT_ID")
	stackpathClientSecret = os.Getenv("STACKPATH_CLIENT_SECRET")
	stackpathStackID = os.Getenv("STACKPATH_STACK_ID")
	stackpathDomain = os.Getenv("STACKPATH_DOMAIN")

	if len(stackpathClientID) > 0 &&
		len(stackpathClientSecret) > 0 &&
		len(stackpathStackID) > 0 &&
		len(stackpathDomain) > 0 {
		stackpathLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("STACKPATH_CLIENT_ID", stackpathClientID)
	os.Setenv("STACKPATH_CLIENT_SECRET", stackpathClientSecret)
	os.Setenv("STACKPATH_STACK_ID", stackpathStackID)
	os.Setenv("STACKPATH_DOMAIN", stackpathDomain)
}

func TestLiveStackpathPresent(t *testing.T) {
	if !stackpathLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(stackpathDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveStackpathCleanUp(t *testing.T) {
	if !stackpathLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(stackpathDomain, "", "123d==")
	require.NoError(t, err)
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
				"STACKPATH_CLIENT_ID":     "test@example.com",
				"STACKPATH_CLIENT_SECRET": "123",
				"STACKPATH_STACK_ID":      "ID",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"STACKPATH_CLIENT_ID":     "",
				"STACKPATH_CLIENT_SECRET": "",
				"STACKPATH_STACK_ID":      "",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_ID,STACKPATH_CLIENT_SECRET,STACKPATH_STACK_ID",
		},
		{
			desc: "missing client id",
			envVars: map[string]string{
				"STACKPATH_CLIENT_ID":     "",
				"STACKPATH_CLIENT_SECRET": "123",
				"STACKPATH_STACK_ID":      "ID",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				"STACKPATH_CLIENT_ID":     "test@example.com",
				"STACKPATH_CLIENT_SECRET": "",
				"STACKPATH_STACK_ID":      "ID",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_SECRET",
		},
		{
			desc: "missing stack id",
			envVars: map[string]string{
				"STACKPATH_CLIENT_ID":     "test@example.com",
				"STACKPATH_CLIENT_SECRET": "123",
				"STACKPATH_STACK_ID":      "",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_STACK_ID",
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
				assert.NoError(t, err)
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
		test := test
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			p, err := NewDNSProviderConfig(test.config)
			require.EqualError(t, err, test.expectedErr)
			assert.Nil(t, p)
		})
	}
}
