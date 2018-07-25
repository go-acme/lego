package iij

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDomain string

/*
  Some environment variables needed to run this test
  		AccessKey:     os.Getenv("IIJAPI_ACCESS_KEY"),
		SecretKey:     os.Getenv("IIJAPI_SECRET_KEY"),
		DoServiceCode: os.Getenv("DOSERVICECODE"),
		TestDomain:    os.Getenv("IIJAPI_TESTDOMAIN"),
*/
func init() {
	testDomain = os.Getenv("IIJAPI_TESTDOMAIN")
}

func TestNewDNSProvider(t *testing.T) {
	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestDNSProvider_Present(t *testing.T) {
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	if !provider.ValidateDNSProvider() || testDomain == "" {
		t.Skip("skipping live test")
	}

	err = provider.Present(testDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	if !provider.ValidateDNSProvider() || testDomain == "" {
		t.Skip("skipping live test")
	}

	err = provider.CleanUp(testDomain, "", "123d==")
	assert.NoError(t, err)
}
