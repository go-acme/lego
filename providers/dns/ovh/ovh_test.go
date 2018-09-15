package ovh

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	liveTest          bool
	apiEndpoint       string
	applicationKey    string
	applicationSecret string
	consumerKey       string
	domain            string
)

func init() {
	apiEndpoint = os.Getenv("OVH_ENDPOINT")
	applicationKey = os.Getenv("OVH_APPLICATION_KEY")
	applicationSecret = os.Getenv("OVH_APPLICATION_SECRET")
	consumerKey = os.Getenv("OVH_CONSUMER_KEY")
	liveTest = len(apiEndpoint) > 0 && len(applicationKey) > 0 && len(applicationSecret) > 0 && len(consumerKey) > 0
}

func restoreEnv() {
	os.Setenv("OVH_ENDPOINT", apiEndpoint)
	os.Setenv("OVH_APPLICATION_KEY", applicationKey)
	os.Setenv("OVH_APPLICATION_SECRET", applicationSecret)
	os.Setenv("OVH_CONSUMER_KEY", consumerKey)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("OVH_ENDPOINT", "ovh-eu")
	os.Setenv("OVH_APPLICATION_KEY", "1234")
	os.Setenv("OVH_APPLICATION_SECRET", "5678")
	os.Setenv("OVH_CONSUMER_KEY", "abcde")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "missing OVH_ENDPOINT",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "",
				"OVH_APPLICATION_KEY":    "1234",
				"OVH_APPLICATION_SECRET": "5678",
				"OVH_CONSUMER_KEY":       "abcde",
			},
			expected: "ovh: some credentials information are missing: OVH_ENDPOINT",
		},
		{
			desc: "missing OVH_APPLICATION_KEY",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "",
				"OVH_APPLICATION_SECRET": "5678",
				"OVH_CONSUMER_KEY":       "abcde",
			},
			expected: "ovh: some credentials information are missing: OVH_APPLICATION_KEY",
		},
		{
			desc: "missing OVH_APPLICATION_SECRET",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "1234",
				"OVH_APPLICATION_SECRET": "",
				"OVH_CONSUMER_KEY":       "abcde",
			},
			expected: "ovh: some credentials information are missing: OVH_APPLICATION_SECRET",
		},
		{
			desc: "missing OVH_CONSUMER_KEY",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "ovh-eu",
				"OVH_APPLICATION_KEY":    "1234",
				"OVH_APPLICATION_SECRET": "5678",
				"OVH_CONSUMER_KEY":       "",
			},
			expected: "ovh: some credentials information are missing: OVH_CONSUMER_KEY",
		},
		{
			desc: "all missing",
			envVars: map[string]string{
				"OVH_ENDPOINT":           "",
				"OVH_APPLICATION_KEY":    "",
				"OVH_APPLICATION_SECRET": "",
				"OVH_CONSUMER_KEY":       "",
			},
			expected: "ovh: some credentials information are missing: OVH_ENDPOINT,OVH_APPLICATION_KEY,OVH_APPLICATION_SECRET,OVH_CONSUMER_KEY",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {

			for key, value := range test.envVars {
				os.Setenv(key, value)
			}

			_, err := NewDNSProvider()
			assert.EqualError(t, err, test.expected)
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(domain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(domain, "", "123d==")
	assert.NoError(t, err)
}
