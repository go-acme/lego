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

func handlerMock(method string, jsonData []byte) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, "Incorrect method used", http.StatusBadRequest)
			return
		}

		_, err := rw.Write(jsonData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func TestNewClient(t *testing.T) {
	testCases := []struct {
		desc         string
		authID       string
		subAuthID    string
		authPassword string
		expected     string
	}{
		{
			desc:         "all provided",
			authID:       "1000",
			subAuthID:    "1111",
			authPassword: "no-secret",
		},
		{
			desc:         "missing authID & subAuthID",
			authID:       "",
			subAuthID:    "",
			authPassword: "no-secret",
			expected:     "credentials missing: authID or subAuthID",
		},
		{
			desc:         "missing authID & subAuthID",
			authID:       "",
			subAuthID:    "present",
			authPassword: "",
			expected:     "credentials missing: authPassword",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client, err := NewClient(test.authID, test.subAuthID, test.authPassword)

			if test.expected != "" {
				assert.Nil(t, client)
				require.EqualError(t, err, test.expected)
			} else {
				assert.NotNil(t, client)
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetZone(t *testing.T) {
	type expected struct {
		zone     *Zone
		errorMsg string
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		apiResponse string
		expected
	}{
		{
			desc:        "zone found",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: `{"name": "foo.com", "type": "master", "zone": "zone", "status": "1"}`,
			expected: expected{
				zone: &Zone{
					Name:   "foo.com",
					Type:   "master",
					Zone:   "zone",
					Status: "1",
				},
			},
		},
		{
			desc:        "zone not found",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: ``,
			expected: expected{
				errorMsg: "zone foo.com not found for authFQDN _acme-challenge.foo.com.",
			},
		},
		{
			desc:        "invalid json response",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: `[{}]`,
			expected: expected{
				errorMsg: "failed to unmarshal zone: json: cannot unmarshal array into Go value of type internal.Zone",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, []byte(test.apiResponse)))
			t.Cleanup(server.Close)

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			client.BaseURL, _ = url.Parse(server.URL)

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

func TestClient_FindTxtRecord(t *testing.T) {
	type expected struct {
		txtRecord *TXTRecord
		errorMsg  string
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		zoneName    string
		apiResponse string
		expected
	}{
		{
			desc:     "record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: `{
  "5769228": {
    "id": "5769228",
    "type": "TXT",
    "host": "_acme-challenge",
    "record": "txtTXTtxtTXTtxtTXTtxtTXT",
    "failover": "0",
    "ttl": "3600",
    "status": 1
  },
  "181805209": {
    "id": "181805209",
    "type": "TXT",
    "host": "_github-challenge",
    "record": "b66b8324b5",
    "failover": "0",
    "ttl": "300",
    "status": 1
  }
}`,
			expected: expected{
				txtRecord: &TXTRecord{
					ID:       5769228,
					Type:     "TXT",
					Host:     "_acme-challenge",
					Record:   "txtTXTtxtTXTtxtTXTtxtTXT",
					Failover: 0,
					TTL:      3600,
					Status:   1,
				},
			},
		},
		{
			desc:     "no record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: `{
  "5769228": {
    "id": "5769228",
    "type": "TXT",
    "host": "_other-challenge",
    "record": "txtTXTtxtTXTtxtTXTtxtTXT",
    "failover": "0",
    "ttl": "3600",
    "status": 1
  },
  "181805209": {
    "id": "181805209",
    "type": "TXT",
    "host": "_github-challenge",
    "record": "b66b8324b5",
    "failover": "0",
    "ttl": "300",
    "status": 1
  }
}`,
		},
		{
			desc:        "zero records",
			authFQDN:    "_acme-challenge.example.com.",
			zoneName:    "example.com",
			apiResponse: `[]`,
		},
		{
			desc:        "invalid json response",
			authFQDN:    "_acme-challenge.example.com.",
			zoneName:    "example.com",
			apiResponse: `[{}]`,
			expected: expected{
				errorMsg: "failed to unmarshall TXT records: json: cannot unmarshal array into Go value of type map[string]internal.TXTRecord: [{}]",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, []byte(test.apiResponse)))
			t.Cleanup(server.Close)

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			client.BaseURL, _ = url.Parse(server.URL)

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

func TestClient_ListTxtRecord(t *testing.T) {
	type expected struct {
		txtRecords []TXTRecord
		errorMsg   string
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		zoneName    string
		apiResponse string
		expected
	}{
		{
			desc:     "record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: `{
  "5769228": {
    "id": "5769228",
    "type": "TXT",
    "host": "_acme-challenge",
    "record": "txtTXTtxtTXTtxtTXTtxtTXT",
    "failover": "0",
    "ttl": "3600",
    "status": 1
  },
  "181805209": {
    "id": "181805209",
    "type": "TXT",
    "host": "_github-challenge",
    "record": "b66b8324b5",
    "failover": "0",
    "ttl": "300",
    "status": 1
  }
}`,
			expected: expected{
				txtRecords: []TXTRecord{
					{
						ID:       5769228,
						Type:     "TXT",
						Host:     "_acme-challenge",
						Record:   "txtTXTtxtTXTtxtTXTtxtTXT",
						Failover: 0,
						TTL:      3600,
						Status:   1,
					},
				},
			},
		},
		{
			desc:     "no record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: `{
  "5769228": {
    "id": "5769228",
    "type": "TXT",
    "host": "_other-challenge",
    "record": "txtTXTtxtTXTtxtTXTtxtTXT",
    "failover": "0",
    "ttl": "3600",
    "status": 1
  },
  "181805209": {
    "id": "181805209",
    "type": "TXT",
    "host": "_github-challenge",
    "record": "b66b8324b5",
    "failover": "0",
    "ttl": "300",
    "status": 1
  }
}`,
		},
		{
			desc:        "zero records",
			authFQDN:    "_acme-challenge.example.com.",
			zoneName:    "example.com",
			apiResponse: `[]`,
		},
		{
			desc:        "invalid json response",
			authFQDN:    "_acme-challenge.example.com.",
			zoneName:    "example.com",
			apiResponse: `[{}]`,
			expected: expected{
				errorMsg: "failed to unmarshall TXT records: json: cannot unmarshal array into Go value of type map[string]internal.TXTRecord: [{}]",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, []byte(test.apiResponse)))
			t.Cleanup(server.Close)

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			client.BaseURL, _ = url.Parse(server.URL)

			txtRecords, err := client.ListTxtRecords(test.zoneName, test.authFQDN)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.txtRecords, txtRecords)
			}
		})
	}
}

func TestClient_AddTxtRecord(t *testing.T) {
	type expected struct {
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
		apiResponse string
		expected
	}{
		{
			desc:        "sub-zone",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.foo.bar.com.",
			value:       "txtTXTtxtTXTtxtTXTtxtTXT",
			ttl:         60,
			apiResponse: `{"status":"Success","statusDescription":"The record was added successfully."}`,
			expected: expected{
				query: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge.foo&record=txtTXTtxtTXTtxtTXTtxtTXT&record-type=TXT&ttl=60`,
			},
		},
		{
			desc:        "main zone (authID)",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         60,
			apiResponse: `{"status":"Success","statusDescription":"The record was added successfully."}`,
			expected: expected{
				query: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=60`,
			},
		},
		{
			desc:        "main zone (subAuthID)",
			subAuthID:   "mySubAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         60,
			apiResponse: `{"status":"Success","statusDescription":"The record was added successfully."}`,
			expected: expected{
				query: `auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&sub-auth-id=mySubAuthID&ttl=60`,
			},
		},
		{
			desc:        "invalid status",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: `{"status":"Failed","statusDescription":"Invalid TTL. Choose from the list of the values we support."}`,
			expected: expected{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=300`,
				errorMsg: "failed to add TXT record: Failed Invalid TTL. Choose from the list of the values we support.",
			},
		},
		{
			desc:        "invalid json response",
			authID:      "myAuthID",
			zoneName:    "bar.com",
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: `[{}]`,
			expected: expected{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=300`,
				errorMsg: "failed to unmarshal API response: json: cannot unmarshal array into Go value of type internal.apiResponse: [{}]",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if test.expected.query != req.URL.RawQuery {
					msg := fmt.Sprintf("got: %s, want: %s", test.expected.query, req.URL.RawQuery)
					http.Error(rw, msg, http.StatusBadRequest)
					return
				}

				handlerMock(http.MethodPost, []byte(test.apiResponse))(rw, req)
			}))
			t.Cleanup(server.Close)

			client, err := NewClient(test.authID, test.subAuthID, "myAuthPassword")
			require.NoError(t, err)

			client.BaseURL, _ = url.Parse(server.URL)

			err = client.AddTxtRecord(test.zoneName, test.authFQDN, test.value, test.ttl)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	type expected struct {
		query    string
		errorMsg string
	}

	testCases := []struct {
		desc        string
		id          int
		zoneName    string
		apiResponse string
		expected
	}{
		{
			desc:        "record found",
			id:          5769228,
			zoneName:    "foo.com",
			apiResponse: `{ "status": "Success", "statusDescription": "The record was deleted successfully." }`,
			expected: expected{
				query: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=foo.com&record-id=5769228`,
			},
		},
		{
			desc:        "record not found",
			id:          5769000,
			zoneName:    "foo.com",
			apiResponse: `{ "status": "Failed", "statusDescription": "Invalid record-id param." }`,
			expected: expected{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=foo.com&record-id=5769000`,
				errorMsg: "failed to remove TXT record: Failed Invalid record-id param.",
			},
		},
		{
			desc:        "invalid json response",
			id:          44,
			zoneName:    "foo-plus.com",
			apiResponse: `[{}]`,
			expected: expected{
				query:    `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=foo-plus.com&record-id=44`,
				errorMsg: "failed to unmarshal API response: json: cannot unmarshal array into Go value of type internal.apiResponse: [{}]",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if test.expected.query != req.URL.RawQuery {
					msg := fmt.Sprintf("got: %s, want: %s", test.expected.query, req.URL.RawQuery)
					http.Error(rw, msg, http.StatusBadRequest)
					return
				}

				handlerMock(http.MethodPost, []byte(test.apiResponse))(rw, req)
			}))
			t.Cleanup(server.Close)

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			client.BaseURL, _ = url.Parse(server.URL)

			err = client.RemoveTxtRecord(test.id, test.zoneName)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetUpdateStatus(t *testing.T) {
	type expected struct {
		progress *SyncProgress
		errorMsg string
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		zoneName    string
		apiResponse string
		expected
	}{
		{
			desc:     "50% sync",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: `[
{"server": "ns101.foo.com.", "ip4": "10.11.12.13", "ip6": "2a00:2a00:2a00:9::5", "updated": true },
{"server": "ns102.foo.com.", "ip4": "10.14.16.17", "ip6": "2100:2100:2100:3::1", "updated": false }
]`,
			expected: expected{progress: &SyncProgress{Updated: 1, Total: 2}},
		},
		{
			desc:     "100% sync",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: `[
{"server": "ns101.foo.com.", "ip4": "10.11.12.13", "ip6": "2a00:2a00:2a00:9::5", "updated": true },
{"server": "ns102.foo.com.", "ip4": "10.14.16.17", "ip6": "2100:2100:2100:3::1", "updated": true }
]`,
			expected: expected{progress: &SyncProgress{Complete: true, Updated: 2, Total: 2}},
		},
		{
			desc:        "record not found",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: `[]`,
			expected:    expected{errorMsg: "no nameservers records returned"},
		},
		{
			desc:        "invalid json response",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: `[x]`,
			expected:    expected{errorMsg: "failed to unmarshal UpdateRecord: invalid character 'x' looking for beginning of value: [x]"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, []byte(test.apiResponse)))
			t.Cleanup(server.Close)

			client, err := NewClient("myAuthID", "", "myAuthPassword")
			require.NoError(t, err)

			client.BaseURL, _ = url.Parse(server.URL)

			syncProgress, err := client.GetUpdateStatus(test.zoneName)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.expected.progress, syncProgress)
		})
	}
}
