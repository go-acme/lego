package internal // import "github.com/xenolf/lego/providers/dns/cloudns/internal"

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
			desc:        "record found",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "foo.com",
			apiResponse: []byte(`{"1":{"id":"1","type":"TXT","host":"_acme-challenge","record":"txtTXTtxtTXTtxtTXTtxtTXT","failover":"1","ttl":"30","status":1}}`),
			expected: result{
				txtRecord: &TXTRecord{
					ID:       1,
					Type:     "TXT",
					Host:     "_acme-challenge",
					Record:   "txtTXTtxtTXTtxtTXTtxtTXT",
					Failover: 1,
					TTL:      30,
					Status:   1,
				},
			},
		},
		{
			desc:        "record not found",
			authFQDN:    "_acme-challenge.foo.com.",
			zoneName:    "test-zone",
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
	testCases := []struct {
		desc     string
		zone     *Zone
		authFQDN string
		value    string
		ttl      int
		expected string
	}{
		{
			desc: "sub-zone",
			zone: &Zone{
				Name:   "bar.com",
				Type:   "master",
				Zone:   "domain",
				Status: "1",
			},
			authFQDN: "_acme-challenge.foo.bar.com.",
			value:    "txtTXTtxtTXTtxtTXTtxtTXT",
			ttl:      60,
			expected: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge.foo&record=txtTXTtxtTXTtxtTXTtxtTXT&record-type=TXT&ttl=60`,
		},
		{
			desc: "main zone",
			zone: &Zone{
				Name:   "bar.com",
				Type:   "master",
				Zone:   "domain",
				Status: "1",
			},
			authFQDN: "_acme-challenge.bar.com.",
			value:    "TXTtxtTXTtxtTXTtxtTXTtxt",
			ttl:      60,
			expected: `auth-id=myAuthID&auth-password=myAuthPassword&domain-name=bar.com&host=_acme-challenge&record=TXTtxtTXTtxtTXTtxtTXTtxt&record-type=TXT&ttl=60`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.NotNil(t, req.URL.RawQuery)
				assert.Equal(t, test.expected, req.URL.RawQuery)

				handlerMock(http.MethodPost, nil).ServeHTTP(rw, req)
			}))

			client, _ := NewClient("myAuthID", "myAuthPassword")
			mockBaseURL, _ := url.Parse(fmt.Sprintf("%s/", server.URL))
			client.BaseURL = mockBaseURL

			err := client.AddTxtRecord(test.zone.Name, test.authFQDN, test.value, test.ttl)
			require.NoError(t, err)
		})
	}
}
