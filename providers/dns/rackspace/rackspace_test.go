package rackspace

import (
	"os"
	"testing"
	//"fmt"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	rackspaceLiveTest bool
	rackspaceUser    string
	rackspaceAPIKey   string
	rackspaceDomain   string
)

func init() {
	rackspaceUser = os.Getenv("RACKSPACE_USER")
	rackspaceAPIKey = os.Getenv("RACKSPACE_API_KEY")
	rackspaceDomain = os.Getenv("RACKSPACE_DOMAIN")
	if len(rackspaceUser) > 0 && len(rackspaceAPIKey) > 0 && len(rackspaceDomain) > 0 {
		rackspaceLiveTest = true
	}
}

func restoreRackspaceEnv() {
	os.Setenv("RACKSPACE_USER", rackspaceUser)
	os.Setenv("RACKSPACE_API_KEY", rackspaceAPIKey)
}

/*
func TestNewDNSProviderValidEnv(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)
	assert.Contains(t, provider.cloudDNSEndpoint, "https://dns.api.rackspacecloud.com/v1.0/", "The endpoint URL should contain the base")
	restoreRackspaceEnv()
}

func TestRackspaceGetHostedZoneID(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(rackspaceUser, rackspaceAPIKey)
	assert.NoError(t, err)

	zoneID, err := provider.getHostedZoneID("_test." + rackspaceDomain + ".")
	assert.NoError(t, err)
	assert.NotEmpty(t, zoneID)
}

func TestRackspaceFindTXTRecord(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

    fqdn := "_acme-challenge_test." + rackspaceDomain + "."

	provider, err := NewDNSProviderCredentials(rackspaceUser, rackspaceAPIKey)
	assert.NoError(t, err)

	zoneID, err := provider.getHostedZoneID(fqdn)
	assert.NoError(t, err)
	assert.NotEmpty(t, zoneID)

    record, err := provider.findTxtRecord(fqdn, zoneID)
	assert.NoError(t, err)
	assert.EqualValues(t, record, &RackspaceRecord{"_acme-challenge_test.www.hanley.it", "TXT", "Testing", 300, "TXT-993802"})
}
*/

/*
func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("RACKSPACE_USER", "")
	os.Setenv("RACKSPACE_API_KEY", "")
	_, err := NewDNSProviderCredentials("username", "123")
	assert.NoError(t, err)
	restoreRackspaceEnv()
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("RACKSPACE_USER", "username")
	os.Setenv("RACKSPACE_API_KEY", "123")
	_, err := NewDNSProvider()
	assert.NoError(t, err)
	restoreRackspaceEnv()
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("RACKSPACE_USER", "")
	os.Setenv("RACKSPACE_API_KEY", "")
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "Rackspace credentials missing")
	restoreRackspaceEnv()
}
*/

/*
func TestRackspaceFindTXTRecord(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

    fqdn := "_acme-challenge_test." + rackspaceDomain + "."

	provider, err := NewDNSProviderCredentials(rackspaceUser, rackspaceAPIKey)
	assert.NoError(t, err)

	zoneID, err := provider.getHostedZoneID(fqdn)
	assert.NoError(t, err)
	assert.NotEmpty(t, zoneID)

    record, err := provider.findTxtRecord(fqdn, zoneID)
	assert.NoError(t, err)
	assert.EqualValues(t, record, &RackspaceRecord{"_acme-challenge_test.www.hanley.it", "TXT", "Testing", 300, "TXT-993802"})

    _, err = provider.makeRequest("DELETE", fmt.Sprintf("/domains/%d/records?id=%s", zoneID, record.ID), nil)
 	assert.NoError(t, err)
}
*/


func TestRackspacePresent(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(rackspaceUser, rackspaceAPIKey)
	assert.NoError(t, err)

	err = provider.Present(rackspaceDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestRackspaceCleanUp(t *testing.T) {
	if !rackspaceLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 15)

	provider, err := NewDNSProviderCredentials(rackspaceUser, rackspaceAPIKey)
	assert.NoError(t, err)

	err = provider.CleanUp(rackspaceDomain, "", "123d==")
	assert.NoError(t, err)
}

