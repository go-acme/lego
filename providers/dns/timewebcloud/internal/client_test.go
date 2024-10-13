package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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

func checkAuthorizationHeader(req *http.Request) error {
	val := req.Header.Get("Authorization")
	if val != "Bearer secret" {
		return fmt.Errorf("invalid header value, got: %s want %s", val, "Bearer secret")
	}
	return nil
}

func writeResponse(rw http.ResponseWriter, statusCode int, filename string) error {
	file, err := os.Open(filepath.Join("fixtures", filename))
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	rw.WriteHeader(statusCode)

	_, err = io.Copy(rw, file)
	if err != nil {
		return err
	}

	return nil
}

func TestClient_CreateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("POST /v1/domains/example.com/dns-records", func(rw http.ResponseWriter, req *http.Request) {
		err := checkAuthorizationHeader(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		content, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if string(bytes.TrimSpace(content)) != `{"type":"TXT","value":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","subdomain":"_acme-challenge"}` {
			http.Error(rw, "invalid request body: "+string(content), http.StatusBadRequest)
			return
		}

		err = writeResponse(rw, http.StatusOK, "createDomainDNSRecord.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	payload := DNSRecord{
		Type:      "TXT",
		Value:     "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		SubDomain: "_acme-challenge",
	}

	response, err := client.CreateRecord(context.Background(), "example.com.", payload)
	require.NoError(t, err)

	expected := &DNSRecord{
		Type: "TXT",
		ID:   123,
	}

	assert.Equal(t, expected, response)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("POST /v1/domains/example.com/dns-records", func(rw http.ResponseWriter, _ *http.Request) {
		err := writeResponse(rw, http.StatusBadRequest, "error_bad_request.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	_, err := client.CreateRecord(context.Background(), "example.com.", DNSRecord{})
	require.Error(t, err)

	assert.EqualError(t, err, "400: Value must be a number conforming to the specified constraints (bad_request) [15095f25-aac3-4d60-a788-96cb5136f186]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("DELETE /v1/domains/example.com/dns-records/123", func(rw http.ResponseWriter, req *http.Request) {
		err := checkAuthorizationHeader(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteRecord(context.Background(), "example.com.", 123)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("DELETE /v1/domains/example.com/dns-records/123", func(rw http.ResponseWriter, _ *http.Request) {
		err := writeResponse(rw, http.StatusBadRequest, "error_unauthorized.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := client.DeleteRecord(context.Background(), "example.com.", 123)
	require.Error(t, err)

	assert.EqualError(t, err, "401: Unauthorized (unauthorized) [15095f25-aac3-4d60-a788-96cb5136f186]")
}
