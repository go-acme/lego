package namedotcom

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	namedotcomLiveTest bool
	namedotcomUsername string
	namedotcomAPIToken string
	namedotcomDomain   string
	namedotcomServer   string
)

func init() {
	namedotcomUsername = os.Getenv("NAMEDOTCOM_USERNAME")
	namedotcomAPIToken = os.Getenv("NAMEDOTCOM_API_TOKEN")
	namedotcomDomain = os.Getenv("NAMEDOTCOM_DOMAIN")
	namedotcomServer = os.Getenv("NAMEDOTCOM_SERVER")

	if len(namedotcomAPIToken) > 0 && len(namedotcomUsername) > 0 && len(namedotcomDomain) > 0 {
		namedotcomLiveTest = true
	}
}

func TestLiveNamedotcomPresent(t *testing.T) {
	if !namedotcomLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.Username = namedotcomUsername
	config.APIToken = namedotcomAPIToken
	config.Server = namedotcomServer

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(namedotcomDomain, "", "123d==")
	assert.NoError(t, err)
}

//
// Cleanup
//

func TestLiveNamedotcomCleanUp(t *testing.T) {
	if !namedotcomLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.Username = namedotcomUsername
	config.APIToken = namedotcomAPIToken
	config.Server = namedotcomServer

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(namedotcomDomain, "", "123d==")
	assert.NoError(t, err)
}
