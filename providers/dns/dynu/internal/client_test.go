package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, method, pattern string, status int, file string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

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
			pattern: "/dns/getroot/test.lego.freeddns.org",
			status:  http.StatusOK,
			file:    "./fixtures/get_root_domain.json",
			expected: expected{
				domain: &DNSHostname{
					APIException: &APIException{
						StatusCode: 200,
					},
					ID:         9007481,
					DomainName: "lego.freeddns.org",
					Hostname:   "test.lego.freeddns.org",
					Node:       "test",
				},
			},
		},
		{
			desc:    "invalid",
			pattern: "/dns/getroot/test.lego.freeddns.org",
			status:  http.StatusNotImplemented,
			file:    "./fixtures/get_root_domain_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(t, http.MethodGet, test.pattern, test.status, test.file)

			domain, err := client.GetRootDomain(context.Background(), "test.lego.freeddns.org")

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, domain) //nolint:testifylint // false positive https://github.com/Antonboom/testifylint/issues/95
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
			pattern: "/dns/record/_acme-challenge.lego.freeddns.org",
			status:  http.StatusOK,
			file:    "./fixtures/get_records.json",
			expected: expected{
				records: []DNSRecord{
					{
						ID:         6041417,
						Type:       "TXT",
						DomainID:   9007481,
						DomainName: "lego.freeddns.org",
						NodeName:   "_acme-challenge",
						Hostname:   "_acme-challenge.lego.freeddns.org",
						State:      true,
						Content:    `_acme-challenge.lego.freeddns.org. 300 IN TXT "txt_txt_txt_txt_txt_txt_txt"`,
						TextData:   "txt_txt_txt_txt_txt_txt_txt",
						TTL:        300,
					},
					{
						ID:         6041422,
						Type:       "TXT",
						DomainID:   9007481,
						DomainName: "lego.freeddns.org",
						NodeName:   "_acme-challenge",
						Hostname:   "_acme-challenge.lego.freeddns.org",
						State:      true,
						Content:    `_acme-challenge.lego.freeddns.org. 300 IN TXT "txt_txt_txt_txt_txt_txt_txt_2"`,
						TextData:   "txt_txt_txt_txt_txt_txt_txt_2",
						TTL:        300,
					},
				},
			},
		},
		{
			desc:    "empty",
			pattern: "/dns/record/_acme-challenge.lego.freeddns.org",
			status:  http.StatusOK,
			file:    "./fixtures/get_records_empty.json",
			expected: expected{
				records: []DNSRecord{},
			},
		},
		{
			desc:    "invalid",
			pattern: "/dns/record/_acme-challenge.lego.freeddns.org",
			status:  http.StatusNotImplemented,
			file:    "./fixtures/get_records_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(t, http.MethodGet, test.pattern, test.status, test.file)

			records, err := client.GetRecords(context.Background(), "_acme-challenge.lego.freeddns.org", "TXT")

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)

			assert.NotNil(t, records)
			assert.Equal(t, test.expected.records, records)
		})
	}
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(t, http.MethodPost, test.pattern, test.status, test.file)

			record := DNSRecord{
				Type:       "TXT",
				DomainName: "lego.freeddns.org",
				Hostname:   "_acme-challenge.lego.freeddns.org",
				NodeName:   "_acme-challenge",
				TextData:   "txt_txt_txt_txt_txt_txt_txt_2",
				State:      true,
				TTL:        300,
			}

			err := client.AddNewRecord(context.Background(), 9007481, record)

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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setupTest(t, http.MethodDelete, test.pattern, test.status, test.file)

			err := client.DeleteRecord(context.Background(), 9007481, 6041418)

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)
		})
	}
}
