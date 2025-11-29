package internal

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("apikeyvaluehere")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_FindTXTRecords(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		response string
		query    url.Values
		expected []Record
	}{
		{
			desc:     "simple",
			domain:   "example.com",
			response: "find_records.xml",
			query: url.Values{
				"name":    []string{"example.com"},
				"type":    []string{"TXT"},
				"action":  []string{"QUERY"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: []Record{
				{
					Name:     "example.org",
					Type:     "TXT",
					Content:  "txttxtx",
					TTL:      "3600 seconds",
					Priority: "0",
				},
			},
		},
		{
			desc:     "pattern",
			domain:   "**.example.com",
			response: "find_records_pattern.xml",
			query: url.Values{
				"name":    []string{"**.example.com"},
				"type":    []string{"TXT"},
				"action":  []string{"QUERY"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: []Record{
				{
					Name:     "_test.example.org",
					Type:     "TXT",
					Content:  "txttxtx",
					TTL:      "3600 seconds",
					Priority: "0",
				},
				{
					Name:     "example.org",
					Type:     "TXT",
					Content:  "txttxtx",
					TTL:      "3600 seconds",
					Priority: "0",
				},
			},
		},
		{
			desc:     "empty",
			domain:   "empty.com",
			response: "find_records_empty.xml",
			query: url.Values{
				"name":    []string{"empty.com"},
				"type":    []string{"TXT"},
				"action":  []string{"QUERY"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient).
				Route("GET /",
					servermock.ResponseFromFixture(test.response),
					servermock.CheckQueryParameter().Strict().
						WithValues(test.query)).
				Build(t)

			records, err := client.FindTXTRecords(t.Context(), test.domain)
			require.NoError(t, err)

			assert.Equal(t, test.expected, records)
		})
	}
}

func TestClient_DoActions(t *testing.T) {
	testCases := []struct {
		desc     string
		actions  []ActionParameter
		query    url.Values
		response string
		expected *DNSAPIResult
	}{
		{
			desc: "SET simple",
			actions: []ActionParameter{
				NewAddRecordAction("example.org", "txttxtx", 0),
			},
			response: "add_record.xml",
			query: url.Values{
				"action":  []string{"SET"},
				"name":    []string{"example.org"},
				"type":    []string{"TXT"},
				"value":   []string{"txttxtx"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: &DNSAPIResult{
				XMLName:      xml.Name{Space: "", Local: "dnsapi_result"},
				IsOk:         "OK:",
				ResultCounts: ResultCounts{Added: "1", Changed: "0", Unchanged: "0", Deleted: "0"},
				Actions: Actions{
					Action: Action{
						Action: "SET",
						Host:   "example.org",
						Type:   "TXT",
						Records: []Record{{
							Name:     "example.org",
							Type:     "TXT",
							Content:  "txttxtx",
							TTL:      "3600 seconds",
							Priority: "0",
						}},
					},
				},
			},
		},
		{
			desc: "SET multiple values",
			actions: []ActionParameter{
				NewAddRecordAction("example.org", "txttxtx", 0),
				NewAddRecordAction("example.org", "sample", 0),
			},
			response: "add_record_same_domain.xml",
			query: url.Values{
				"api_key":   []string{"apikeyvaluehere"},
				"action[0]": []string{"SET"},
				"name[0]":   []string{"example.org"},
				"ttl[0]":    []string{"0"},
				"type[0]":   []string{"TXT"},
				"value[0]":  []string{"txttxtx"},
				"action[1]": []string{"SET"},
				"name[1]":   []string{"example.org"},
				"ttl[1]":    []string{"0"},
				"type[1]":   []string{"TXT"},
				"value[1]":  []string{"sample"},
			},
			expected: &DNSAPIResult{
				XMLName:      xml.Name{Space: "", Local: "dnsapi_result"},
				IsOk:         "OK:",
				ResultCounts: ResultCounts{Added: "2", Changed: "0", Unchanged: "0", Deleted: "0"},
				Actions: Actions{
					Action: Action{
						Action: "SET",
						Host:   "example.org",
						Type:   "TXT",
						Records: []Record{
							{
								Name:     "example.org",
								Type:     "TXT",
								Content:  "txttxtx",
								TTL:      "0 seconds",
								Priority: "0",
							},
							{
								Name:     "example.org",
								Type:     "TXT",
								Content:  "sample",
								TTL:      "0 seconds",
								Priority: "0",
							},
						},
					},
				},
			},
		},
		{
			desc: "DELETE nothing",
			actions: []ActionParameter{
				NewDeleteRecordAction("example.org", "nothing"),
			},
			response: "delete_record_nothing.xml",
			query: url.Values{
				"action":  []string{"DELETE"},
				"name":    []string{"example.org"},
				"type":    []string{"TXT"},
				"value":   []string{"nothing"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: &DNSAPIResult{
				XMLName:      xml.Name{Space: "", Local: "dnsapi_result"},
				IsOk:         "OK:",
				ResultCounts: ResultCounts{Added: "0", Changed: "0", Unchanged: "0", Deleted: "0"},
				Actions: Actions{
					Action: Action{
						Action:  "DELETE",
						Host:    "example.org",
						Type:    "TXT",
						Records: nil,
					},
				},
			},
		},
		{
			desc: "DELETE simple",
			actions: []ActionParameter{
				NewDeleteRecordAction("example.org", "txttxtx"),
			},
			response: "delete_record.xml",
			query: url.Values{
				"action":  []string{"DELETE"},
				"name":    []string{"example.org"},
				"type":    []string{"TXT"},
				"value":   []string{"txttxtx"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: &DNSAPIResult{
				XMLName:      xml.Name{Space: "", Local: "dnsapi_result"},
				IsOk:         "OK:",
				ResultCounts: ResultCounts{Added: "0", Changed: "0", Unchanged: "0", Deleted: "1"},
				Actions: Actions{
					Action: Action{
						Action: "DELETE",
						Host:   "example.org",
						Type:   "TXT",
						Records: []Record{{
							Name:     "example.org",
							Type:     "TXT",
							Content:  "txttxtx",
							TTL:      "3600 seconds",
							Priority: "0",
						}},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient).
				Route("GET /",
					servermock.ResponseFromFixture(test.response),
					servermock.CheckQueryParameter().Strict().
						WithValues(test.query)).
				Build(t)

			resp, err := client.DoActions(t.Context(), test.actions...)
			require.NoError(t, err)

			assert.Equal(t, test.expected, resp)
		})
	}
}

func TestClient_DoActions_error(t *testing.T) {
	testCases := []struct {
		desc     string
		actions  []ActionParameter
		query    url.Values
		response string
		expected string
	}{
		{
			desc: "SET error",
			actions: []ActionParameter{
				NewAddRecordAction("example.com", "txttxtx", 0),
			},
			response: "add_record_error.xml",
			query: url.Values{
				"action":  []string{"SET"},
				"name":    []string{"example.com"},
				"type":    []string{"TXT"},
				"value":   []string{"txttxtx"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: "ERROR: No zone found for example.com",
		},
		{
			desc: "DELETE error",
			actions: []ActionParameter{
				NewDeleteRecordAction("example.com", "txttxtx"),
			},
			response: "delete_record_error.xml",
			query: url.Values{
				"action":  []string{"DELETE"},
				"name":    []string{"example.com"},
				"type":    []string{"TXT"},
				"value":   []string{"txttxtx"},
				"api_key": []string{"apikeyvaluehere"},
			},
			expected: "ERROR: No zone found for example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient).
				Route("GET /",
					servermock.ResponseFromFixture(test.response).
						WithStatusCode(http.StatusInternalServerError),
					servermock.CheckQueryParameter().Strict().
						WithValues(test.query)).
				Build(t)

			_, err := client.DoActions(t.Context(), test.actions...)
			require.EqualError(t, err, test.expected)
		})
	}
}
