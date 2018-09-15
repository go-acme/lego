package sakuracloud

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
)

var (
	sakuracloudLiveTest     bool
	sakuracloudAccessToken  string
	sakuracloudAccessSecret string
	sakuracloudDomain       string
)

func init() {
	sakuracloudAccessToken = os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	sakuracloudAccessSecret = os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	sakuracloudDomain = os.Getenv("SAKURACLOUD_DOMAIN")

	if len(sakuracloudAccessToken) > 0 && len(sakuracloudAccessSecret) > 0 && len(sakuracloudDomain) > 0 {
		sakuracloudLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", sakuracloudAccessToken)
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", sakuracloudAccessSecret)
}

//
// NewDNSProvider
//

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()

	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "123")
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "456")
	provider, err := NewDNSProvider()

	assert.NotNil(t, provider)
	assert.Equal(t, acme.UserAgent, provider.client.UserAgent)
	assert.NoError(t, err)
}

func TestNewDNSProviderInvalidWithMissingAccessToken(t *testing.T) {
	defer restoreEnv()

	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "")
	provider, err := NewDNSProvider()

	assert.Nil(t, provider)
	assert.EqualError(t, err, "sakuracloud: some credentials information are missing: SAKURACLOUD_ACCESS_TOKEN,SAKURACLOUD_ACCESS_TOKEN_SECRET")
}

//
// NewDNSProviderCredentials
//

func TestNewDNSProviderCredentialsValid(t *testing.T) {
	config := NewDefaultConfig()
	config.Token = "123"
	config.Secret = "456"

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	assert.NotNil(t, provider)
	assert.Equal(t, acme.UserAgent, provider.client.UserAgent)
}

func TestNewDNSProviderCredentialsInvalidWithMissingAccessToken(t *testing.T) {
	config := NewDefaultConfig()

	provider, err := NewDNSProviderConfig(config)
	assert.Nil(t, provider)
	assert.EqualError(t, err, "sakuracloud: AccessToken is missing")
}

//
// Present
//

func TestLiveSakuraCloudPresent(t *testing.T) {
	if !sakuracloudLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.Token = sakuracloudAccessToken
	config.Secret = sakuracloudAccessSecret

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(sakuracloudDomain, "", "123d==")
	assert.NoError(t, err)
}

//
// Cleanup
//

func TestLiveSakuraCloudCleanUp(t *testing.T) {
	if !sakuracloudLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.Token = sakuracloudAccessToken
	config.Secret = sakuracloudAccessSecret

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(sakuracloudDomain, "", "123d==")
	assert.NoError(t, err)
}
