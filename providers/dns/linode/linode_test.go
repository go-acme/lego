package linode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewasted/linode"
	"github.com/timewasted/linode/dns"
)

type (
	LinodeResponse struct {
		Action string                 `json:"ACTION"`
		Data   interface{}            `json:"DATA"`
		Errors []linode.ResponseError `json:"ERRORARRAY"`
	}
	MockResponse struct {
		Response interface{}
		Errors   []linode.ResponseError
	}
	MockResponseMap map[string]MockResponse
)

var (
	apiKey     string
	isTestLive bool
)

func init() {
	apiKey = os.Getenv("LINODE_API_KEY")
	isTestLive = len(apiKey) != 0
}

func restoreEnv() {
	os.Setenv("LINODE_API_KEY", apiKey)
}

func newMockServer(t *testing.T, responses MockResponseMap) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure that we support the requested action.
		action := r.URL.Query().Get("api_action")
		resp, ok := responses[action]
		if !ok {
			require.FailNowf(t, "Unsupported mock action: %q", action)
		}

		// Build the response that the server will return.
		linodeResponse := LinodeResponse{
			Action: action,
			Data:   resp.Response,
			Errors: resp.Errors,
		}
		rawResponse, err := json.Marshal(linodeResponse)
		if err != nil {
			msg := fmt.Sprintf("Failed to JSON encode response: %v", err)
			require.FailNow(t, msg)
		}

		// Send the response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(rawResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	time.Sleep(100 * time.Millisecond)
	return srv
}

func TestNewDNSProviderWithEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_API_KEY", "testing")

	_, err := NewDNSProvider()
	require.NoError(t, err)
}

func TestNewDNSProviderWithoutEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "linode: some credentials information are missing: LINODE_API_KEY")
}

func TestNewDNSProviderWithKey(t *testing.T) {
	config := NewDefaultConfig()
	config.APIKey = "testing"

	_, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
}

func TestNewDNSProviderWithoutKey(t *testing.T) {
	config := NewDefaultConfig()

	_, err := NewDNSProviderConfig(config)
	assert.EqualError(t, err, "linode: credentials missing")
}

func TestDNSProvider_Present(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_API_KEY", "testing")

	p, err := NewDNSProvider()
	require.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="

	testCases := []struct {
		desc          string
		mockResponses MockResponseMap
		expectedError string
	}{
		{
			desc: "success",
			mockResponses: MockResponseMap{
				"domain.list": MockResponse{
					Response: []dns.Domain{
						{
							Domain:   domain,
							DomainID: 1234,
						},
					},
				},
				"domain.resource.create": MockResponse{
					Response: dns.ResourceResponse{
						ResourceID: 1234,
					},
				},
			},
		},
		{
			desc: "NoDomain",
			mockResponses: MockResponseMap{
				"domain.list": MockResponse{
					Response: []dns.Domain{{
						Domain:   "foobar.com",
						DomainID: 1234,
					}},
				},
			},
			expectedError: "dns: requested domain not found",
		},
		{
			desc: "CreateFailed",
			mockResponses: MockResponseMap{
				"domain.list": MockResponse{
					Response: []dns.Domain{
						{
							Domain:   domain,
							DomainID: 1234,
						},
					},
				},
				"domain.resource.create": MockResponse{
					Response: nil,
					Errors: []linode.ResponseError{
						{
							Code:    1234,
							Message: "Failed to create domain resource",
						},
					},
				},
			},
			expectedError: "Failed to create domain resource",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			mockSrv := newMockServer(t, test.mockResponses)
			defer mockSrv.Close()

			p.client.ToLinode().SetEndpoint(mockSrv.URL)

			err = p.Present(domain, "", keyAuth)
			if len(test.expectedError) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_PresentLive(t *testing.T) {
	if !isTestLive {
		t.Skip("Skipping live test")
	}
	// TODO implement this test
}

func TestDNSProvider_CleanUp(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_API_KEY", "testing")

	p, err := NewDNSProvider()
	require.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="

	testCases := []struct {
		desc          string
		mockResponses MockResponseMap
		expectedError string
	}{
		{
			desc: "success",
			mockResponses: MockResponseMap{
				"domain.list": MockResponse{
					Response: []dns.Domain{
						{
							Domain:   domain,
							DomainID: 1234,
						},
					},
				},
				"domain.resource.list": MockResponse{
					Response: []dns.Resource{
						{
							DomainID:   1234,
							Name:       "_acme-challenge",
							ResourceID: 1234,
							Target:     "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
							Type:       "TXT",
						},
					},
				},
				"domain.resource.delete": MockResponse{
					Response: dns.ResourceResponse{
						ResourceID: 1234,
					},
				},
			},
		},
		{
			desc: "NoDomain",
			mockResponses: MockResponseMap{
				"domain.list": MockResponse{
					Response: []dns.Domain{
						{
							Domain:   "foobar.com",
							DomainID: 1234,
						},
					},
				},
			},
			expectedError: "dns: requested domain not found",
		},
		{
			desc: "DeleteFailed",
			mockResponses: MockResponseMap{
				"domain.list": MockResponse{
					Response: []dns.Domain{
						{
							Domain:   domain,
							DomainID: 1234,
						},
					},
				},
				"domain.resource.list": MockResponse{
					Response: []dns.Resource{
						{
							DomainID:   1234,
							Name:       "_acme-challenge",
							ResourceID: 1234,
							Target:     "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
							Type:       "TXT",
						},
					},
				},
				"domain.resource.delete": MockResponse{
					Response: nil,
					Errors: []linode.ResponseError{
						{
							Code:    1234,
							Message: "Failed to delete domain resource",
						},
					},
				},
			},
			expectedError: "Failed to delete domain resource",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mockSrv := newMockServer(t, test.mockResponses)
			defer mockSrv.Close()

			p.client.ToLinode().SetEndpoint(mockSrv.URL)

			err = p.CleanUp(domain, "", keyAuth)
			if len(test.expectedError) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}
