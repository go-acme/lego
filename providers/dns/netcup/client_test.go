package netcup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
)

func TestLiveClientAuth(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))

	for i := 1; i < 4; i++ {
		i := i
		t.Run("Test"+string(i), func(t *testing.T) {
			t.Parallel()

			sessionID, err := client.Login()
			require.NoError(t, err)

			err = client.Logout(sessionID)
			require.NoError(t, err)
		})
	}

}

func TestLiveClientGetDnsRecords(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))

	sessionID, err := client.Login()
	require.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(envTest.GetDomain(), "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	require.NoError(t, err, "error finding DNSZone")

	zone = acme.UnFqdn(zone)

	// TestMethod
	_, err = client.GetDNSRecords(zone, sessionID)
	require.NoError(t, err)

	// Tear down
	err = client.Logout(sessionID)
	require.NoError(t, err)
}

func TestLiveClientUpdateDnsRecord(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))

	sessionID, err := client.Login()
	require.NoError(t, err)

	fqdn, _, _ := acme.DNS01Record(envTest.GetDomain(), "123d==")

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	require.NoError(t, err, fmt.Errorf("error finding DNSZone, %v", err))

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	record := CreateTxtRecord(hostname, "asdf5678", 120)

	// test
	zone = acme.UnFqdn(zone)

	err = client.UpdateDNSRecord(sessionID, zone, record)
	require.NoError(t, err)

	records, err := client.GetDNSRecords(zone, sessionID)
	require.NoError(t, err)

	recordIdx, err := GetDNSRecordIdx(records, record)
	require.NoError(t, err)

	assert.Equal(t, record.Hostname, records[recordIdx].Hostname)
	assert.Equal(t, record.RecordType, records[recordIdx].RecordType)
	assert.Equal(t, record.Destination, records[recordIdx].Destination)
	assert.Equal(t, record.DeleteRecord, records[recordIdx].DeleteRecord)

	records[recordIdx].DeleteRecord = true

	// Tear down
	err = client.UpdateDNSRecord(sessionID, envTest.GetDomain(), records[recordIdx])
	require.NoError(t, err, "Did not remove record! Please do so yourself.")

	err = client.Logout(sessionID)
	require.NoError(t, err)
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
