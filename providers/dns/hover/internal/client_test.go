package internal

import (
	"context"
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

func setupTest(t *testing.T, pattern string, status int, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if status == http.StatusOK {
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    "Hello",
				Path:     req.URL.Path,
				MaxAge:   3600,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}

			http.SetCookie(rw, cookie)
		}

		if filename == "" {
			rw.WriteHeader(status)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(status)

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_GetDomains(t *testing.T) {
	client := setupTest(t, "GET /domains", http.StatusOK, "domains.json")

	domains, err := client.GetDomains(context.Background())
	require.NoError(t, err)

	expected := []Domain{
		{
			ID:         "b04df6be-0151-48e5-b07a-09f706514070",
			DomainName: "example.com",
		},
		{
			ID:         "1c833aa3-bd63-413f-aee2-ea73ab9fa8f1",
			DomainName: "example.org",
		},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_GetDomains_error(t *testing.T) {
	client := setupTest(t, "GET /domains", http.StatusUnauthorized, "error.json")

	_, err := client.GetDomains(context.Background())
	require.Error(t, err)

	require.EqualError(t, err, "401: bad_login: Unknown username or password")
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "GET /domains/b04df6be-0151-48e5-b07a-09f706514070/dns/", http.StatusOK, "records.json")

	records, err := client.GetRecords(context.Background(), "b04df6be-0151-48e5-b07a-09f706514070")
	require.NoError(t, err)

	expected := []Record{
		{ID: "39ea9822-2aac-41df-b862-90131a0b1bf9", Default: true, Name: "Pear", TTL: 60, Type: "A", Content: "qHRcY"},
		{ID: "9f996fa7-b156-44fe-b3f7-c809fb1d627e", Name: "Eggplant", TTL: 120, Type: "A", Content: "XtoGQzK"},
		{ID: "8eb26180-2602-4e37-bdef-aed6927d771e", Name: "Apple", TTL: 360, Type: "TXT", Content: "ZfYFjfqBc"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := setupTest(t, "GET /domains/example.com/dns/", http.StatusUnauthorized, "error.json")

	_, err := client.GetRecords(context.Background(), "example.com")
	require.Error(t, err)

	require.EqualError(t, err, "401: bad_login: Unknown username or password")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "DELETE /domains/123/dns/456", http.StatusOK, "")

	err := client.DeleteRecord(context.Background(), "123", "456")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "DELETE /domains/123/dns/456", http.StatusUnauthorized, "error.json")

	err := client.DeleteRecord(context.Background(), "123", "456")
	require.Error(t, err)

	require.EqualError(t, err, "401: bad_login: Unknown username or password")
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := setupTest(t, "POST /domains/123/dns", http.StatusOK, "")

	err := client.AddTXTRecord(context.Background(), "123", "foo.example.com", "txt")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := setupTest(t, "POST /domains/123/dns", http.StatusUnauthorized, "error.json")

	err := client.AddTXTRecord(context.Background(), "123", "foo.example.com", "txt")
	require.Error(t, err)

	require.EqualError(t, err, "401: bad_login: Unknown username or password")
}
