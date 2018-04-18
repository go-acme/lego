package nifcloud

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	nifcloudLiveTest  bool
	nifcloudAccessKey string
	nifcloudSecretKey string
	nifcloudDomain    string
)

func init() {
	nifcloudAccessKey = os.Getenv("NIFCLOUD_ACCESS_KEY_ID")
	nifcloudSecretKey = os.Getenv("NIFCLOUD_SECRET_ACCESS_KEY")
	nifcloudDomain = os.Getenv("NIFCLOUD_DOMAIN")

	if len(nifcloudAccessKey) > 0 && len(nifcloudSecretKey) > 0 && len(nifcloudDomain) > 0 {
		nifcloudLiveTest = true
	}
}

func TestLivenifcloudPresent(t *testing.T) {
	if !nifcloudLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(nifcloudDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLivenifcloudCleanUp(t *testing.T) {
	if !nifcloudLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(nifcloudDomain, "", "123d==")
	assert.NoError(t, err)
}
