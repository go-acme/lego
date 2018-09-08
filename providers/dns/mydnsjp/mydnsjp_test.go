package mydnsjp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	mydnsjpLiveTest bool
	mydnsjpMasterid string
	mydnsjpPassword string
	mydnsjpDomain   string
)

func init() {
	mydnsjpMasterid = os.Getenv("MYDNSJP_MASTERID")
	mydnsjpPassword = os.Getenv("MYDNSJP_PASSWORD")
	mydnsjpDomain = os.Getenv("MYDNSJP_DOMAIN")
	if len(mydnsjpMasterid) > 0 && len(mydnsjpPassword) > 0 && len(mydnsjpDomain) > 0 {
		mydnsjpLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("MYDNSJP_MASTERID", mydnsjpMasterid)
	os.Setenv("MYDNSJP_PASSWORD", mydnsjpPassword)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("MYDNSJP_MASTERID", "")
	os.Setenv("MYDNSJP_PASSWORD", "")
	defer restoreEnv()

	_, err := NewDNSProviderCredentials("123", "123")

	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("MYDNSJP_MASTERID", "test@example.com")
	os.Setenv("MYDNSJP_PASSWORD", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("MYDNSJP_MASTERID", "")
	os.Setenv("MYDNSJP_PASSWORD", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "MyDNS.jp: some credentials information are missing: MYDNSJP_MASTERID,MYDNSJP_PASSWORD")
}

func TestNewDNSProviderMissingCredErrSingle(t *testing.T) {
	defer restoreEnv()
	os.Setenv("MYDNSJP_MASTERID", "awesome@possum.com")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "MyDNS.jp: some credentials information are missing: MYDNSJP_PASSWORD")
}

func TestMyDNSJPPresent(t *testing.T) {
	if !mydnsjpLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(mydnsjpMasterid, mydnsjpPassword)
	assert.NoError(t, err)

	err = provider.Present(mydnsjpDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestMyDNSJPCleanUp(t *testing.T) {
	if !mydnsjpLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	provider, err := NewDNSProviderCredentials(mydnsjpMasterid, mydnsjpPassword)
	assert.NoError(t, err)

	err = provider.CleanUp(mydnsjpDomain, "", "123d==")
	assert.NoError(t, err)
}
