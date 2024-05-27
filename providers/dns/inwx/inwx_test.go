package inwx

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvSharedSecret,
	EnvSandbox,
	EnvTTL).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvUsername, EnvPassword, envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvPassword: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "",
			},
			expected: "inwx: some credentials information are missing: INWX_USERNAME,INWX_PASSWORD",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "456",
			},
			expected: "inwx: some credentials information are missing: INWX_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvPassword: "",
			},
			expected: "inwx: some credentials information are missing: INWX_PASSWORD",
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
		desc     string
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "123",
			password: "456",
		},
		{
			desc:     "missing credentials",
			expected: "inwx: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password

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

func TestLivePresentAndCleanup(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	envTest.Apply(map[string]string{
		EnvSandbox: "true",
		EnvTTL:     "3600", // In sandbox mode, the minimum allowed TTL is 3600
	})
	defer envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)

	// Verify that no error is thrown if record already exists
	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func Test_computeSleep(t *testing.T) {
	testCases := []struct {
		desc     string
		now      string
		expected time.Duration
	}{
		{
			desc:     "after 30s",
			now:      "2024-01-01T06:30:30Z",
			expected: 0 * time.Second,
		},
		{
			desc:     "0s",
			now:      "2024-01-01T06:30:00Z",
			expected: 0 * time.Second,
		},
		{
			desc:     "before 30s",
			now:      "2024-01-01T06:29:40Z", // 10 s
			expected: 20 * time.Second,
		},
	}

	previous, err := time.Parse(time.RFC3339, "2024-01-01T06:29:30Z")
	require.NoError(t, err)

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			now, err := time.Parse(time.RFC3339, test.now)
			require.NoError(t, err)

			d := &DNSProvider{previousUnlock: previous}

			sleep := d.computeSleep(now)
			assert.Equal(t, test.expected, sleep)
		})
	}
}
