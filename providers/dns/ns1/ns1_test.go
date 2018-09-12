package ns1

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
	apiKey = os.Getenv("NS1_API_KEY")
	domain = os.Getenv("NS1_DOMAIN")
	if len(apiKey) > 0 && len(domain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("NS1_API_KEY", apiKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("NS1_API_KEY", "")

	config := NewDefaultConfig()
	config.APIKey = "123"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("NS1_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "ns1: some credentials information are missing: NS1_API_KEY")
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.APIKey = apiKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(domain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.APIKey = apiKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(domain, "", "123d==")
	assert.NoError(t, err)
}
