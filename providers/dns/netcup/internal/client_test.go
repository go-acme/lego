package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("a", "b", "c")
	require.NoError(t, err)

	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, mux
}

func TestGetDNSRecordIdx(t *testing.T) {
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

func TestClient_GetDNSRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if string(bytes.TrimSpace(raw)) != `{"action":"infoDnsRecords","param":{"domainname":"example.com","customernumber":"a","apikey":"b","apisessionid":""}}` {
			http.Error(rw, fmt.Sprintf("invalid request body: %s", string(raw)), http.StatusBadRequest)
			return
		}

		response := `
			{
			  "serverrequestid":"srv-request-id",
			  "clientrequestid":"",
			  "action":"infoDnsRecords",
			  "status":"success",
			  "statuscode":2000,
			  "shortmessage":"Login successful",
			  "longmessage":"Session has been created successful.",
			  "responsedata":{
			    "apisessionid":"api-session-id",
			    "dnsrecords":[
			      {
			        "id":"1",
			        "hostname":"example.com",
			        "type":"TXT",
			        "priority":"1",
			        "destination":"bGVnbzE=",
			        "state":"yes",
			        "ttl":300
			      },
			      {
			        "id":"2",
			        "hostname":"example2.com",
			        "type":"TXT",
			        "priority":"1",
			        "destination":"bGVnbw==",
			        "state":"yes",
			        "ttl":300
			      }
			    ]
			  }
			}`
		_, err = rw.Write([]byte(response))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	expected := []DNSRecord{{
		ID:           1,
		Hostname:     "example.com",
		RecordType:   "TXT",
		Priority:     "1",
		Destination:  "bGVnbzE=",
		DeleteRecord: false,
		State:        "yes",
	}, {
		ID:           2,
		Hostname:     "example2.com",
		RecordType:   "TXT",
		Priority:     "1",
		Destination:  "bGVnbw==",
		DeleteRecord: false,
		State:        "yes",
	}}

	records, err := client.GetDNSRecords(context.Background(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, expected, records)
}

func TestClient_GetDNSRecords_errors(t *testing.T) {
	testCases := []struct {
		desc    string
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			desc: "HTTP error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				http.Error(rw, "error message", http.StatusInternalServerError)
			},
		},
		{
			desc: "API error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				response := `
					{
						"serverrequestid":"YxTr4EzdbJ101T211zR4yzUEMVE",
						"clientrequestid":"",
						"action":"infoDnsRecords",
						"status":"error",
						"statuscode":4013,
						"shortmessage":"Validation Error.",
						"longmessage":"Message is empty.",
						"responsedata":""
					}`
				_, err := rw.Write([]byte(response))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			},
		},
		{
			desc: "responsedata marshaling error",
			handler: func(rw http.ResponseWriter, req *http.Request) {
				raw, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}

				if string(raw) != `{"action":"infoDnsRecords","param":{"domainname":"example.com","customernumber":"a","apikey":"b","apisessionid":"api-session-id"}}` {
					http.Error(rw, fmt.Sprintf("invalid request body: %s", string(raw)), http.StatusBadRequest)
					return
				}

				response := `
			{
			  "serverrequestid":"srv-request-id",
			  "clientrequestid":"",
			  "action":"infoDnsRecords",
			  "status":"success",
			  "statuscode":2000,
			  "shortmessage":"Login successful",
			  "longmessage":"Session has been created successful.",
			  "responsedata":""
			}`
				_, err = rw.Write([]byte(response))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client, mux := setupTest(t)

			mux.HandleFunc("/", test.handler)

			records, err := client.GetDNSRecords(context.Background(), "example.com")
			require.Error(t, err)
			assert.Empty(t, records)
		})
	}
}

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

	ctx, err := client.CreateSessionContext(context.Background())
	require.NoError(t, err)

	info := dns01.GetChallengeInfo(envTest.GetDomain(), "123d==")

	zone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	require.NoError(t, err, "error finding DNSZone")

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

	ctx, err := client.CreateSessionContext(context.Background())
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
	require.NoError(t, err, "Did not remove record! Please do so yourself.")

	err = client.Logout(ctx)
	require.NoError(t, err)
}
