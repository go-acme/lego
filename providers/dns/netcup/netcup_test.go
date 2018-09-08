package netcup

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xenolf/lego/acme"
)

var (
	testLive           bool
	testCustomerNumber string
	testAPIKey         string
	testAPIPassword    string
	testDomain         string
)

func init() {
	testCustomerNumber = os.Getenv("NETCUP_CUSTOMER_NUMBER")
	testAPIKey = os.Getenv("NETCUP_API_KEY")
	testAPIPassword = os.Getenv("NETCUP_API_PASSWORD")
	testDomain = os.Getenv("NETCUP_DOMAIN")

	if len(testCustomerNumber) > 0 && len(testAPIKey) > 0 && len(testAPIPassword) > 0 && len(testDomain) > 0 {
		testLive = true
	}
}

func TestDNSProviderPresentAndCleanup(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(testDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	assert.NoError(t, err, "error finding DNSZone")

	zone = acme.UnFqdn(zone)

	testCases := []string{
		zone,
		"sub." + zone,
		"*." + zone,
		"*.sub." + zone,
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("domain(%s)", tc), func(t *testing.T) {
			err = p.Present(tc, "987d", "123d==")
			assert.NoError(t, err)

			err = p.CleanUp(tc, "987d", "123d==")
			assert.NoError(t, err, "Did not clean up! Please remove record yourself.")
		})
	}
}
