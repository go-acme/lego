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
	alidnsAPIKey = os.Getenv("ALIDNS_API_KEY")
	alidnsSecretKey = os.Getenv("ALIDNS_SECRET_KEY")
	alidnsDomain = os.Getenv("ALIDNS_DOMAIN")

	if len(alidnsAPIKey) > 0 && len(alidnsSecretKey) > 0 && len(alidnsDomain) > 0 {
		alidnsLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("ALIDNS_API_KEY", alidnsAPIKey)
	os.Setenv("ALIDNS_SECRET_KEY", alidnsSecretKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("ALIDNS_API_KEY", "")
	os.Setenv("ALIDNS_SECRET_KEY", "")

	_, err := NewDNSProviderCredentials("123", "123", "")
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("ALIDNS_API_KEY", "123")
	os.Setenv("ALIDNS_SECRET_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDXNS_API_KEY", "")
	os.Setenv("CLOUDXNS_SECRET_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "AliDNS: some credentials information are missing: ALIDNS_API_KEY,ALIDNS_SECRET_KEY")
}

func TestCloudXNSPresent(t *testing.T) {
	if !alidnsLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(alidnsAPIKey, alidnsSecretKey, "")
	assert.NoError(t, err)

	err = provider.Present(alidnsDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLivednspodCleanUp(t *testing.T) {
	if !alidnsLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)
	provider, err := NewDNSProviderCredentials(alidnsAPIKey, alidnsSecretKey, "")
	assert.NoError(t, err)
	err = provider.CleanUp(alidnsDomain, "", "123d==")
	assert.NoError(t, err)
}
