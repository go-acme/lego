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

	client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"), "123")
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func TestClient_CreateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v2/domains/example.com/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid API token: %s", auth), http.StatusUnauthorized)
			return
		}

		teamID := req.URL.Query().Get("teamId")
		if teamID != "123" {
			http.Error(rw, fmt.Sprintf("invalid team ID: %s", teamID), http.StatusUnauthorized)
			return
		}

		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		expectedReqBody := `{"name":"_acme-challenge.example.com.","type":"TXT","value":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":60}`
		assert.Equal(t, expectedReqBody, string(bytes.TrimSpace(reqBody)))

		rw.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(rw, `{
			"uid": "9e2eab60-0ba5-4dff-b481-2999c9764b84",
			"updated": 1
		}`)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Name:  "_acme-challenge.example.com.",
		Type:  "TXT",
		Value: "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		TTL:   60,
	}

	resp, err := client.CreateRecord(context.Background(), "example.com.", record)
	require.NoError(t, err)

	expected := &CreateRecordResponse{
		UID:     "9e2eab60-0ba5-4dff-b481-2999c9764b84",
		Updated: 1,
	}

	assert.Equal(t, expected, resp)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v2/domains/example.com/records/1234567", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}
		auth := req.Header.Get("Authorization")
		if auth != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid API token: %s", auth), http.StatusUnauthorized)
			return
		}

		teamID := req.URL.Query().Get("teamId")
		if teamID != "123" {
			http.Error(rw, fmt.Sprintf("invalid team ID: %s", teamID), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusOK)
	})

	err := client.DeleteRecord(context.Background(), "example.com.", "1234567")
	require.NoError(t, err)
}
