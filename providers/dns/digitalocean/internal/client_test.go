package internal

import (
	"bytes"
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

func setupTest(t *testing.T, pattern string, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
	client.BaseURL, _ = url.Parse(server.URL)

	mux.HandleFunc(pattern, handler)

	return client
}

func checkHeader(req *http.Request, name, value string) error {
	val := req.Header.Get(name)
	if val != value {
		return fmt.Errorf("invalid header value, got: %s want %s", val, value)
	}
	return nil
}

func writeFixture(rw http.ResponseWriter, filename string) {
	file, err := os.Open(filepath.Join("fixtures", filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	_, _ = io.Copy(rw, file)
}

func TestClient_AddTxtRecord(t *testing.T) {
	client := setupTest(t, "/v2/domains/example.com/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		err := checkHeader(req, "Accept", "application/json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = checkHeader(req, "Content-Type", "application/json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = checkHeader(req, "Authorization", "Bearer secret")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expectedReqBody := `{"type":"TXT","name":"_acme-challenge.example.com.","data":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":30}`
		if expectedReqBody != string(bytes.TrimSpace(reqBody)) {
			http.Error(rw, fmt.Sprintf("unexpected request body: %s", string(bytes.TrimSpace(reqBody))), http.StatusBadRequest)
			return
		}

		rw.WriteHeader(http.StatusCreated)
		writeFixture(rw, "domains-records_POST.json")
	})

	record := Record{
		Type: "TXT",
		Name: "_acme-challenge.example.com.",
		Data: "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		TTL:  30,
	}

	newRecord, err := client.AddTxtRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &TxtRecordResponse{DomainRecord: Record{
		ID:   1234567,
		Type: "TXT",
		Name: "_acme-challenge",
		Data: "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		TTL:  0,
	}}

	assert.Equal(t, expected, newRecord)
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	client := setupTest(t, "/v2/domains/example.com/records/1234567", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		err := checkHeader(req, "Accept", "application/json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = checkHeader(req, "Authorization", "Bearer secret")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.RemoveTxtRecord(t.Context(), "example.com", 1234567)
	require.NoError(t, err)
}
