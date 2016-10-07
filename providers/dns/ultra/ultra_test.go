package ultra

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	ultraLiveTest bool
	ultraUserName string
	ultraPassword string
	ultraDomain   string
)

func init() {
	ultraUserName = os.Getenv("ULTRA_USER_NAME")
	ultraPassword = os.Getenv("ULTRA_PASSWORD")
	ultraDomain = os.Getenv("ULTRA_DOMAIN")
	if len(ultraUserName) > 0 && len(ultraPassword) > 0 && len(ultraDomain) > 0 {
		ultraLiveTest = true
	}
}

func TestLiveUltraDNSPresent(t *testing.T) {
	if !ultraLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(ultraDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveUltraDNSCleanUp(t *testing.T) {
	if !ultraLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(ultraDomain, "", "123d==")
	assert.NoError(t, err)
}
