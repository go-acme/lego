package internal

import (
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

func setupTest(t *testing.T, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(serverURL, "user", "secret")
	client.HTTPClient = server.Client()

	mux.HandleFunc("/enterprise/control/agent.php", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		login := req.Header.Get("Http_auth_login")
		if login != "user" {
			http.Error(rw, fmt.Sprintf("invalid login: %s", login), http.StatusUnauthorized)
			return
		}

		password := req.Header.Get("Http_auth_passwd")
		if password != "secret" {
			http.Error(rw, fmt.Sprintf("invalid password: %s", password), http.StatusUnauthorized)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return client
}

func TestClient_GetSite(t *testing.T) {
	client := setupTest(t, "get-site.xml")

	siteID, err := client.GetSite(context.Background(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 82, siteID)
}

func TestClient_GetSite_error(t *testing.T) {
	client := setupTest(t, "get-site-error.xml")

	siteID, err := client.GetSite(context.Background(), "example.com")
	require.Error(t, err)

	assert.Equal(t, 0, siteID)
}

func TestClient_GetSite_system_error(t *testing.T) {
	client := setupTest(t, "global-error.xml")

	siteID, err := client.GetSite(context.Background(), "example.com")
	require.Error(t, err)

	assert.Equal(t, 0, siteID)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "add-record.xml")

	recordID, err := client.AddRecord(context.Background(), 123, "_acme-challenge.example.com", "txtTXTtxt")
	require.NoError(t, err)

	assert.Equal(t, 4537, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "add-record-error.xml")

	recordID, err := client.AddRecord(context.Background(), 123, "_acme-challenge.example.com", "txtTXTtxt")
	require.ErrorAs(t, err, new(RecResult))

	assert.Equal(t, 0, recordID)
}

func TestClient_AddRecord_system_error(t *testing.T) {
	client := setupTest(t, "global-error.xml")

	recordID, err := client.AddRecord(context.Background(), 123, "_acme-challenge.example.com", "txtTXTtxt")
	require.ErrorAs(t, err, new(*System))

	assert.Equal(t, 0, recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "delete-record.xml")

	recordID, err := client.DeleteRecord(context.Background(), 4537)
	require.NoError(t, err)

	assert.Equal(t, 4537, recordID)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "delete-record-error.xml")

	recordID, err := client.DeleteRecord(context.Background(), 4537)
	require.ErrorAs(t, err, new(RecResult))

	assert.Equal(t, 0, recordID)
}

func TestClient_DeleteRecord_system_error(t *testing.T) {
	client := setupTest(t, "global-error.xml")

	recordID, err := client.DeleteRecord(context.Background(), 4537)
	require.ErrorAs(t, err, new(*System))

	assert.Equal(t, 0, recordID)
}
