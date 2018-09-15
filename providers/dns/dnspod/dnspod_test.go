package dnspod

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	dnspodLiveTest bool
	dnspodAPIKey   string
	dnspodDomain   string
)

func init() {
	dnspodAPIKey = os.Getenv("DNSPOD_API_KEY")
	dnspodDomain = os.Getenv("DNSPOD_DOMAIN")
	if len(dnspodAPIKey) > 0 && len(dnspodDomain) > 0 {
		dnspodLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("DNSPOD_API_KEY", dnspodAPIKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("DNSPOD_API_KEY", "")

	config := NewDefaultConfig()
	config.LoginToken = "123"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}
func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("DNSPOD_API_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("DNSPOD_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "dnspod: some credentials information are missing: DNSPOD_API_KEY")
}

func TestLivednspodPresent(t *testing.T) {
	if !dnspodLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.LoginToken = dnspodAPIKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(dnspodDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLivednspodCleanUp(t *testing.T) {
	if !dnspodLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.LoginToken = dnspodAPIKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(dnspodDomain, "", "123d==")
	assert.NoError(t, err)
}
