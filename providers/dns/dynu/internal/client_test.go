package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient()
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
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
			pattern: "GET /dns/getroot/test.lego.freeddns.org",
			status:  http.StatusOK,
			file:    "get_root_domain.json",
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
			pattern: "GET /dns/getroot/test.lego.freeddns.org",
			status:  http.StatusNotImplemented,
			file:    "get_root_domain_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route(test.pattern, servermock.ResponseFromFixture(test.file).WithStatusCode(test.status)).
				Build(t)

			domain, err := client.GetRootDomain(t.Context(), "test.lego.freeddns.org")

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
			pattern: "GET /dns/record/_acme-challenge.lego.freeddns.org",
			status:  http.StatusOK,
			file:    "get_records.json",
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
			pattern: "GET /dns/record/_acme-challenge.lego.freeddns.org",
			status:  http.StatusOK,
			file:    "get_records_empty.json",
			expected: expected{
				records: []DNSRecord{},
			},
		},
		{
			desc:    "invalid",
			pattern: "GET /dns/record/_acme-challenge.lego.freeddns.org",
			status:  http.StatusNotImplemented,
			file:    "get_records_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route(test.pattern, servermock.ResponseFromFixture(test.file).WithStatusCode(test.status),
					servermock.CheckQueryParameter().Strict().
						With("recordType", "TXT")).
				Build(t)

			records, err := client.GetRecords(t.Context(), "_acme-challenge.lego.freeddns.org", "TXT")

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
			pattern: "POST /dns/9007481/record",
			status:  http.StatusOK,
			file:    "add_new_record.json",
		},
		{
			desc:    "invalid",
			pattern: "POST /dns/9007481/record",
			status:  http.StatusNotImplemented,
			file:    "add_new_record_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route(test.pattern, servermock.ResponseFromFixture(test.file).WithStatusCode(test.status),
					servermock.CheckRequestJSONBodyFromFile("add_new_record-request.json")).
				Build(t)

			record := DNSRecord{
				Type:       "TXT",
				DomainName: "lego.freeddns.org",
				Hostname:   "_acme-challenge.lego.freeddns.org",
				NodeName:   "_acme-challenge",
				TextData:   "txt_txt_txt_txt_txt_txt_txt_2",
				State:      true,
				TTL:        300,
			}

			err := client.AddNewRecord(t.Context(), 9007481, record)

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
			pattern: "DELETE /",
			status:  http.StatusOK,
			file:    "delete_record.json",
		},
		{
			desc:    "invalid",
			pattern: "DELETE /",
			status:  http.StatusNotImplemented,
			file:    "delete_record_invalid.json",
			expected: expected{
				error: "API error: 501: Argument Exception: Invalid.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route(test.pattern, servermock.ResponseFromFixture(test.file).WithStatusCode(test.status)).
				Build(t)

			err := client.DeleteRecord(t.Context(), 9007481, 6041418)

			if test.expected.error != "" {
				assert.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)
		})
	}
}
