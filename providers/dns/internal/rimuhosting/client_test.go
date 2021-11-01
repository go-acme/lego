package rimuhosting

import (
	"encoding/xml"
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

func TestClient_FindTXTRecords(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		var fixture string
		switch query.Get("name") {
		case "example.com":
			fixture = "./fixtures/find_records.xml"
		case "**.example.com":
			fixture = "./fixtures/find_records_pattern.xml"
		default:
			fixture = "./fixtures/find_records_empty.xml"
		}

		err := writeResponse(rw, fixture)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient("apikeyvaluehere")
	client.BaseURL = server.URL

	testCases := []struct {
		desc     string
		domain   string
		expected []Record
	}{
		{
			desc:   "simple",
			domain: "example.com",
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
			desc:   "pattern",
			domain: "**.example.com",
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
			expected: nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			records, err := client.FindTXTRecords(test.domain)
			require.NoError(t, err)

			assert.Equal(t, test.expected, records)
		})
	}
}

func TestClient_DoActions(t *testing.T) {
	type expected struct {
		Query string
		Resp  *DNSAPIResult
		Error string
	}

	testCases := []struct {
		desc     string
		actions  []ActionParameter
		fixture  string
		expected expected
	}{
		{
			desc: "SET error",
			actions: []ActionParameter{
				AddRecord("example.com", "txttxtx", 0),
			},
			fixture: "./fixtures/add_record_error.xml",
			expected: expected{
				Query: "action=SET&api_key=apikeyvaluehere&name=example.com&type=TXT&value=txttxtx",
				Error: "ERROR: No zone found for example.com",
			},
		},
		{
			desc: "SET simple",
			actions: []ActionParameter{
				AddRecord("example.org", "txttxtx", 0),
			},
			fixture: "./fixtures/add_record.xml",
			expected: expected{
				Query: "action=SET&api_key=apikeyvaluehere&name=example.org&type=TXT&value=txttxtx",
				Resp: &DNSAPIResult{
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
		},
		{
			desc: "SET multiple values",
			actions: []ActionParameter{
				AddRecord("example.org", "txttxtx", 0),
				AddRecord("example.org", "sample", 0),
			},
			fixture: "./fixtures/add_record_same_domain.xml",
			expected: expected{
				Query: "action[0]=SET&action[1]=SET&api_key=apikeyvaluehere&name[0]=example.org&name[1]=example.org&ttl[0]=0&ttl[1]=0&type[0]=TXT&type[1]=TXT&value[0]=txttxtx&value[1]=sample",
				Resp: &DNSAPIResult{
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
		},
		{
			desc: "DELETE error",
			actions: []ActionParameter{
				DeleteRecord("example.com", "txttxtx"),
			},
			fixture: "./fixtures/delete_record_error.xml",
			expected: expected{
				Query: "action=DELETE&api_key=apikeyvaluehere&name=example.com&type=TXT&value=txttxtx",
				Error: "ERROR: No zone found for example.com",
			},
		},
		{
			desc: "DELETE nothing",
			actions: []ActionParameter{
				DeleteRecord("example.org", "nothing"),
			},
			fixture: "./fixtures/delete_record_nothing.xml",
			expected: expected{
				Query: "action=DELETE&api_key=apikeyvaluehere&name=example.org&type=TXT&value=nothing",
				Resp: &DNSAPIResult{
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
		},
		{
			desc: "DELETE simple",
			actions: []ActionParameter{
				DeleteRecord("example.org", "txttxtx"),
			},
			fixture: "./fixtures/delete_record.xml",
			expected: expected{
				Query: "action=DELETE&api_key=apikeyvaluehere&name=example.org&type=TXT&value=txttxtx",
				Resp: &DNSAPIResult{
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
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
				query, err := url.QueryUnescape(req.URL.RawQuery)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}

				if test.expected.Query != query {
					http.Error(rw, fmt.Sprintf("invalid query: %s", query), http.StatusBadRequest)
					return
				}

				if test.expected.Error != "" {
					rw.WriteHeader(http.StatusInternalServerError)
				}

				err = writeResponse(rw, test.fixture)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			})

			client := NewClient("apikeyvaluehere")
			client.BaseURL = server.URL

			resp, err := client.DoActions(test.actions...)
			if test.expected.Error != "" {
				require.EqualError(t, err, test.expected.Error)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, test.expected.Resp, resp)
		})
	}
}

func writeResponse(rw io.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	_, err = io.Copy(rw, file)
	return err
}
