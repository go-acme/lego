package hostingde

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest      bool
	envTestAPIKey string
	envTestZone   string
	envTestDomain string
)

func init() {
	envTestAPIKey = os.Getenv("HOSTINGDE_API_KEY")
	envTestZone = os.Getenv("HOSTINGDE_ZONE_NAME")
	envTestDomain = os.Getenv("HOSTINGDE_DOMAIN")
	if len(envTestZone) > 0 && len(envTestAPIKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("HOSTINGDE_ZONE_NAME", envTestZone)
	os.Setenv("HOSTINGDE_API_KEY", envTestAPIKey)
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
				"HOSTINGDE_API_KEY":   "123",
				"HOSTINGDE_ZONE_NAME": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"HOSTINGDE_API_KEY":   "",
				"HOSTINGDE_ZONE_NAME": "",
			},
			expected: "hostingde: some credentials information are missing: HOSTINGDE_API_KEY,HOSTINGDE_ZONE_NAME",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				"HOSTINGDE_API_KEY":   "",
				"HOSTINGDE_ZONE_NAME": "456",
			},
			expected: "hostingde: some credentials information are missing: HOSTINGDE_API_KEY",
		},
		{
			desc: "missing zone name",
			envVars: map[string]string{
				"HOSTINGDE_API_KEY":   "123",
				"HOSTINGDE_ZONE_NAME": "",
			},
			expected: "hostingde: some credentials information are missing: HOSTINGDE_ZONE_NAME",
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
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		zoneName string
		expected string
	}{
		{
			desc:     "success",
			apiKey:   "123",
			zoneName: "456",
		},
		{
			desc:     "missing credentials",
			expected: "hostingde: API key missing",
		},
		{
			desc:     "missing api key",
			zoneName: "456",
			expected: "hostingde: API key missing",
		},
		{
			desc:     "missing zone name",
			apiKey:   "123",
			expected: "hostingde: Zone Name missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("HOSTINGDE_API_KEY")
			os.Unsetenv("HOSTINGDE_ZONE_NAME")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.ZoneName = test.zoneName

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
