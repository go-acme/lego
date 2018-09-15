package alidns

import (
	"os"
	"testing"
	"time"
	// "time"
	"github.com/stretchr/testify/assert"
)

var (
	alidnsLiveTest  bool
	alidnsAPIKey    string
	alidnsSecretKey string
	alidnsDomain    string
)

func init() {
	alidnsAPIKey = os.Getenv("ALICLOUD_ACCESS_KEY")
	alidnsSecretKey = os.Getenv("ALICLOUD_SECRET_KEY")
	alidnsDomain = os.Getenv("ALIDNS_DOMAIN")

	if len(alidnsAPIKey) > 0 && len(alidnsSecretKey) > 0 && len(alidnsDomain) > 0 {
		alidnsLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("ALICLOUD_ACCESS_KEY", alidnsAPIKey)
	os.Setenv("ALICLOUD_SECRET_KEY", alidnsSecretKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("ALICLOUD_ACCESS_KEY", "")
	os.Setenv("ALICLOUD_SECRET_KEY", "")

	config := NewDefaultConfig()
	config.APIKey = "123"
	config.SecretKey = "123"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("ALICLOUD_ACCESS_KEY", "123")
	os.Setenv("ALICLOUD_SECRET_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("ALICLOUD_ACCESS_KEY", "")
	os.Setenv("ALICLOUD_SECRET_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "alicloud: some credentials information are missing: ALICLOUD_ACCESS_KEY,ALICLOUD_SECRET_KEY")
}

func TestCloudXNSPresent(t *testing.T) {
	if !alidnsLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.APIKey = alidnsAPIKey
	config.SecretKey = alidnsSecretKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(alidnsDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLivednspodCleanUp(t *testing.T) {
	if !alidnsLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.APIKey = alidnsAPIKey
	config.SecretKey = alidnsSecretKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
	err = provider.CleanUp(alidnsDomain, "", "123d==")
	assert.NoError(t, err)
}
