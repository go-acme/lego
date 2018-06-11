package he

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	heUsername string
	hePassword string
	heDomain   string
	heLiveTest bool
)

func init() {
	heUsername = os.Getenv("HE_USERNAME")
	hePassword = os.Getenv("HE_PASSWORD")
	heDomain = os.Getenv("HE_DOMAIN_NAME")

	if len(heUsername) > 0 && len(hePassword) > 0 && len(heDomain) > 0 {
		heLiveTest = true
	}
}

func TestNewDNSProvider(t *testing.T) {
	provider, err := NewDNSProvider()

	if !heLiveTest {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, provider)
		assert.NoError(t, err)
	}
}

func TestDNSProvider_Present(t *testing.T) {
	if !heLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(heDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	if !heLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(heDomain, "", "123d==")
	assert.NoError(t, err)
}
