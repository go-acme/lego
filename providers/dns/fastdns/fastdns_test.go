package fastdns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AKAMAI_HOST", "somehost")
	os.Setenv("AKAMAI_CLIENT_TOKEN", "someclienttoken")
	os.Setenv("AKAMAI_CLIENT_SECRET", "someclientsecret")
	os.Setenv("AKAMAI_ACCESS_TOKEN", "someaccesstoken")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	err = provider.Present(testDomain, "", "123d==")
	assert.NoError(t, err)

	// Present Twice to handle create / update
	err = provider.Present(testDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestExtractRootRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.Host = "somehost"
	config.ClientToken = "someclienttoken"
	config.ClientSecret = "someclientsecret"
	config.AccessToken = "someaccesstoken"

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	zone, recordName, err := provider.findZoneAndRecordName("_acme-challenge.bar.com.", "bar.com")
	assert.NoError(t, err)
	assert.Equal(t, "bar.com", zone)
	assert.Equal(t, "_acme-challenge", recordName)
}

func TestExtractSubRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.Host = "somehost"
	config.ClientToken = "someclienttoken"
	config.ClientSecret = "someclientsecret"
	config.AccessToken = "someaccesstoken"

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	zone, recordName, err := provider.findZoneAndRecordName("_acme-challenge.foo.bar.com.", "foo.bar.com")
	assert.NoError(t, err)
	assert.Equal(t, "bar.com", zone)
	assert.Equal(t, "_acme-challenge.foo", recordName)
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
	assert.NoError(t, err)

	err = provider.CleanUp(testDomain, "", "123d==")
	assert.NoError(t, err)
}
