package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("a", "b", "c")
			if err != nil {
				return nil, err
			}

			client.baseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
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
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("get_dns_records.json"),
			servermock.CheckRequestJSONBodyFromFile("get_dns_records-request.json")).
		Build(t)

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

	records, err := client.GetDNSRecords(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, expected, records)
}

func TestClient_GetDNSRecords_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		handler  http.Handler
		expected string
	}{
		{
			desc:     "HTTP error",
			handler:  servermock.Noop().WithStatusCode(http.StatusInternalServerError),
			expected: `error when sending the request: unexpected status code: [status code: 500] body: `,
		},
		{
			desc:     "API error",
			handler:  servermock.ResponseFromFixture("get_dns_records_error.json"),
			expected: `error when sending the request: an error occurred during the action infoDnsRecords: [Status=error, StatusCode=4013, ShortMessage=Validation Error., LongMessage=Message is empty.]`,
		},
		{
			desc:     "responsedata marshaling error",
			handler:  servermock.ResponseFromFixture("get_dns_records_error_unmarshal.json"),
			expected: `error when sending the request: unable to unmarshal response: [status code: 200] body: "" error: json: cannot unmarshal string into Go value of type internal.InfoDNSRecordsResponse`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /", test.handler).
				Build(t)

			records, err := client.GetDNSRecords(t.Context(), "example.com")
			require.EqualError(t, err, test.expected)
			assert.Empty(t, records)
		})
	}
}
