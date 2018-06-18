package hurricanedns

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
	heDomain = os.Getenv("HE_DOMAIN_NAME")
	heUsername = os.Getenv("HE_USERNAME")
	hePassword = os.Getenv("HE_PASSWORD")

	if len(heUsername) > 0 && len(hePassword) > 0 && len(heDomain) > 0 {
		heLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("HE_USERNAME", heUsername)
	os.Setenv("HE_PASSWORD", hePassword)
	os.Setenv("HE_DOMAIN_NAME", heDomain)
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

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("HE_USERNAME", "")
	os.Setenv("HE_PASSWORD", "")
	os.Setenv("HE_DOMAIN_NAME", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "Hurricanedns: some credentials information are missing: HE_DOMAIN_NAME,HE_USERNAME,HE_PASSWORD")
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
