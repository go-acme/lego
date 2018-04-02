package glesys

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	glesysAPIUser  string
	glesysAPIKey   string
	glesysDomain   string
	glesysLiveTest bool
)

func init() {
	glesysAPIUser = os.Getenv("GLESYS_API_USER")
	glesysAPIKey = os.Getenv("GLESYS_API_KEY")
	glesysDomain = os.Getenv("GLESYS_DOMAIN")

	if len(glesysAPIUser) > 0 && len(glesysAPIKey) > 0 && len(glesysDomain) > 0 {
		glesysLiveTest = true
	}
}

func TestNewDNSProvider(t *testing.T) {
	provider, err := NewDNSProvider()

	if !glesysLiveTest {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, provider)
		assert.NoError(t, err)
	}
}

func TestDNSProvider_Present(t *testing.T) {
	if !glesysLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(glesysDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	if !glesysLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(glesysDomain, "", "123d==")
	assert.NoError(t, err)
}
