package gcloud

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

var (
	gcloudLiveTest bool
	gcloudProject  string
	gcloudDomain   string
)

func init() {
	gcloudProject = os.Getenv("GCE_PROJECT")
	gcloudDomain = os.Getenv("GCE_DOMAIN")
	_, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err == nil && len(gcloudProject) > 0 && len(gcloudDomain) > 0 {
		gcloudLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("GCE_PROJECT", gcloudProject)
}

func TestNewDNSProviderValid(t *testing.T) {
	if !gcloudLiveTest {
		t.Skip("skipping live test (requires credentials)")
	}

	defer restoreEnv()
	os.Setenv("GCE_PROJECT", "")

	_, err := NewDNSProviderCredentials("my-project")
	require.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	if !gcloudLiveTest {
		t.Skip("skipping live test (requires credentials)")
	}

	defer restoreEnv()
	os.Setenv("GCE_PROJECT", "my-project")

	_, err := NewDNSProvider()
	require.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("GCE_PROJECT", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "googlecloud: project name missing")
}

func TestLiveGoogleCloudPresent(t *testing.T) {
	if !gcloudLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(gcloudProject)
	require.NoError(t, err)

	err = provider.Present(gcloudDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveGoogleCloudPresentMultiple(t *testing.T) {
	if !gcloudLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(gcloudProject)
	require.NoError(t, err)

	// Check that we're able to create multiple entries
	err = provider.Present(gcloudDomain, "1", "123d==")
	require.NoError(t, err)
	err = provider.Present(gcloudDomain, "2", "123d==")
	require.NoError(t, err)
}

func TestLiveGoogleCloudCleanUp(t *testing.T) {
	if !gcloudLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProviderCredentials(gcloudProject)
	require.NoError(t, err)

	err = provider.CleanUp(gcloudDomain, "", "123d==")
	require.NoError(t, err)
}
