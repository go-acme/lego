package alicloud

import (
	"os"
	"github.com/stretchr/testify/assert"
	"time"
	"testing"
)

var (
	liveTest  bool
	apiKey    string
	apiSecret string
	domain    string
)

func init() {
	apiKey = os.Getenv("ALICLOUD_API_KEY")
	apiSecret = os.Getenv("ALICLOUD_API_SECRET")
	domain = os.Getenv("ALICLOUD_TEST_DOMAIN")
	liveTest = len(apiKey) > 0 && len(apiSecret) > 0 && len(domain) > 0
}

func restoreEnv() {
	os.Setenv("ALICLOUD_API_KEY", apiKey)
	os.Setenv("ALICLOUD_API_SECRET", apiSecret)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("ALICLOUD_API_KEY", "123")
	os.Setenv("ALICLOUD_API_SECRET", "123")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.NoError(t, err)
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

	time.Sleep(time.Second * 60)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(domain, "", "123d==")
	assert.NoError(t, err)
}