package netcup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xenolf/lego/acme"
)

func TestClientAuth(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	// Setup
	client := NewClient(testCustomerNumber, testAPIKey, testAPIPassword)

	for i := 1; i < 4; i++ {
		i := i
		t.Run("Test"+string(i), func(t *testing.T) {
			t.Parallel()

			sessionID, err := client.Login()
			assert.NoError(t, err)

			err = client.Logout(sessionID)
			assert.NoError(t, err)
		})
	}

}

func TestClientGetDnsRecords(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	client := NewClient(testCustomerNumber, testAPIKey, testAPIPassword)

	// Setup
	sessionID, err := client.Login()
	assert.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(testDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	assert.NoError(t, err, "error finding DNSZone")

	zone = acme.UnFqdn(zone)

	// TestMethod
	_, err = client.GetDNSRecords(zone, sessionID)
	assert.NoError(t, err)

	// Tear down
	err = client.Logout(sessionID)
	assert.NoError(t, err)
}

func TestClientUpdateDnsRecord(t *testing.T) {
	if !testLive {
		t.Skip("skipping live test")
	}

	// Setup
	client := NewClient(testCustomerNumber, testAPIKey, testAPIPassword)

	sessionID, err := client.Login()
	assert.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(testDomain, "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	assert.NoError(t, err, fmt.Errorf("error finding DNSZone, %v", err))

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	record := CreateTxtRecord(hostname, "asdf5678", 120)

	// test
	zone = acme.UnFqdn(zone)

	err = client.UpdateDNSRecord(sessionID, zone, record)
	assert.NoError(t, err)

	records, err := client.GetDNSRecords(zone, sessionID)
	assert.NoError(t, err)

	recordIdx, err := GetDNSRecordIdx(records, record)
	assert.NoError(t, err)

	assert.Equal(t, record.Hostname, records[recordIdx].Hostname)
	assert.Equal(t, record.RecordType, records[recordIdx].RecordType)
	assert.Equal(t, record.Destination, records[recordIdx].Destination)
	assert.Equal(t, record.DeleteRecord, records[recordIdx].DeleteRecord)

	records[recordIdx].DeleteRecord = true

	// Tear down
	err = client.UpdateDNSRecord(sessionID, testDomain, records[recordIdx])
	assert.NoError(t, err, "Did not remove record! Please do so yourself.")

	err = client.Logout(sessionID)
	assert.NoError(t, err)
}

func TestClientGetDnsRecordIdx(t *testing.T) {
	records := []DNSRecord{
		{
			ID:           12345,
			Hostname:     "asdf",
			RecordType:   "TXT",
			Priority:     "0",
			Destination:  "randomtext",
			DeleteRecord: false,
			State:        "yes",
		},
		{
			ID:           23456,
			Hostname:     "@",
			RecordType:   "A",
			Priority:     "0",
			Destination:  "127.0.0.1",
			DeleteRecord: false,
			State:        "yes",
		},
		{
			ID:           34567,
			Hostname:     "dfgh",
			RecordType:   "CNAME",
			Priority:     "0",
			Destination:  "example.com",
			DeleteRecord: false,
			State:        "yes",
		},
		{
			ID:           45678,
			Hostname:     "fghj",
			RecordType:   "MX",
			Priority:     "10",
			Destination:  "mail.example.com",
			DeleteRecord: false,
			State:        "yes",
		},
	}

	testCases := []struct {
		desc        string
		record      DNSRecord
		expectError bool
	}{
		{
			desc: "simple",
			record: DNSRecord{
				ID:           12345,
				Hostname:     "asdf",
				RecordType:   "TXT",
				Priority:     "0",
				Destination:  "randomtext",
				DeleteRecord: false,
				State:        "yes",
			},
		},
		{
			desc: "wrong Destination",
			record: DNSRecord{
				ID:           12345,
				Hostname:     "asdf",
				RecordType:   "TXT",
				Priority:     "0",
				Destination:  "wrong",
				DeleteRecord: false,
				State:        "yes",
			},
			expectError: true,
		},
		{
			desc: "record type CNAME",
			record: DNSRecord{
				ID:           12345,
				Hostname:     "asdf",
				RecordType:   "CNAME",
				Priority:     "0",
				Destination:  "randomtext",
				DeleteRecord: false,
				State:        "yes",
			},
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			idx, err := GetDNSRecordIdx(records, test.record)
			if test.expectError {
				assert.Error(t, err)
				assert.Equal(t, -1, idx)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, records[idx], test.record)
			}
		})
	}
}
