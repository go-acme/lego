package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, handler http.Handler) (*httptest.Server, *Client) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })

	client, err := NewClient("token", "secret", 60)
	require.NoError(t, err)

	// Override API endpoint
	client.apiEndpoint = server.URL
	client.httpClient = server.Client()
	client.httpClient.Timeout = 2 * time.Second

	return server, client
}

func Test_NewClient(t *testing.T) {
	testCases := []struct {
		desc   string
		token  string
		secret string
		ttl    int
		error  bool
	}{
		{
			desc:   "success",
			token:  "token",
			secret: "secret",
			ttl:    60,
			error:  false,
		},
		{
			desc:   "missing token",
			token:  "",
			secret: "secret",
			ttl:    60,
			error:  true,
		},
		{
			desc:   "missing secret",
			token:  "token",
			secret: "",
			ttl:    60,
			error:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client, err := NewClient(test.token, test.secret, test.ttl)

			if test.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				require.Equal(t, test.token, client.token)
				require.Equal(t, test.secret, client.secret)
				require.Equal(t, test.ttl, client.ttl)
			}
		})
	}
}

func TestClient_GetZoneByDomain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/commonserviceitem", func(w http.ResponseWriter, _ *http.Request) {
		fixedResponse := `{
			"is1a": {
				"Success": true,
				"Response": {
					"DNS": [
						{
							"ID": "123456789012",
							"Name": "example.com"
						}
					]
				}
			}
		}`
		_, _ = w.Write([]byte(fixedResponse))
	})

	_, client := setupTest(t, mux)

	zone, err := client.GetZoneByDomain(context.Background(), "example.com")
	require.NoError(t, err)
	require.NotNil(t, zone)
	assert.Equal(t, "123456789012", zone.ID)
	assert.Equal(t, "example.com", zone.Name)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/commonserviceitem/123456789012/dnsrecord", func(w http.ResponseWriter, _ *http.Request) {
		response := `{
			"is1a": {
				"Success": true
			}
		}`
		_, _ = w.Write([]byte(response))
	})

	_, client := setupTest(t, mux)

	err := client.CreateTXTRecord(context.Background(), "123456789012", "_acme-challenge", "token-value")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	mux := http.NewServeMux()

	// For getting records
	mux.HandleFunc("/commonserviceitem/123456789012/dnsrecord", func(w http.ResponseWriter, _ *http.Request) {
		response := `{
			"is1a": {
				"Success": true,
				"Response": {
					"DNSRecords": [
						{
							"ID": "100",
							"Name": "_acme-challenge",
							"Type": "TXT",
							"RData": "\"token-value\"",
							"TTL": 60
						}
					]
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	})

	// For deleting a record
	mux.HandleFunc("/commonserviceitem/123456789012/dnsrecord/100", func(w http.ResponseWriter, _ *http.Request) {
		response := `{
			"is1a": {
				"Success": true
			}
		}`
		_, _ = w.Write([]byte(response))
	})

	_, client := setupTest(t, mux)

	err := client.DeleteTXTRecord(context.Background(), "123456789012", "_acme-challenge", "token-value")
	require.NoError(t, err)
}
