package gigahostno

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/gigahostno/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvToken).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success with username/password",
			envVars: map[string]string{
				EnvUsername: "user@example.com",
				EnvPassword: "secret",
			},
		},
		{
			desc: "success with token",
			envVars: map[string]string{
				EnvToken: "test-token-12345",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "secret",
			},
			expected: "gigahostno: some credentials information are missing: GIGAHOSTNO_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "user@example.com",
				EnvPassword: "",
			},
			expected: "gigahostno: some credentials information are missing: GIGAHOSTNO_PASSWORD",
		},
		{
			desc:     "missing all credentials",
			envVars:  map[string]string{},
			expected: "gigahostno: some credentials information are missing: GIGAHOSTNO_USERNAME,GIGAHOSTNO_PASSWORD",
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
		desc     string
		username string
		password string
		token    string
		ttl      int
		expected string
	}{
		{
			desc:     "success with username/password",
			username: "user@example.com",
			password: "secret",
			ttl:      minTTL,
		},
		{
			desc:  "success with token",
			token: "test-token-12345",
			ttl:   minTTL,
		},
		{
			desc:     "missing credentials",
			username: "",
			password: "",
			token:    "",
			ttl:      minTTL,
			expected: "gigahostno: credentials missing (provide either GIGAHOSTNO_TOKEN or GIGAHOSTNO_USERNAME+GIGAHOSTNO_PASSWORD)",
		},
		{
			desc:     "invalid TTL with username/password",
			username: "user@example.com",
			password: "secret",
			ttl:      10,
			expected: "gigahostno: invalid TTL, TTL (10) must be greater than 60",
		},
		{
			desc:     "invalid TTL with token",
			token:    "test-token-12345",
			ttl:      10,
			expected: "gigahostno: invalid TTL, TTL (10) must be greater than 60",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.Token = test.token
			config.TTL = test.ttl

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

func TestGetPossibleZones(t *testing.T) {
	testCases := []struct {
		domain   string
		expected []string
	}{
		{
			domain:   "sub.example.com",
			expected: []string{"sub.example.com", "example.com"},
		},
		{
			domain:   "deep.sub.example.com",
			expected: []string{"deep.sub.example.com", "sub.example.com", "example.com"},
		},
		{
			domain:   "example.no",
			expected: []string{"example.no"},
		},
	}

	for _, test := range testCases {
		t.Run(test.domain, func(t *testing.T) {
			zones := getPossibleZones(test.domain)
			assert.Equal(t, test.expected, zones)
		})
	}
}

func TestTokenCaching(t *testing.T) {
	t.Run("username/password auth", func(t *testing.T) {
		config := NewDefaultConfig()
		config.Username = "user@example.com"
		config.Password = "secret"
		config.TTL = minTTL

		provider, err := NewDNSProviderConfig(config)
		require.NoError(t, err)

		// Verify no token initially
		assert.Nil(t, provider.token)

		// Note: We can't easily test actual caching without mocking the client,
		// but we can verify the token field exists and follows the expected pattern
		assert.NotNil(t, provider.client)
		assert.Empty(t, provider.config.Token)
	})

	t.Run("token-based auth", func(t *testing.T) {
		config := NewDefaultConfig()
		config.Token = "test-token-12345"
		config.TTL = minTTL

		provider, err := NewDNSProviderConfig(config)
		require.NoError(t, err)

		// Verify token is in config but not cached
		assert.Equal(t, "test-token-12345", provider.config.Token)
		assert.Nil(t, provider.token)
		assert.NotNil(t, provider.client)
	})
}

func TestFindBestZone(t *testing.T) {
	zones := []internal.Zone{
		{
			ID:     "1",
			Name:   "example.com",
			Active: "1",
		},
		{
			ID:     "2",
			Name:   "sub.example.com",
			Active: "1",
		},
		{
			ID:     "3",
			Name:   "other.com",
			Active: "1",
		},
		{
			ID:     "4",
			Name:   "inactive.com",
			Active: "0",
		},
	}

	testCases := []struct {
		domain   string
		expected string
	}{
		{
			domain:   "test.sub.example.com",
			expected: "sub.example.com",
		},
		{
			domain:   "www.example.com",
			expected: "example.com",
		},
		{
			domain:   "other.com",
			expected: "other.com",
		},
		{
			domain:   "inactive.com",
			expected: "",
		},
		{
			domain:   "notfound.com",
			expected: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.domain, func(t *testing.T) {
			zone := findBestZone(zones, test.domain)
			if test.expected == "" {
				assert.Nil(t, zone)
			} else {
				require.NotNil(t, zone)
				assert.Equal(t, test.expected, zone.Name)
			}
		})
	}
}
