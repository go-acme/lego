package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handlerMock(method string, jsonData []byte) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, "Incorrect method used", http.StatusBadRequest)
			return
		}

		_, err := rw.Write(jsonData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func TestClientNewCient(t *testing.T) {
	type expectedResult struct {
		client   *Client
		errorMsg string
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	testCases := []struct {
		desc         string
		authID       string
		subAuthID    string
		authPassword string
		expected     expectedResult
	}{
		{
			desc:         "all provided",
			authID:       "1000",
			subAuthID:    "1111",
			authPassword: "no-secret",
			expected: expectedResult{
				client:   &Client{"1000", "1111", "no-secret", &http.Client{}, baseURL},
				errorMsg: "",
			},
		},
		{
			desc:         "missing authID & subAuthID",
			authID:       "",
			subAuthID:    "",
			authPassword: "no-secret",
			expected: expectedResult{
				client:   nil,
				errorMsg: "credentials missing: authID or subAuthID",
			},
		},
		{
			desc:         "missing authID & subAuthID",
			authID:       "",
			subAuthID:    "present",
			authPassword: "",
			expected: expectedResult{
				client:   nil,
				errorMsg: "credentials missing: authPassword",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client, err := NewClient(test.authID, test.subAuthID, test.authPassword)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.client, client)
			}
		})
	}
}

func TestClientGetZone(t *testing.T) {
	type expectedResult struct {
		zone     *Zone
		errorMsg string
	}
	testCases := []struct {
		desc        string
		authFQDN    string
		apiResponse []byte
		expected    expectedResult
	}{
		{
			desc:        "zone found",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: []byte(`{"name": "foo.com", "type": "master", "zone": "zone", "status": "1"}`),
			expected: expectedResult{
				zone: &Zone{
					Name:   "foo.com",
					Type:   "master",
					Zone:   "zone",
					Status: "1",
				},
				errorMsg: "",
			},
		},
		{
			desc:        "zone not found",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: []byte(``),
			expected: expectedResult{
				zone:     nil,
				errorMsg: "zone foo.com not found for authFQDN _acme-challenge.foo.com.",
			},
		},
		{
			desc:        "invalid json",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: []byte(`[{}]`),
			expected: expectedResult{
				zone:     nil,
				errorMsg: "GetZone() unmarshaling error: json: cannot unmarshal array into Go value of type internal.Zone",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.apiResponse))

			client, _ := NewClient("myAuthID", "", "myAuthPassword")
			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			zone, err := client.GetZone(test.authFQDN)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.zone, zone)
			}
		})
	}
}

