package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(subAuthID string) func(server *httptest.Server) (*Client, error) {
	return func(server *httptest.Server) (*Client, error) {
		client, err := NewClient("myAuthID", subAuthID, "myAuthPassword")
		if err != nil {
			return nil, err
		}

		client.BaseURL, _ = url.Parse(server.URL)
		client.HTTPClient = server.Client()
		return client, nil
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
		expected    expected
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
				errorMsg: "unable to unmarshal response: [status code: 200] body: [{}] error: json: cannot unmarshal array into Go value of type internal.Zone",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient("")).
				Route("GET /get-zone-info.json",
					servermock.RawStringResponse(test.apiResponse),
					servermock.CheckQueryParameter().Strict().
						With("auth-id", "myAuthID").
						With("auth-password", "myAuthPassword").
						With("domain-name", "foo.com"),
				).
				Build(t)

			zone, err := client.GetZone(t.Context(), test.authFQDN)

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
		expected    expected
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
				errorMsg: "unable to unmarshal response: [status code: 200] body: [{}] error: json: cannot unmarshal array into Go value of type map[string]internal.TXTRecord",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient("")).
				Route("GET /records.json",
					servermock.RawStringResponse(test.apiResponse),
					servermock.CheckQueryParameter().Strict().
						With("auth-id", "myAuthID").
						With("auth-password", "myAuthPassword").
						With("type", "TXT").
						With("host", "_acme-challenge").
						With("domain-name", test.zoneName),
				).
				Build(t)

			txtRecord, err := client.FindTxtRecord(t.Context(), test.zoneName, test.authFQDN)

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
		expected    expected
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
				errorMsg: "unable to unmarshal response: [status code: 200] body: [{}] error: json: cannot unmarshal array into Go value of type map[string]internal.TXTRecord",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient("")).
				Route("GET /records.json",
					servermock.RawStringResponse(test.apiResponse),
					servermock.CheckQueryParameter().Strict().
						With("auth-id", "myAuthID").
						With("auth-password", "myAuthPassword").
						With("type", "TXT").
						With("host", "_acme-challenge").
						With("domain-name", test.zoneName),
				).
				Build(t)

			txtRecords, err := client.ListTxtRecords(t.Context(), test.zoneName, test.authFQDN)

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
		query    url.Values
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
		expected    expected
	}{
		{
			desc:        "sub-zone",
			authID:      "myAuthID",
			zoneName:    "example.com",
			authFQDN:    "_acme-challenge.foo.example.com.",
			value:       "txtTXTtxtTXTtxtTXTtxtTXT",
			ttl:         60,
			apiResponse: `{"status":"Success","statusDescription":"The record was added successfully."}`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"example.com"},
					"host":          {"_acme-challenge.foo"},
					"record":        {"txtTXTtxtTXTtxtTXTtxtTXT"},
					"record-type":   {"TXT"},
					"ttl":           {"60"},
				},
			},
		},
		{
			desc:        "main zone (authID)",
			authID:      "myAuthID",
			zoneName:    "example.com",
			authFQDN:    "_acme-challenge.example.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         60,
			apiResponse: `{"status":"Success","statusDescription":"The record was added successfully."}`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"example.com"},
					"host":          {"_acme-challenge"},
					"record":        {"TXTtxtTXTtxtTXTtxtTXTtxt"},
					"record-type":   {"TXT"},
					"ttl":           {"60"},
				},
			},
		},
		{
			desc:        "main zone (subAuthID)",
			subAuthID:   "mySubAuthID",
			zoneName:    "example.com",
			authFQDN:    "_acme-challenge.example.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         60,
			apiResponse: `{"status":"Success","statusDescription":"The record was added successfully."}`,
			expected: expected{
				query: url.Values{
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"example.com"},
					"host":          {"_acme-challenge"},
					"record":        {"TXTtxtTXTtxtTXTtxtTXTtxt"},
					"record-type":   {"TXT"},
					"sub-auth-id":   {"mySubAuthID"},
					"ttl":           {"60"},
				},
			},
		},
		{
			desc:        "invalid status",
			authID:      "myAuthID",
			zoneName:    "example.com",
			authFQDN:    "_acme-challenge.example.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: `{"status":"Failed","statusDescription":"Invalid TTL. Choose from the list of the values we support."}`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"example.com"},
					"host":          {"_acme-challenge"},
					"record":        {"TXTtxtTXTtxtTXTtxtTXTtxt"},
					"record-type":   {"TXT"},
					"ttl":           {"300"},
				},
				errorMsg: "failed to add TXT record: Failed Invalid TTL. Choose from the list of the values we support.",
			},
		},
		{
			desc:        "invalid json response",
			authID:      "myAuthID",
			zoneName:    "example.com",
			authFQDN:    "_acme-challenge.example.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: `[{}]`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"example.com"},
					"host":          {"_acme-challenge"},
					"record":        {"TXTtxtTXTtxtTXTtxtTXTtxt"},
					"record-type":   {"TXT"},
					"ttl":           {"300"},
				},
				errorMsg: "unable to unmarshal response: [status code: 200] body: [{}] error: json: cannot unmarshal array into Go value of type internal.apiResponse",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient(test.subAuthID)).
				Route("POST /add-record.json",
					servermock.RawStringResponse(test.apiResponse),
					servermock.CheckQueryParameter().Strict().
						WithValues(test.expected.query),
				).
				Build(t)

			err := client.AddTxtRecord(t.Context(), test.zoneName, test.authFQDN, test.value, test.ttl)

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
		query    url.Values
		errorMsg string
	}

	testCases := []struct {
		desc        string
		id          int
		zoneName    string
		apiResponse string
		expected    expected
	}{
		{
			desc:        "record found",
			id:          5769228,
			zoneName:    "foo.com",
			apiResponse: `{ "status": "Success", "statusDescription": "The record was deleted successfully." }`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"foo.com"},
					"record-id":     {"5769228"},
				},
			},
		},
		{
			desc:        "record not found",
			id:          5769000,
			zoneName:    "foo.com",
			apiResponse: `{ "status": "Failed", "statusDescription": "Invalid record-id param." }`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"foo.com"},
					"record-id":     {"5769000"},
				},
				errorMsg: "failed to remove TXT record: Failed Invalid record-id param.",
			},
		},
		{
			desc:        "invalid json response",
			id:          44,
			zoneName:    "foo-plus.com",
			apiResponse: `[{}]`,
			expected: expected{
				query: url.Values{
					"auth-id":       {"myAuthID"},
					"auth-password": {"myAuthPassword"},
					"domain-name":   {"foo-plus.com"},
					"record-id":     {"44"},
				},
				errorMsg: "unable to unmarshal response: [status code: 200] body: [{}] error: json: cannot unmarshal array into Go value of type internal.apiResponse",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient("")).
				Route("POST /delete-record.json",
					servermock.RawStringResponse(test.apiResponse),
					servermock.CheckQueryParameter().Strict().
						WithValues(test.expected.query),
				).
				Build(t)

			err := client.RemoveTxtRecord(t.Context(), test.id, test.zoneName)

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
		expected    expected
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
			expected:    expected{errorMsg: "unable to unmarshal response: [status code: 200] body: [x] error: invalid character 'x' looking for beginning of value"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := servermock.NewBuilder[*Client](setupClient("")).
				Route("GET /update-status.json",
					servermock.RawStringResponse(test.apiResponse),
					servermock.CheckQueryParameter().Strict().
						With("auth-id", "myAuthID").
						With("auth-password", "myAuthPassword").
						With("domain-name", test.zoneName),
				).
				Build(t)

			syncProgress, err := client.GetUpdateStatus(t.Context(), test.zoneName)

			if test.expected.errorMsg != "" {
				require.EqualError(t, err, test.expected.errorMsg)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.expected.progress, syncProgress)
		})
	}
}
