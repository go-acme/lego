package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nrdcg/goacmedns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern, filename string, statusCode int) *HTTPStorage {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if filename == "" {
			rw.WriteHeader(statusCode)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(statusCode)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	storage, err := NewHTTPStorage(server.URL)
	require.NoError(t, err)

	storage.client = server.Client()

	return storage
}

func TestHTTPStorage_Fetch(t *testing.T) {
	storage := setupTest(t, "GET /example.com", "fetch.json", http.StatusOK)

	account, err := storage.Fetch(t.Context(), "example.com")
	require.NoError(t, err)

	expected := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	assert.Equal(t, expected, account)
}

func TestHTTPStorage_Fetch_error(t *testing.T) {
	storage := setupTest(t, "GET /example.com", "error.json", http.StatusInternalServerError)

	_, err := storage.Fetch(t.Context(), "example.com")
	require.Error(t, err)
}

func TestHTTPStorage_FetchAll(t *testing.T) {
	storage := setupTest(t, "GET /", "fetch-all.json", http.StatusOK)

	account, err := storage.FetchAll(t.Context())
	require.NoError(t, err)

	expected := map[string]goacmedns.Account{
		"a": {
			FullDomain: "foo.example.com",
			SubDomain:  "foo",
			Username:   "user",
			Password:   "secret",
			ServerURL:  "https://example.com",
		},
		"b": {
			FullDomain: "bar.example.com",
			SubDomain:  "bar",
			Username:   "user",
			Password:   "secret",
			ServerURL:  "https://example.com",
		},
	}

	assert.Equal(t, expected, account)
}

func TestHTTPStorage_FetchAll_error(t *testing.T) {
	storage := setupTest(t, "GET /", "error.json", http.StatusInternalServerError)

	_, err := storage.FetchAll(t.Context())
	require.Error(t, err)
}

func TestHTTPStorage_Put(t *testing.T) {
	storage := setupTest(t, "POST /example.com", "", http.StatusOK)

	account := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	err := storage.Put(t.Context(), "example.com", account)
	require.NoError(t, err)
}

func TestHTTPStorage_Put_error(t *testing.T) {
	storage := setupTest(t, "POST /example.com", "error.json", http.StatusInternalServerError)

	account := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	err := storage.Put(t.Context(), "example.com", account)
	require.Error(t, err)
}

func TestHTTPStorage_Put_CNAME_created(t *testing.T) {
	storage := setupTest(t, "POST /example.com", "", http.StatusCreated)

	account := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	err := storage.Put(t.Context(), "example.com", account)
	require.ErrorIs(t, err, ErrCNAMEAlreadyCreated)
}
