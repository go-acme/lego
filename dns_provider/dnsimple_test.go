package dns_provider

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	dnsimpleLiveTest bool
	dnsimpleEmail    string
	dnsimpleAPIKey   string
	dnsimpleDomain   string
)

func init() {
	dnsimpleEmail = os.Getenv("DNSIMPLE_EMAIL")
	dnsimpleAPIKey = os.Getenv("DNSIMPLE_API_KEY")
	dnsimpleDomain = os.Getenv("DNSIMPLE_DOMAIN")
	if len(dnsimpleEmail) > 0 && len(dnsimpleAPIKey) > 0 && len(dnsimpleDomain) > 0 {
		dnsimpleLiveTest = true
	}
}

func restoreDNSimpleEnv() {
	os.Setenv("DNSIMPLE_EMAIL", dnsimpleEmail)
	os.Setenv("DNSIMPLE_API_KEY", dnsimpleAPIKey)
}

func TestNewDNSProviderDNSimpleValid(t *testing.T) {
	os.Setenv("DNSIMPLE_EMAIL", "")
	os.Setenv("DNSIMPLE_API_KEY", "")
	_, err := NewDNSProviderDNSimple("example@example.com", "123")
	assert.NoError(t, err)
	restoreDNSimpleEnv()
}
func TestNewDNSProviderDNSimpleValidEnv(t *testing.T) {
	os.Setenv("DNSIMPLE_EMAIL", "example@example.com")
	os.Setenv("DNSIMPLE_API_KEY", "123")
	_, err := NewDNSProviderDNSimple("", "")
	assert.NoError(t, err)
	restoreDNSimpleEnv()
}

func TestNewDNSProviderDNSimpleMissingCredErr(t *testing.T) {
	os.Setenv("DNSIMPLE_EMAIL", "")
	os.Setenv("DNSIMPLE_API_KEY", "")
	_, err := NewDNSProviderDNSimple("", "")
	assert.EqualError(t, err, "DNSimple credentials missing")
	restoreDNSimpleEnv()
}

func TestLiveDNSimplePresent(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderDNSimple(dnsimpleEmail, dnsimpleAPIKey)
	assert.NoError(t, err)

	err = provider.Present(dnsimpleDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveDNSimpleCleanUp(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProviderDNSimple(cflareEmail, cflareAPIKey)
	assert.NoError(t, err)

	err = provider.CleanUp(dnsimpleDomain, "", "123d==")
	assert.NoError(t, err)
}
