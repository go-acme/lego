package cloudxns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	cxLiveTest  bool
	cxAPIKey    string
	cxSecretKey string
	cxDomain    string
)

func init() {
	cxAPIKey = os.Getenv("CLOUDXNS_API_KEY")
	cxSecretKey = os.Getenv("CLOUDXNS_SECRET_KEY")
	cxDomain = os.Getenv("CLOUDXNS_DOMAIN")
	if len(cxAPIKey) > 0 && len(cxSecretKey) > 0 && len(cxDomain) > 0 {
		cxLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("CLOUDXNS_API_KEY", cxAPIKey)
	os.Setenv("CLOUDXNS_SECRET_KEY", cxSecretKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDXNS_API_KEY", "")
	os.Setenv("CLOUDXNS_SECRET_KEY", "")

	_, err := NewDNSProviderCredentials("123", "123")
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDXNS_API_KEY", "123")
	os.Setenv("CLOUDXNS_SECRET_KEY", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("CLOUDXNS_API_KEY", "")
	os.Setenv("CLOUDXNS_SECRET_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "CloudXNS: some credentials information are missing: CLOUDXNS_API_KEY,CLOUDXNS_SECRET_KEY")
}

func TestCloudXNSPresent(t *testing.T) {
	if !cxLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(cxAPIKey, cxSecretKey)
	assert.NoError(t, err)

	err = provider.Present(cxDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestCloudXNSCleanUp(t *testing.T) {
	if !cxLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	provider, err := NewDNSProviderCredentials(cxAPIKey, cxSecretKey)
	assert.NoError(t, err)

	err = provider.CleanUp(cxDomain, "", "123d==")
	assert.NoError(t, err)
}
