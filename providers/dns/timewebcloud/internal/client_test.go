package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func TestClient_CreateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains/example.com/dns-records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid authentication token: %s", auth), http.StatusUnauthorized)
			return
		}

		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		expectedReqBody := `{"type":"TXT","value":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","subdomain":"_acme-challenge"}`
		assert.Equal(t, expectedReqBody, string(bytes.TrimSpace(reqBody)))

		rw.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(rw, `{
		  "dns_record": {
			"type": "TXT",
			"id": 123,
			"data": {
				"priority": 0,
				"subdomain": "example.com",
				"value": "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"
			}
		  },
		  "response_id": "15095f25-aac3-4d60-a788-96cb5136f186"
		}`)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	payload := CreateRecordPayload{
		Type:      "TXT",
		Value:     "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		SubDomain: "_acme-challenge",
	}

	response, err := client.CreateRecord(context.Background(), "example.com.", payload)
	require.NoError(t, err)

	expectedResponse := &CreateRecordResponse{
		DNSRecord: DNSRecord{
			Type: "TXT",
			ID:   123,
		},
	}
	assert.Equal(t, expectedResponse, response)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains/example.com/dns-records/123", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid authentication token: %s", auth), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteRecord(context.Background(), "example.com.", 123)
	require.NoError(t, err)
}
