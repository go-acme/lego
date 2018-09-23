package linodev4

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	MockResponseMap map[string]interface{}
)

var (
	apiToken   string
	isTestLive bool
)

func init() {
	apiToken = os.Getenv("LINODE_TOKEN")
	isTestLive = len(apiToken) != 0
}

func restoreEnv() {
	os.Setenv("LINODE_TOKEN", apiToken)
}

func newMockServer(t *testing.T, responses MockResponseMap) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure that we support the requested action.
		action := r.Method + ":" + r.URL.Path
		resp, ok := responses[action]
		if !ok {
			require.FailNowf(t, "Unsupported mock", "action: %s", action)
		}

		rawResponse, err := json.Marshal(resp)
		if err != nil {
			msg := fmt.Sprintf("Failed to JSON encode response: %v", err)
			require.FailNow(t, msg)
		}

		// Send the response.
		w.Header().Set("Content-Type", "application/json")
		if err, ok := resp.(linodego.APIError); ok {
			if err.Errors[0].Reason == "Not found" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write(rawResponse)
	}))

	time.Sleep(100 * time.Millisecond)
	return srv
}

func TestNewDNSProviderWithEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderWithoutEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "linodev4: some credentials information are missing: LINODE_TOKEN")
}

func TestNewDNSProviderCredentialsWithKey(t *testing.T) {
	config := NewDefaultConfig()
	config.Token = "testing"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderCredentialsWithoutKey(t *testing.T) {
	config := NewDefaultConfig()
	config.Token = ""

	_, err := NewDNSProviderConfig(config)
	assert.EqualError(t, err, "linodev4: Linode Access Token missing")
}

func TestDNSProvider_Present(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="

	mockResponses := MockResponseMap{
		"GET:/domains": linodego.DomainsPagedResponse{
			PageOptions: &linodego.PageOptions{
				Pages:   1,
				Results: 1,
				Page:    1,
			},
			Data: []linodego.Domain{{
				Domain: domain,
				ID:     1234,
			}},
		},
		"POST:/domains/1234/records": linodego.DomainRecord{
			ID: 1234,
		},
	}

	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	p.client.SetBaseURL(mockSrv.URL)

	err = p.Present(domain, "", keyAuth)
	assert.NoError(t, err)
}

func TestDNSProvider_PresentNoDomain(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"GET:/domains": linodego.APIError{
			Errors: []linodego.APIErrorReason{{
				Reason: "Not found",
			}},
		},
	}

	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	p.client.SetBaseURL(mockSrv.URL)

	err = p.Present(domain, "", keyAuth)
	assert.EqualError(t, err, "[404] Not found")
}

func TestDNSProvider_PresentCreateFailed(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"GET:/domains": &linodego.DomainsPagedResponse{
			PageOptions: &linodego.PageOptions{
				Pages:   1,
				Results: 1,
				Page:    1,
			},
			Data: []linodego.Domain{{
				Domain: "foobar.com",
				ID:     1234,
			}},
		},
		"POST:/domains/1234/records": linodego.APIError{
			Errors: []linodego.APIErrorReason{{
				Reason: "Failed to create domain resource",
				Field:  "somefield",
			}},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	p.client.SetBaseURL(mockSrv.URL)

	err = p.Present(domain, "", keyAuth)
	assert.EqualError(t, err, "[400] [somefield] Failed to create domain resource")
}

func TestDNSProvider_PresentLive(t *testing.T) {
	if !isTestLive {
		t.Skip("Skipping live test")
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"GET:/domains": &linodego.DomainsPagedResponse{
			PageOptions: &linodego.PageOptions{
				Pages:   1,
				Results: 1,
				Page:    1,
			},
			Data: []linodego.Domain{{
				Domain: "foobar.com",
				ID:     1234,
			}},
		},
		"GET:/domains/1234/records": &linodego.DomainRecordsPagedResponse{
			PageOptions: &linodego.PageOptions{
				Pages:   1,
				Results: 1,
				Page:    1,
			},
			Data: []linodego.DomainRecord{{
				ID:     1234,
				Name:   "_acme-challenge",
				Target: "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
				Type:   "TXT",
			}},
		},
		"DELETE:/domains/1234/records/1234": struct{}{},
	}

	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	p.client.SetBaseURL(mockSrv.URL)

	err = p.CleanUp(domain, "", keyAuth)
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUpNoDomain(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"GET:/domains": linodego.APIError{
			Errors: []linodego.APIErrorReason{{
				Reason: "Not found",
			}},
		},
		"GET:/domains/1234/records": linodego.APIError{
			Errors: []linodego.APIErrorReason{{
				Reason: "Not found",
			}},
		},
	}

	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	p.client.SetBaseURL(mockSrv.URL)

	err = p.CleanUp(domain, "", keyAuth)
	assert.EqualError(t, err, "[404] Not found")
}

func TestDNSProvider_CleanUpDeleteFailed(t *testing.T) {
	defer restoreEnv()
	os.Setenv("LINODE_TOKEN", "testing")

	p, err := NewDNSProvider()
	assert.NoError(t, err)

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="
	mockResponses := MockResponseMap{
		"GET:/domains": linodego.DomainsPagedResponse{
			PageOptions: &linodego.PageOptions{
				Pages:   1,
				Results: 1,
				Page:    1,
			},
			Data: []linodego.Domain{{
				ID:     1234,
				Domain: "example.com",
			}},
		},
		"GET:/domains/1234/records": linodego.DomainRecordsPagedResponse{
			PageOptions: &linodego.PageOptions{
				Pages:   1,
				Results: 1,
				Page:    1,
			},
			Data: []linodego.DomainRecord{{
				ID:     1234,
				Name:   "_acme-challenge",
				Target: "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
				Type:   "TXT",
			}},
		},
		"DELETE:/domains/1234/records/1234": linodego.APIError{
			Errors: []linodego.APIErrorReason{{
				Reason: "Failed to delete domain resource",
			}},
		},
	}

	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	p.client.SetBaseURL(mockSrv.URL)

	err = p.CleanUp(domain, "", keyAuth)
	assert.EqualError(t, err, "[400] Failed to delete domain resource")
}
