package mydnsjp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mydnsjpLiveTest bool
	mydnsjpMasterID string
	mydnsjpPassword string
	mydnsjpDomain   string
)

func init() {
	mydnsjpMasterID = os.Getenv("MYDNSJP_MASTER_ID")
	mydnsjpPassword = os.Getenv("MYDNSJP_PASSWORD")
	mydnsjpDomain = os.Getenv("MYDNSJP_DOMAIN")
	if len(mydnsjpMasterID) > 0 && len(mydnsjpPassword) > 0 && len(mydnsjpDomain) > 0 {
		mydnsjpLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("MYDNSJP_MASTER_ID", mydnsjpMasterID)
	os.Setenv("MYDNSJP_PASSWORD", mydnsjpPassword)
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
				"MYDNSJP_MASTER_ID": "test@example.com",
				"MYDNSJP_PASSWORD":  "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"MYDNSJP_MASTER_ID": "",
				"MYDNSJP_PASSWORD":  "",
			},
			expected: "mydnsjp: some credentials information are missing: MYDNSJP_MASTER_ID,MYDNSJP_PASSWORD",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				"MYDNSJP_MASTER_ID": "",
				"MYDNSJP_PASSWORD":  "key",
			},
			expected: "mydnsjp: some credentials information are missing: MYDNSJP_MASTER_ID",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"MYDNSJP_MASTER_ID": "awesome@possum.com",
				"MYDNSJP_PASSWORD":  "",
			},
			expected: "mydnsjp: some credentials information are missing: MYDNSJP_PASSWORD",
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
		desc     string
		masterID string
		password string
		expected string
	}{
		{
			desc:     "success",
			masterID: "test@example.com",
			password: "123",
		},
		{
			desc:     "missing credentials",
			expected: "mydnsjp: some credentials information are missing",
		},
		{
			desc:     "missing email",
			password: "123",
			expected: "mydnsjp: some credentials information are missing",
		},
		{
			desc:     "missing api key",
			masterID: "test@example.com",
			expected: "mydnsjp: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("MYDNSJP_MASTER_ID")
			os.Unsetenv("MYDNSJP_PASSWORD")

			config := NewDefaultConfig()
			config.MasterID = test.masterID
			config.Password = test.password

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

func TestLivePresent(t *testing.T) {
	if !mydnsjpLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.MasterID = mydnsjpMasterID
	config.Password = mydnsjpPassword

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(mydnsjpDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !mydnsjpLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(2 * time.Second)

	config := NewDefaultConfig()
	config.MasterID = mydnsjpMasterID
	config.Password = mydnsjpPassword

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(mydnsjpDomain, "", "123d==")
	assert.NoError(t, err)
}
