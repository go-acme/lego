package internal

import (
	"encoding/json"
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

func newJSONErrorf(reason string, a ...any) string {
	err := APIError{
		Message: "Cannot View Dns Record",
		Result:  fmt.Sprintf(reason, a...),
	}

	data, _ := json.Marshal(err)

	return string(data)
}

func testHandler(kv map[string]string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		domain := req.URL.Query().Get("domain")
		if domain != "example.com" {
			http.Error(rw, newJSONErrorf("invalid domain: %s", domain), http.StatusUnauthorized)
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
				http.Error(rw, newJSONErrorf("invalid %q: %s", k, actual), http.StatusBadRequest)
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

	err := client.SetRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_SetRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/CMD_API_DNS_CONTROL", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, newJSONErrorf("OOPS"), http.StatusInternalServerError)
	})

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
		TTL:   123,
	}

	err := client.SetRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "[status code 500] Cannot View Dns Record: OOPS")
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

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/CMD_API_DNS_CONTROL", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, newJSONErrorf("OOPS"), http.StatusInternalServerError)
	})

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "[status code 500] Cannot View Dns Record: OOPS")
}
