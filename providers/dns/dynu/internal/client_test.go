package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(method string, pattern string, status int, file string) *Client {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		open, err := os.Open(file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient()
	client.BaseURL = server.URL

	return client
}

func TestGetRootDomain(t *testing.T) {
	type expected struct {
		domain *DNSHostname
		error  string
	}

	testCases := []struct {
		desc     string
		pattern  string
		status   int
		file     string
		expected expected
	}{
		{
			desc:    "success",
			pattern: "/dns/getroot/gh.ldez.freeddns.org",
			status:  http.StatusOK,
			file:    "./fixtures/get_root_domain.json",
			expected: expected{
				domain: &DNSHostname{
					APIException: &APIException{
						StatusCode: 200,
					},
					ID:         9007481,
					DomainName: "ldez.freeddns.org",
					Hostname:   "gh.ldez.freeddns.org",
					Node:       "gh",
				},
			},
		},
		{
			desc:    "invalid",
			pattern: "/dns/getroot/gh.ldez.freeddns.org",
			status:  http.StatusNotImplemented,
			file:    "./fixtures/get_root_domain_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(http.MethodGet, test.pattern, test.status, test.file)

			domain, err := client.GetRootDomain("gh.ldez.freeddns.org")

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, domain)
			assert.Equal(t, test.expected.domain, domain)
		})
	}
}

func TestGetRecords(t *testing.T) {
	type expected struct {
		records []DNSRecord
		error   string
	}

	testCases := []struct {
		desc     string
		pattern  string
		status   int
		file     string
		expected expected
	}{
		{
			desc:    "success",
			pattern: "/dns/record/_acme-challenge.ldez.freeddns.org",
			status:  http.StatusOK,
			file:    "./fixtures/get_records.json",
			expected: expected{
				records: []DNSRecord{{
					ID:         6041417,
					Type:       "TXT",
					DomainID:   9007481,
					DomainName: "ldez.freeddns.org",
					NodeName:   "_acme-challenge",
					Hostname:   "_acme-challenge.ldez.freeddns.org",
					State:      true,
					Content:    `_acme-challenge.ldez.freeddns.org. 300 IN TXT "txt_txt_txt_txt_txt_txt_txt"`,
					TextData:   "txt_txt_txt_txt_txt_txt_txt",
					TTL:        300,
				},
					{
						ID:         6041422,
						Type:       "TXT",
						DomainID:   9007481,
						DomainName: "ldez.freeddns.org",
						NodeName:   "_acme-challenge",
						Hostname:   "_acme-challenge.ldez.freeddns.org",
						State:      true,
						Content:    `_acme-challenge.ldez.freeddns.org. 300 IN TXT "txt_txt_txt_txt_txt_txt_txt_2"`,
						TextData:   "txt_txt_txt_txt_txt_txt_txt_2",
						TTL:        300,
					},
				},
			},
		},
		{
			desc:    "empty",
			pattern: "/dns/record/_acme-challenge.ldez.freeddns.org",
			status:  http.StatusOK,
			file:    "./fixtures/get_records_empty.json",
			expected: expected{
				records: []DNSRecord{},
			},
		},
		{
			desc:    "invalid",
			pattern: "/dns/record/_acme-challenge.ldez.freeddns.org",
			status:  http.StatusNotImplemented,
			file:    "./fixtures/get_records_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(http.MethodGet, test.pattern, test.status, test.file)

			records, err := client.GetRecords("_acme-challenge.ldez.freeddns.org", "TXT")

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, records)
			assert.Equal(t, test.expected.records, records)
		})
	}

	t.Skip("now")
	client := NewClient()

	// {"statusCode":200,"dnsRecords":[{"id":6041417,"domainId":9007481,"domainName":"ldez.freeddns.org","nodeName":"_acme-challenge","hostname":"_acme-challenge.ldez.freeddns.org","recordType":"TXT","ttl":300,"state":true,"content":"_acme-challenge.ldez.freeddns.org. 300 IN TXT \"txt_txt_txt_txt_txt_txt_txt\"","updatedOn":"2020-03-10T04:00:36.923","textData":"txt_txt_txt_txt_txt_txt_txt"},{"id":6041422,"domainId":9007481,"domainName":"ldez.freeddns.org","nodeName":"_acme-challenge","hostname":"_acme-challenge.ldez.freeddns.org","recordType":"TXT","ttl":300,"state":true,"content":"_acme-challenge.ldez.freeddns.org. 300 IN TXT \"txt_txt_txt_txt_txt_txt_txt_2\"","updatedOn":"2020-03-10T04:03:17.563","textData":"txt_txt_txt_txt_txt_txt_txt_2"}]}

	// {"statusCode":200,"dnsRecords":[]}

	// {"statusCode":501,"type":"Argument Exception","message":"Invalid."}

	record, err := client.GetRecords("challengeldez.freeddns.org", "TXT")
	// record, err := client.GetRecords("_acme-challenge.ldez.freeddns.org", "TXT")
	require.NoError(t, err)

	fmt.Println(record)
}

func TestAddNewRecord(t *testing.T) {
	type expected struct {
		error string
	}

	testCases := []struct {
		desc     string
		pattern  string
		status   int
		file     string
		expected expected
	}{
		{
			desc:    "success",
			pattern: "/dns/9007481/record",
			status:  http.StatusOK,
			file:    "./fixtures/add_new_record.json",
		},
		{
			desc:    "invalid",
			pattern: "/dns/9007481/record",
			status:  http.StatusNotImplemented,
			file:    "./fixtures/add_new_record_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(http.MethodPost, test.pattern, test.status, test.file)

			record := DNSRecord{
				Type:       "TXT",
				DomainName: "ldez.freeddns.org",
				Hostname:   "_acme-challenge.ldez.freeddns.org",
				NodeName:   "_acme-challenge",
				TextData:   "txt_txt_txt_txt_txt_txt_txt_2",
				State:      true,
				TTL:        300,
			}

			err := client.AddNewRecord(9007481, record)

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDeleteRecord(t *testing.T) {
	type expected struct {
		error string
	}

	testCases := []struct {
		desc     string
		pattern  string
		status   int
		file     string
		expected expected
	}{
		{
			desc:    "success",
			pattern: "/",
			status:  http.StatusOK,
			file:    "./fixtures/delete_record.json",
		},
		{
			desc:    "invalid",
			pattern: "/",
			status:  http.StatusNotImplemented,
			file:    "./fixtures/delete_record_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(http.MethodDelete, test.pattern, test.status, test.file)

			err := client.DeleteRecord(9007481, 6041418)

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)
		})
	}
}
