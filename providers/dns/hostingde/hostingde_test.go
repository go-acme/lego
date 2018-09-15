package hostingde

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	hostingdeLiveTest bool
	hostingdeAPIKey   string
	hostingdeZone     string
	hostingdeDomain   string
)

func init() {
	hostingdeAPIKey = os.Getenv("HOSTINGDE_API_KEY")
	hostingdeZone = os.Getenv("HOSTINGDE_ZONE_NAME")
	hostingdeDomain = os.Getenv("HOSTINGDE_DOMAIN")
	if len(hostingdeZone) > 0 && len(hostingdeAPIKey) > 0 && len(hostingdeDomain) > 0 {
		hostingdeLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("HOSTINGDE_ZONE_NAME", hostingdeZone)
	os.Setenv("HOSTINGDE_API_KEY", hostingdeAPIKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("HOSTINGDE_ZONE_NAME", "")
	os.Setenv("HOSTINGDE_API_KEY", "")
	defer restoreEnv()

	config := NewDefaultConfig()
	config.APIKey = "123"
	config.ZoneName = "example.com"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("HOSTINGDE_ZONE_NAME", "example.com")
	os.Setenv("HOSTINGDE_API_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("HOSTINGDE_ZONE_NAME", "")
	os.Setenv("HOSTINGDE_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "hostingde: some credentials information are missing: HOSTINGDE_API_KEY,HOSTINGDE_ZONE_NAME")
}

func TestNewDNSProviderMissingCredErrSingle(t *testing.T) {
	defer restoreEnv()
	os.Setenv("HOSTINGDE_ZONE_NAME", "example.com")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "hostingde: some credentials information are missing: HOSTINGDE_API_KEY")
}

func TestHostingdePresent(t *testing.T) {
	if !hostingdeLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.APIKey = hostingdeZone
	config.ZoneName = hostingdeAPIKey

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(hostingdeDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestHostingdeCleanUp(t *testing.T) {
	if !hostingdeLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	config := NewDefaultConfig()
	config.APIKey = hostingdeZone
	config.ZoneName = hostingdeAPIKey

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(hostingdeDomain, "", "123d==")
	assert.NoError(t, err)
}
