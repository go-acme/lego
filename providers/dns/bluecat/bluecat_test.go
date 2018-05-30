package bluecat

import (
	"os"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

var (
	bluecatLiveTest   bool
	bluecatServer     string
	bluecatUserName   string
	bluecatPassword   string
	bluecatConfigName string
	bluecatDNSView    string
	bluecatDomain     string
)

func init() {
	bluecatServer = os.Getenv("BLUECAT_SERVER_URL")
	bluecatUserName = os.Getenv("BLUECAT_USER_NAME")
	bluecatPassword = os.Getenv("BLUECAT_PASSWORD")
	bluecatDomain = os.Getenv("BLUECAT_DOMAIN")
	bluecatConfigName = os.Getenv("BLUECAT_CONFIG_NAME")
	bluecatDNSView = os.Getenv("BLUECAT_DNS_VIEW")
	if len(bluecatServer) > 0 &&
		len(bluecatDomain) > 0 &&
		len(bluecatUserName) > 0 &&
		len(bluecatPassword) > 0 &&
		len(bluecatConfigName) > 0 &&
		len(bluecatDNSView) > 0 {
		bluecatLiveTest = true
	}
}

func TestLiveBluecatPresent(t *testing.T) {
	if !bluecatLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(bluecatDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveBluecatCleanUp(t *testing.T) {
	if !bluecatLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(bluecatDomain, "", "123d==")
	assert.NoError(t, err)
}
