package fastdns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fastdnsLiveTest bool
	host            string
	clientToken     string
	clientSecret    string
	accessToken     string
	testDomain      string
)

func init() {
	host = os.Getenv("AKAMAI_HOST")
	clientToken = os.Getenv("AKAMAI_CLIENT_TOKEN")
	clientSecret = os.Getenv("AKAMAI_CLIENT_SECRET")
	accessToken = os.Getenv("AKAMAI_ACCESS_TOKEN")
	testDomain = os.Getenv("AKAMAI_TEST_DOMAIN")

	if len(host) > 0 && len(clientToken) > 0 && len(clientSecret) > 0 && len(accessToken) > 0 {
		fastdnsLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("AKAMAI_HOST", host)
	os.Setenv("AKAMAI_CLIENT_TOKEN", clientToken)
	os.Setenv("AKAMAI_CLIENT_SECRET", clientSecret)
	os.Setenv("AKAMAI_ACCESS_TOKEN", accessToken)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AKAMAI_HOST", "")
	os.Setenv("AKAMAI_CLIENT_TOKEN", "")
	os.Setenv("AKAMAI_CLIENT_SECRET", "")
	os.Setenv("AKAMAI_ACCESS_TOKEN", "")

	config := NewDefaultConfig()
	config.Host = "somehost"
	config.ClientToken = "someclienttoken"
	config.ClientSecret = "someclientsecret"
	config.AccessToken = "someaccesstoken"

	_, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AKAMAI_HOST", "somehost")
	os.Setenv("AKAMAI_CLIENT_TOKEN", "someclienttoken")
	os.Setenv("AKAMAI_CLIENT_SECRET", "someclientsecret")
	os.Setenv("AKAMAI_ACCESS_TOKEN", "someaccesstoken")

	_, err := NewDNSProvider()
	require.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AKAMAI_HOST", "")
	os.Setenv("AKAMAI_CLIENT_TOKEN", "")
	os.Setenv("AKAMAI_CLIENT_SECRET", "")
	os.Setenv("AKAMAI_ACCESS_TOKEN", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "fastdns: some credentials information are missing: AKAMAI_HOST,AKAMAI_CLIENT_TOKEN,AKAMAI_CLIENT_SECRET,AKAMAI_ACCESS_TOKEN")
}

func TestLiveFastdnsPresent(t *testing.T) {
	if !fastdnsLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.Host = host
	config.ClientToken = clientToken
	config.ClientSecret = clientSecret
	config.AccessToken = accessToken

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(testDomain, "", "123d==")
	require.NoError(t, err)

	// Present Twice to handle create / update
	err = provider.Present(testDomain, "", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_findZoneAndRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.Host = "somehost"
	config.ClientToken = "someclienttoken"
	config.ClientSecret = "someclientsecret"
	config.AccessToken = "someaccesstoken"

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	type expected struct {
		zone       string
		recordName string
	}

	testCases := []struct {
		desc     string
		fqdn     string
		domain   string
		expected expected
	}{
		{
			desc:   "Extract root record name",
			fqdn:   "_acme-challenge.bar.com.",
			domain: "bar.com",
			expected: expected{
				zone:       "bar.com",
				recordName: "_acme-challenge",
			},
		},
		{
			desc:   "Extract sub record name",
			fqdn:   "_acme-challenge.foo.bar.com.",
			domain: "foo.bar.com",
			expected: expected{
				zone:       "bar.com",
				recordName: "_acme-challenge.foo",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, recordName, err := provider.findZoneAndRecordName(test.fqdn, test.domain)
			require.NoError(t, err)
			assert.Equal(t, test.expected.zone, zone)
			assert.Equal(t, test.expected.recordName, recordName)
		})
	}
}

func TestLiveFastdnsCleanUp(t *testing.T) {
	if !fastdnsLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.Host = host
	config.ClientToken = clientToken
	config.ClientSecret = clientSecret
	config.AccessToken = accessToken

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(testDomain, "", "123d==")
	require.NoError(t, err)
}
