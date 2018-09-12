package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handlerMock(method string, response *APIResponse, data interface{}) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			content, err := json.Marshal(APIResponse{
				Success: false,
				Errors: []*APIError{
					{
						Code:       666,
						Message:    fmt.Sprintf("invalid method: got %s want %s", req.Method, method),
						ErrorChain: nil,
					},
				},
			})
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Error(rw, string(content), http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		response.Result = jsonData

		content, err := json.Marshal(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(content)
	})
}

func TestClient_GetHostedZoneID(t *testing.T) {
	type result struct {
		zoneID string
		error  bool
	}

	testCases := []struct {
		desc     string
		fqdn     string
		response *APIResponse
		data     []HostedZone
		expected result
	}{
		{
			desc:     "zone found",
			fqdn:     "_acme-challenge.foo.com.",
			response: &APIResponse{Success: true},
			data: []HostedZone{
				{
					ID:   "A",
					Name: "ZONE_A",
				},
			},
			expected: result{zoneID: "A"},
		},
		{
			desc:     "no many zones",
			fqdn:     "_acme-challenge.foo.com.",
			response: &APIResponse{Success: true},
			data: []HostedZone{
				{
					ID:   "A",
					Name: "ZONE_A",
				},
				{
					ID:   "B",
					Name: "ZONE_B",
				},
			},
			expected: result{error: true},
		},
		{
			desc:     "no zone found",
			fqdn:     "_acme-challenge.foo.com.",
			response: &APIResponse{Success: true},
			expected: result{error: true},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.response, test.data))

			client, _ := NewClient("authEmail", "authKey")
			client.BaseURL = server.URL

			zoneID, err := client.GetHostedZoneID(test.fqdn)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.zoneID, zoneID)
			}
		})
	}
}

func TestClient_FindTxtRecord(t *testing.T) {
	type result struct {
		txtRecord *TxtRecord
		error     bool
	}

	testCases := []struct {
		desc     string
		fqdn     string
		zoneID   string
		response *APIResponse
		data     []TxtRecord
		expected result
	}{
		{
			desc:     "TXT record found",
			fqdn:     "_acme-challenge.foo.com.",
			zoneID:   "ZONE_A",
			response: &APIResponse{Success: true},
			data: []TxtRecord{
				{
					Name:    "_acme-challenge.foo.com",
					Type:    "TXT",
					Content: "txtTXTtxtTXTtxtTXTtxtTXT",
					ID:      "A",
					TTL:     50,
					ZoneID:  "ZONE_A",
				},
			},
			expected: result{
				txtRecord: &TxtRecord{
					Name:    "_acme-challenge.foo.com",
					Type:    "TXT",
					Content: "txtTXTtxtTXTtxtTXTtxtTXT",
					ID:      "A",
					TTL:     50,
					ZoneID:  "ZONE_A",
				},
			},
		},
		{
			desc:     "TXT record not found",
			fqdn:     "_acme-challenge.foo.com.",
			zoneID:   "ZONE_A",
			response: &APIResponse{Success: true},
			expected: result{error: true},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(handlerMock(http.MethodGet, test.response, test.data))

			client, _ := NewClient("authEmail", "authKey")
			client.BaseURL = server.URL

			txtRecord, err := client.FindTxtRecord(test.zoneID, test.fqdn)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.txtRecord, txtRecord)
			}
		})
	}
}
