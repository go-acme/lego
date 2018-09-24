package cloudflare

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cflareLiveTest bool
	cflareEmail    string
	cflareAPIKey   string
	cflareDomain   string
)

func init() {
	cflareEmail = os.Getenv("CLOUDFLARE_EMAIL")
	cflareAPIKey = os.Getenv("CLOUDFLARE_API_KEY")
	cflareDomain = os.Getenv("CLOUDFLARE_DOMAIN")
	if len(cflareEmail) > 0 && len(cflareAPIKey) > 0 && len(cflareDomain) > 0 {
		cflareLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("CLOUDFLARE_EMAIL", cflareEmail)
	os.Setenv("CLOUDFLARE_API_KEY", cflareAPIKey)
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
				"CLOUDFLARE_EMAIL":   "test@example.com",
				"CLOUDFLARE_API_KEY": "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":   "",
				"CLOUDFLARE_API_KEY": "",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL,CLOUDFLARE_API_KEY",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":   "",
				"CLOUDFLARE_API_KEY": "key",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":   "awesome@possum.com",
				"CLOUDFLARE_API_KEY": "",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_API_KEY",
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
	testCases := []struct {
		desc      string
		authEmail string
		authKey   string
		expected  string
	}{
		{
			desc:      "success",
			authEmail: "test@example.com",
			authKey:   "123",
		},
		{
			desc:     "missing credentials",
			expected: "invalid credentials: key & email must not be empty",
		},
		{
			desc:     "missing email",
			authKey:  "123",
			expected: "invalid credentials: key & email must not be empty",
		},
		{
			desc:      "missing api key",
			authEmail: "test@example.com",
			expected:  "invalid credentials: key & email must not be empty",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("CLOUDFLARE_EMAIL")
			os.Unsetenv("CLOUDFLARE_API_KEY")

			config := NewDefaultConfig()
			config.AuthEmail = test.authEmail
			config.AuthKey = test.authKey

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				assert.NoError(t, err)
				assert.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestCloudFlarePresent(t *testing.T) {
	if !cflareLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.AuthEmail = cflareEmail
	config.AuthKey = cflareAPIKey

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(cflareDomain, "", "123d==")
	require.NoError(t, err)
}

func TestCloudFlareCleanUp(t *testing.T) {
	if !cflareLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	config := NewDefaultConfig()
	config.AuthEmail = cflareEmail
	config.AuthKey = cflareAPIKey

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(cflareDomain, "", "123d==")
	require.NoError(t, err)
}
