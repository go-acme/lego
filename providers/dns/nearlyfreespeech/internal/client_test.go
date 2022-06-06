package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()

	server := httptest.NewServer(mux)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func testHandler(params map[string]string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get(authenticationHeader) == "" {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		for k, v := range params {
			if req.PostForm.Get(k) != v {
				http.Error(rw, fmt.Sprintf("data: got %s want %s", k, v), http.StatusBadRequest)
				return
			}
		}
	}
}

func testErrorHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open("./fixtures/error.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusUnauthorized)

		_, _ = io.Copy(rw, file)
	}
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	params := map[string]string{
		"data": "txtTXTtxt",
		"name": "sub",
		"type": "TXT",
		"ttl":  "30",
	}

	mux.Handle("/dns/example.com/addRR", testHandler(params))

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord("example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.Handle("/dns/example.com/addRR", testErrorHandler())

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord("example.com", record)
	require.Error(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client, mux := setupTest(t)

	params := map[string]string{
		"data": "txtTXTtxt",
		"name": "sub",
		"type": "TXT",
	}

	mux.Handle("/dns/example.com/removeRR", testHandler(params))

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord("example.com", record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.Handle("/dns/example.com/removeRR", testErrorHandler())

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord("example.com", record)
	require.Error(t, err)
}
