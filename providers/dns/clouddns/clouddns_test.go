package clouddns

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"CLOUDDNS_CLIENT_ID",
	"CLOUDDNS_EMAIL",
	"CLOUDDNS_PASSWORD").
	WithDomain("CLOUDDNS_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success client_id, email, password",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "client123",
				"CLOUDDNS_EMAIL":     "test@example.com",
				"CLOUDDNS_PASSWORD":  "password123",
			},
		},
		{
			desc: "missing clientId",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "",
				"CLOUDDNS_EMAIL":     "test@example.com",
				"CLOUDDNS_PASSWORD":  "password123",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_CLIENT_ID",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "client123",
				"CLOUDDNS_EMAIL":     "",
				"CLOUDDNS_PASSWORD":  "password123",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_EMAIL",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "client123",
				"CLOUDDNS_EMAIL":     "test@example.com",
				"CLOUDDNS_PASSWORD":  "",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_PASSWORD",
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
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success client_id, email, password",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "client123",
				"CLOUDDNS_EMAIL":     "test@example.com",
				"CLOUDDNS_PASSWORD":  "password123",
			},
		},
		{
			desc: "missing clientId",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "",
				"CLOUDDNS_EMAIL":     "test@example.com",
				"CLOUDDNS_PASSWORD":  "password123",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_CLIENT_ID",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "client123",
				"CLOUDDNS_EMAIL":     "",
				"CLOUDDNS_PASSWORD":  "password123",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_EMAIL",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				"CLOUDDNS_CLIENT_ID": "client123",
				"CLOUDDNS_EMAIL":     "test@example.com",
				"CLOUDDNS_PASSWORD":  "",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			p, err := NewDNSProviderConfig()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.ClientID)
				require.NotNil(t, p.Email)
				require.NotNil(t, p.Password)
				require.NotNil(t, p.TTL)
				require.NotNil(t, p.PropagationTimeout)
				require.NotNil(t, p.PollingInterval)
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
