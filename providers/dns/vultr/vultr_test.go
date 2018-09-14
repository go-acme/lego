package vultr

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	liveTest bool
	apiKey   string
	domain   string
)

func init() {
	apiKey = os.Getenv("VULTR_API_KEY")
	domain = os.Getenv("VULTR_TEST_DOMAIN")
	liveTest = len(apiKey) > 0 && len(domain) > 0
}

func restoreEnv() {
	os.Setenv("VULTR_API_KEY", apiKey)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("VULTR_API_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("VULTR_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "vultr: some credentials information are missing: VULTR_API_KEY")
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
