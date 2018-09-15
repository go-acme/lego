package cloudflare

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	cflareLiveTest bool
	cflareEmail    string
	cflareAPIKey   string
	cflareDomain   string
)

func init() {
	cflareEmail = os.Getenv("CLOUDFLARE_EMAIL")
	cflareAPIKey = os.Getenv("CLOUDFLARE_API_KEY")
	cflareDomain = os.Getenv("CLOUDFLARE_DOMAIN")
	if len(cflareEmail) > 0 && len(cflareAPIKey) > 0 && len(cflareDomain) > 0 {
		cflareLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("CLOUDFLARE_EMAIL", cflareEmail)
	os.Setenv("CLOUDFLARE_API_KEY", cflareAPIKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("CLOUDFLARE_EMAIL", "")
	os.Setenv("CLOUDFLARE_API_KEY", "")
	defer restoreEnv()

	config := NewDefaultConfig()
	config.AuthEmail = "123"
	config.AuthKey = "123"

	_, err := NewDNSProviderConfig(config)

	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDFLARE_EMAIL", "test@example.com")
	os.Setenv("CLOUDFLARE_API_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDFLARE_EMAIL", "")
	os.Setenv("CLOUDFLARE_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL,CLOUDFLARE_API_KEY")
}

func TestNewDNSProviderMissingCredErrSingle(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDFLARE_EMAIL", "awesome@possum.com")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "cloudflare: some credentials information are missing: CLOUDFLARE_API_KEY")
}

func TestCloudFlarePresent(t *testing.T) {
	if !cflareLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.AuthEmail = cflareEmail
	config.AuthKey = cflareAPIKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(cflareDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestCloudFlareCleanUp(t *testing.T) {
	if !cflareLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	config := NewDefaultConfig()
	config.AuthEmail = cflareEmail
	config.AuthKey = cflareAPIKey

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(cflareDomain, "", "123d==")
	assert.NoError(t, err)
}
