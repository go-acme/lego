package acme

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dnsimpleEmail  string
	dnsimpleAPIKey string
)

func init() {
	dnsimpleEmail = os.Getenv("DNSIMPLE_EMAIL")
	dnsimpleAPIKey = os.Getenv("DNSIMPLE_API_KEY")
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