func TestClientFindTxtRecord(t *testing.T) {
	type expectedResult struct {
		txtRecord *TXTRecord
		errorMsg  string
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		zoneName    string
		apiResponse []byte
		expected    expectedResult
	}{
		{
			desc:     "record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: []byte(`{   "5769228": {"id": "5769228",   "type": "TXT", "host": "_acme-challenge",   "record": "txtTXTtxtTXTtxtTXTtxtTXT", "failover": "0","ttl": "3600","status": 1},
								   "181805209": {"id": "181805209", "type": "TXT", "host": "_github-challenge", "record": "b66b8324b5",               "failover": "0","ttl": "300","status": 1}}`),
			expected: expectedResult{
				txtRecord: &TXTRecord{
					ID:       5769228,
					Type:     "TXT",
					Host:     "_acme-challenge",
					Record:   "txtTXTtxtTXTtxtTXTtxtTXT",
					Failover: 0,
					TTL:      3600,
					Status:   1,
				},
				errorMsg: "",
			},
		},
		{
			desc:     "no record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: []byte(`{   "5769228": {"id": "5769228",   "type": "TXT", "host": "_other-challenge",  "record": "txtTXTtxtTXTtxtTXTtxtTXT", "failover": "0","ttl": "3600","status": 1},
								   "181805209": {"id": "181805209", "type": "TXT", "host": "_github-challenge", "record": "b66b8324b5",               "failover": "0","ttl": "300","status": 1}}`),
			expected: expectedResult{
				txtRecord: nil,
				errorMsg:  "",
			},
		},
		{
			desc:        "zero records",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: []byte(`[]`),
			expected: expectedResult{
				txtRecord: nil,
				errorMsg:  "",
			},
		},
		{
			desc:        "invalid json",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: []byte(`[x]`),
			expected: expectedResult{
				txtRecord: nil,
				errorMsg:  "FindTxtRecord() unmarshaling error: invalid character 'x' looking for beginning of value: [x]",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.apiResponse))

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			txtRecord, err := client.FindTxtRecord(test.zoneName, test.authFQDN)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.txtRecord, txtRecord)
			}
		})
	}
}

func TestClientAddTxtRecord(t *testing.T) {
	type expectedResult struct {
		query    string
		errorMsg string
	}

	testCases := []struct {
		desc        string
		authID      string
		subAuthID   string
		zoneName    string
		authFQDN    string
		value       string
		ttl         int
		apiResponse []byte
		expected    expectedResult
	}{
		{
			desc:        "sub-zone",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.foo.bar.com.",
			value:       "txtTXTtxtTXTtxtTXTtxtTXT",
			ttl:         60,
			apiResponse: []byte(`{"status":"Success","statusDescription":"The record was added successfully."}`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge.foo&record=txtTXTtxtTXTtxtTXTtxtTXT&record-type=TXT&ttl=60`,
				errorMsg: "",
			},
		},
		{
			desc:        "main zone (authID)",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         444444444444444444,
			apiResponse: []byte(`{"status":"Success","statusDescription":"The record was added successfully."}`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=2592000`,
				errorMsg: "",
			},
		},
		{
			desc:        "main zone (subAuthID)",
			subAuthID:   "mySubAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         60,
			apiResponse: []byte(`{"status":"Success","statusDescription":"The record was added successfully."}`),
			expected: expectedResult{
				query:    `auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&sub-auth-id=mySubAuthID&ttl=60`,
				errorMsg: "",
			},
		},
		{
			desc:        "invalid status",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: []byte(`{"status":"Failed","statusDescription":"Invalid TTL. Choose from the list of the values we support."}`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=300`,
				errorMsg: "failed to add TXT record: Failed Invalid TTL. Choose from the list of the values we support.",
			},
		},
		{
			desc:        "invalid json reply",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: []byte(`{},{}`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=300`,
				errorMsg: "AddTxtRecord() unmarshaling error: invalid character ',' after top-level value: {},{}",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.NotNil(t, req.URL.RawQuery)
				assert.Equal(t, test.expected.query, req.URL.RawQuery)

				handlerMock(http.MethodPost, test.apiResponse).ServeHTTP(rw, req)
			}))

			client, err := NewClient(test.authID, test.subAuthID, "myAuthPassword")
			require.NoError(t, err)

			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			err = client.AddTxtRecord(test.zoneName, test.authFQDN, test.value, test.ttl)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientRemoveTxtRecord(t *testing.T) {
	type expectedResult struct {
		query    string
		errorMsg string
	}

	testCases := []struct {
		desc        string
		id          int
		zoneName    string
		apiResponse []byte
		expected    expectedResult
	}{
		{
			desc:        "record found",
			id:          5769228,
			zoneName:    "foo.com",
			apiResponse: []byte(`{ "status": "Success", "statusDescription": "The record was deleted successfully." }`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=foo.com&record-id=5769228`,
				errorMsg: "",
			},
		},
		{
			desc:        "record not found",
			id:          5769000,
			zoneName:    "foo.com",
			apiResponse: []byte(`{ "status": "Failed", "statusDescription": "Invalid record-id param." }`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=foo.com&record-id=5769000`,
				errorMsg: "failed to remove TXT record: Failed Invalid record-id param.",
			},
		},
		{
			desc:        "invalid json",
			id:          44,
			zoneName:    "foo-plus.com",
			apiResponse: []byte(`[]`),
			expected: expectedResult{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=foo-plus.com&record-id=44`,
				errorMsg: "RemoveTxtRecord() unmarshaling error: json: cannot unmarshal array into Go value of type internal.apiResponse: []",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// server := httptest.NewServer(handlerMock(http.MethodPost, test.apiResponse))

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.NotNil(t, req.URL.RawQuery)
				assert.Equal(t, test.expected.query, req.URL.RawQuery)

				handlerMock(http.MethodPost, test.apiResponse).ServeHTTP(rw, req)
			}))

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			err = client.RemoveTxtRecord(test.id, test.zoneName)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
				// assert.Equal(t, test.expected.txtRecord, txtRecord)
			}
		})
	}
}

func TestClientGetUpdateStatus(t *testing.T) {
	type expectedResult struct {
		progress SyncProgress
		errorMsg string
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		zoneName    string
		apiResponse []byte
		expected    expectedResult
	}{
		{
			desc:     "50% sync",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: []byte(`[{"server": "ns101.foo.com.", "ip4": "10.11.12.13", "ip6": "2a00:2a00:2a00:9::5", "updated": true },
								  {"server": "ns102.foo.com.", "ip4": "10.14.16.17", "ip6": "2100:2100:2100:3::1", "updated": false }]`),
			expected: expectedResult{SyncProgress{false, 1, 2}, ""},
		},
		{
			desc:     "100% sync",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: []byte(`[{"server": "ns101.foo.com.", "ip4": "10.11.12.13", "ip6": "2a00:2a00:2a00:9::5", "updated": true },
								  {"server": "ns102.foo.com.", "ip4": "10.14.16.17", "ip6": "2100:2100:2100:3::1", "updated": true }]`),
			expected: expectedResult{SyncProgress{true, 2, 2}, ""},
		},
		{
			desc:        "record not found",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: []byte(`[]`),
			expected:    expectedResult{SyncProgress{false, 0, 0}, "No nameservers records returned"},
		},
		{
			desc:        "invalid json reply",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: []byte(`{},{}`),
			expected:    expectedResult{SyncProgress{false, 0, 0}, "GetUpdateStatus() unmarshaling error: invalid character ',' after top-level value: {},{}"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.apiResponse))

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			syncProgress, err := client.GetUpdateStatus(test.zoneName)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.progress, syncProgress)
			}
		})
	}
}
