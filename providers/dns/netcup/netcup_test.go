package netcup

import (
	"fmt"
	"os"
	"strings"
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

func TestPresentAndCleanup(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(testDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	assert.NoError(t, err, fmt.Errorf("error finding DNSZone, %v", err))

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
			assert.NoError(t, err)
			if err != nil {
				t.Log("Did not clean up! Please remove record yourself.")
			}
		})
	}
}

func TestAuth(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	//Setup
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	for i := 1; i < 4; i++ {
		i := i
		t.Run("Test"+string(i), func(t *testing.T) {
			t.Parallel()
			skey, err := provider.login()
			assert.NoError(t, err)

			err = provider.logout(skey)
			assert.NoError(t, err)
		})
	}

}

func TestGetDnsRecords(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	//Setup
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	sid, err := p.login()
	assert.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(testDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	assert.NoError(t, err, fmt.Errorf("error finding DNSZone, %v", err))

	zone = acme.UnFqdn(zone)

	//TestMethod
	_, err = p.getDNSRecords(zone, sid)
	assert.NoError(t, err)

	//Tear down
	err = p.logout(sid)
	assert.NoError(t, err)
}

func TestUpdateDnsRecord(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	//Setup
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	sid, err := p.login()
	assert.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(testDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	assert.NoError(t, err, fmt.Errorf("error finding DNSZone, %v", err))

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	zone = acme.UnFqdn(zone)

	record := createTxtRecord(hostname, "asdf5678")

	//test
	err = p.updateDNSRecord(sid, zone, record)
	assert.NoError(t, err)

	records, err := p.getDNSRecords(zone, sid)
	assert.NoError(t, err)

	recordIdx, err := p.getDNSRecordIdx(records, record)
	assert.NoError(t, err)

	assert.Equal(t, record.Hostname, records[recordIdx].Hostname)
	assert.Equal(t, record.Recordtype, records[recordIdx].Recordtype)
	assert.Equal(t, record.Destination, records[recordIdx].Destination)
	assert.Equal(t, record.Deleterecord, records[recordIdx].Deleterecord)

	records[recordIdx].Deleterecord = true

	//Tear down
	err = p.updateDNSRecord(sid, testDomain, records[recordIdx])
	assert.NoError(t, err)
	if err != nil {
		t.Logf("Did not remove record! Please do so yourself.")
	}

	err = p.logout(sid)
	assert.NoError(t, err)
}

func TestGetDnsRecordIdx(t *testing.T) {
	//Setup
	p, err := NewDNSProvider()
	assert.NoError(t, err)

	records := []DNSRecord{
		{
			ID:           12345,
			Hostname:     "asdf",
			Recordtype:   "TXT",
			Priority:     "0",
			Destination:  "randomtext",
			Deleterecord: false,
			State:        "yes",
		},
		{
			ID:           23456,
			Hostname:     "@",
			Recordtype:   "A",
			Priority:     "0",
			Destination:  "127.0.0.1",
			Deleterecord: false,
			State:        "yes",
		},
		{
			ID:           34567,
			Hostname:     "dfgh",
			Recordtype:   "CNAME",
			Priority:     "0",
			Destination:  "example.com",
			Deleterecord: false,
			State:        "yes",
		},
		{
			ID:           45678,
			Hostname:     "fghj",
			Recordtype:   "MX",
			Priority:     "10",
			Destination:  "mail.example.com",
			Deleterecord: false,
			State:        "yes",
		},
	}

	record := DNSRecord{
		ID:           12345,
		Hostname:     "asdf",
		Recordtype:   "TXT",
		Priority:     "0",
		Destination:  "randomtext",
		Deleterecord: false,
		State:        "yes",
	}

	//TestMethod
	idx, err := p.getDNSRecordIdx(records, record)
	assert.NoError(t, err)
	assert.Equal(t, record, records[idx])

	record.Destination = "wrong"
	idx, err = p.getDNSRecordIdx(records, record)
	assert.Error(t, err)
	assert.Equal(t, -1, idx)

	record.Destination = "randomtext"
	record.Recordtype = "CNAME"
	idx, err = p.getDNSRecordIdx(records, record)
	assert.Error(t, err)
	assert.Equal(t, -1, idx)
}
