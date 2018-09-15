package mydnsjp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mydnsjpLiveTest bool
	mydnsjpMasterID string
	mydnsjpPassword string
	mydnsjpDomain   string
)

func init() {
	mydnsjpMasterID = os.Getenv("MYDNSJP_MASTER_ID")
	mydnsjpPassword = os.Getenv("MYDNSJP_PASSWORD")
	mydnsjpDomain = os.Getenv("MYDNSJP_DOMAIN")
	if len(mydnsjpMasterID) > 0 && len(mydnsjpPassword) > 0 && len(mydnsjpDomain) > 0 {
		mydnsjpLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("MYDNSJP_MASTER_ID", mydnsjpMasterID)
	os.Setenv("MYDNSJP_PASSWORD", mydnsjpPassword)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("MYDNSJP_MASTER_ID", "")
	os.Setenv("MYDNSJP_PASSWORD", "")
	defer restoreEnv()

	config := NewDefaultConfig()
	config.MasterID = "123"
	config.Password = "123"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("MYDNSJP_MASTER_ID", "test@example.com")
	os.Setenv("MYDNSJP_PASSWORD", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("MYDNSJP_MASTER_ID", "")
	os.Setenv("MYDNSJP_PASSWORD", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "mydnsjp: some credentials information are missing: MYDNSJP_MASTER_ID,MYDNSJP_PASSWORD")
}

func TestNewDNSProviderMissingCredErrSingle(t *testing.T) {
	defer restoreEnv()
	os.Setenv("MYDNSJP_MASTER_ID", "awesome@possum.com")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "mydnsjp: some credentials information are missing: MYDNSJP_PASSWORD")
}

func TestMyDNSJPPresent(t *testing.T) {
	if !mydnsjpLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.MasterID = mydnsjpMasterID
	config.Password = mydnsjpPassword

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(mydnsjpDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestMyDNSJPCleanUp(t *testing.T) {
	if !mydnsjpLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(2 * time.Second)

	config := NewDefaultConfig()
	config.MasterID = mydnsjpMasterID
	config.Password = mydnsjpPassword

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(mydnsjpDomain, "", "123d==")
	assert.NoError(t, err)
}
