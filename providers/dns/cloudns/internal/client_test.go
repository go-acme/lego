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

func TestClientGetZone(t *testing.T) {
	type result struct {
		zone  *Zone
		error bool
	}
	testCases := []struct {
		desc        string
		authFQDN    string
		apiResponse []byte
		expected    result
	}{
		{
			desc:        "zone found",
			authFQDN:    "_acme-challenge.foo.com.",
			apiResponse: []byte(`{"name": "foo.com", "type": "master", "zone": "zone", "status": "1"}`),
			expected: result{
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
			apiResponse: []byte(``),
			expected:    result{error: true},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.apiResponse))

			client, _ := NewClient("myAuthID", "myAuthPassword")
			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			zone, err := client.GetZone(test.authFQDN)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.zone, zone)
			}
		})
	}
}

func TestClientFindTxtRecord(t *testing.T) {
	type result struct {
		txtRecord *TXTRecord
		error     bool
	}

	testCases := []struct {
		desc        string
		authFQDN    string
		zoneName    string
		apiResponse []byte
		expected    result
	}{
		{
			desc:     "record found",
			authFQDN: "_acme-challenge.foo.com.",
			zoneName: "foo.com",
			apiResponse: []byte(`{
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
}`),
			expected: result{
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
			desc:        "record not found",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
			apiResponse: []byte(`[]`),
			expected:    result{txtRecord: nil},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.apiResponse))

			client, _ := NewClient("myAuthID", "myAuthPassword")
			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			txtRecord, err := client.FindTxtRecord(test.zoneName, test.authFQDN)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.txtRecord, txtRecord)
			}
		})
	}
}

func TestClientAddTxtRecord(t *testing.T) {
	type expected struct {
		Query string
		Error string
	}

	testCases := []struct {
		desc        string
		zone        *Zone
		authFQDN    string
		value       string
		ttl         int
		apiResponse []byte
		expected    expected
	}{
		{
			desc: "sub-zone",
			zone: &Zone{
				Name:   "bar.com",
				Type:   "master",
				Zone:   "domain",
				Status: "1",
			},
			authFQDN:    "_acme-challenge.foo.bar.com.",
			value:       "txtTXTtxtTXTtxtTXTtxtTXT",
			ttl:         60,
			apiResponse: []byte(`{"status":"Success","statusDescription":"The record was added successfully."}`),
			expected: expected{
				Query: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge.foo&record=txtTXTtxtTXTtxtTXTtxtTXT&record-type=TXT&ttl=60`,
			},
		},
		{
			desc: "main zone",
			zone: &Zone{
				Name:   "bar.com",
				Type:   "master",
				Zone:   "domain",
				Status: "1",
			},
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         60,
			apiResponse: []byte(`{"status":"Success","statusDescription":"The record was added successfully."}`),
			expected: expected{
				Query: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=60`,
			},
		},
		{
			desc: "invalid status",
			zone: &Zone{
				Name:   "bar.com",
				Type:   "master",
				Zone:   "domain",
				Status: "1",
			},
			authFQDN:    "_acme-challenge.bar.com.",
			value:       "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:         120,
			apiResponse: []byte(`{"status":"Failed","statusDescription":"Invalid TTL. Choose from the list of the values we support."}`),
			expected: expected{
				Query: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=300`,
				Error: "fail to add TXT record: Failed Invalid TTL. Choose from the list of the values we support.",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.NotNil(t, req.URL.RawQuery)
				assert.Equal(t, test.expected.Query, req.URL.RawQuery)

				handlerMock(http.MethodPost, test.apiResponse).ServeHTTP(rw, req)
			}))

			client, _ := NewClient("myAuthID", "myAuthPassword")
			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			err := client.AddTxtRecord(test.zone.Name, test.authFQDN, test.value, test.ttl)

			if test.expected.Error != "" {
				require.EqualError(t, err, test.expected.Error)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
