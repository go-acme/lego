package duckdns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	duckdnsLiveTest bool
	duckdnsToken    string
	duckdnsDomain   string
)

func init() {
	duckdnsToken = os.Getenv("DUCKDNS_TOKEN")
	duckdnsDomain = os.Getenv("DUCKDNS_DOMAIN")
	if len(duckdnsDomain) > 0 && len(duckdnsDomain) > 0 {
		duckdnsLiveTest = true
	}
}

func restoreDuckdnsEnv() {
	os.Setenv("DUCKDNS_TOKEN", duckdnsToken)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("DUCKDNS_TOKEN", "123")
	_, err := NewDNSProvider()
	assert.NoError(t, err)
	restoreDuckdnsEnv()
}
func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("DUCKDNS_TOKEN", "")
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "environment variable DUCKDNS_TOKEN not set")
	restoreDuckdnsEnv()
}
func TestLiveDuckdnsPresent(t *testing.T) {
	if !duckdnsLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(duckdnsDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveDuckdnsCleanUp(t *testing.T) {
	if !duckdnsLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 10)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(duckdnsDomain, "", "123d==")
	assert.NoError(t, err)
}
