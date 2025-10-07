package internal

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"NETCUP_CUSTOMER_NUMBER",
	"NETCUP_API_KEY",
	"NETCUP_API_PASSWORD").
	WithDomain("NETCUP_DOMAIN")

func TestClient_GetDNSRecords_Live(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client, err := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))
	require.NoError(t, err)

	ctx, err := client.CreateSessionContext(t.Context())
	require.NoError(t, err)

	info := dns01.GetChallengeInfo(envTest.GetDomain(), "123d==")

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	require.NoError(t, err)

	zone = dns01.UnFqdn(zone)

	// TestMethod
	_, err = client.GetDNSRecords(ctx, zone)
	require.NoError(t, err)

	// Tear down
	err = client.Logout(ctx)
	require.NoError(t, err)
}

func TestClient_UpdateDNSRecord_Live(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client, err := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))
	require.NoError(t, err)

	ctx, err := client.CreateSessionContext(t.Context())
	require.NoError(t, err)

	info := dns01.GetChallengeInfo(envTest.GetDomain(), "123d==")

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	require.NotErrorIs(t, err, fmt.Errorf("error finding DNSZone, %w", err))

	hostname := strings.Replace(info.EffectiveFQDN, "."+zone, "", 1)

	record := DNSRecord{
		Hostname:     hostname,
		RecordType:   "TXT",
		Destination:  "asdf5678",
		DeleteRecord: false,
	}

	// test
	zone = dns01.UnFqdn(zone)

	err = client.UpdateDNSRecord(ctx, zone, []DNSRecord{record})
	require.NoError(t, err)

	records, err := client.GetDNSRecords(ctx, zone)
	require.NoError(t, err)

	recordIdx, err := GetDNSRecordIdx(records, record)
	require.NoError(t, err)

	assert.Equal(t, record.Hostname, records[recordIdx].Hostname)
	assert.Equal(t, record.RecordType, records[recordIdx].RecordType)
	assert.Equal(t, record.Destination, records[recordIdx].Destination)
	assert.Equal(t, record.DeleteRecord, records[recordIdx].DeleteRecord)

	records[recordIdx].DeleteRecord = true

	// Tear down
	err = client.UpdateDNSRecord(ctx, envTest.GetDomain(), []DNSRecord{records[recordIdx]})
	require.NoError(t, err)

	err = client.Logout(ctx)
	require.NoError(t, err)
}

func TestLiveClientAuth(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client, err := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))
	require.NoError(t, err)

	for i := range 4 {
		t.Run("Test_"+strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			ctx, err := client.CreateSessionContext(t.Context())
			require.NoError(t, err)

			err = client.Logout(ctx)
			require.NoError(t, err)
		})
	}
}
