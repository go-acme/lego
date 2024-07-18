package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, _ := NewClient(server.URL, "user", "secret")
	client.HTTPClient = server.Client()

	return client, mux
}

func testHandler(kv map[string]string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		domain := req.URL.Query().Get("domain")
		if domain != "example.com" {
			http.Error(rw, fmt.Sprintf("invalid domain: %s", domain), http.StatusUnauthorized)
			return
		}

		data, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		values, err := url.ParseQuery(string(data))
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		for k, v := range kv {
			actual := values.Get(k)
			if v != actual {
				http.Error(rw, fmt.Sprintf("invalid %q: %s", k, actual), http.StatusBadRequest)
				return
			}
		}
	}
}

func TestClient_SetRecord(t *testing.T) {
	client, mux := setupTest(t)

	kv := map[string]string{
		"action": "add",
		"name":   "foo",
		"type":   "TXT",
		"value":  "txtTXTtxt",
		"ttl":    "123",
	}

	mux.HandleFunc("/CMD_API_DNS_CONTROL", testHandler(kv))

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
		TTL:   123,
	}

	err := client.SetRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_SetRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/CMD_API_DNS_CONTROL", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
		TTL:   123,
	}

	err := client.SetRecord(context.Background(), "example.com", record)
	require.EqualError(t, err, "error: 500: Internal Server Error\n")
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	kv := map[string]string{
		"action": "delete",
		"name":   "foo",
		"type":   "TXT",
		"value":  "txtTXTtxt",
		"ttl":    "",
	}

	mux.HandleFunc("/CMD_API_DNS_CONTROL", testHandler(kv))

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
	}

	err := client.DeleteRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/CMD_API_DNS_CONTROL", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
	}

	err := client.DeleteRecord(context.Background(), "example.com", record)
	require.EqualError(t, err, "error: 500: Internal Server Error\n")
}
