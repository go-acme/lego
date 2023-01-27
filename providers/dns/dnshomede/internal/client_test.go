package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Add(t *testing.T) {
	txtValue := "123456789012"

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerMock(addAction, txtValue))
	server := httptest.NewServer(mux)

	credentials := map[string]string{
		"example.org": "secret",
	}

	client := NewClient(credentials)
	client.HTTPClient = server.Client()
	client.baseURL = server.URL

	err := client.Add("example.org", txtValue)
	require.NoError(t, err)
}

func TestClient_Add_error(t *testing.T) {
	txtValue := "123456789012"

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerMock(addAction, txtValue))
	server := httptest.NewServer(mux)

	credentials := map[string]string{
		"example.com": "secret",
	}

	client := NewClient(credentials)
	client.HTTPClient = server.Client()
	client.baseURL = server.URL

	err := client.Add("example.org", txtValue)
	require.Error(t, err)
}

func TestClient_Remove(t *testing.T) {
	txtValue := "ABCDEFGHIJKL"

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerMock(removeAction, txtValue))
	server := httptest.NewServer(mux)

	credentials := map[string]string{
		"example.org": "secret",
	}

	client := NewClient(credentials)
	client.HTTPClient = server.Client()
	client.baseURL = server.URL

	err := client.Remove("example.org", txtValue)
	require.NoError(t, err)
}

func TestClient_Remove_error(t *testing.T) {
	txtValue := "ABCDEFGHIJKL"

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerMock(removeAction, txtValue))
	server := httptest.NewServer(mux)

	credentials := map[string]string{
		"example.com": "secret",
	}

	client := NewClient(credentials)
	client.HTTPClient = server.Client()
	client.baseURL = server.URL

	err := client.Remove("example.org", txtValue)
	require.Error(t, err)
}

func handlerMock(action, value string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		query := req.URL.Query()

		if query.Get("acme") != action {
			_, _ = rw.Write([]byte("nochg 1234:1234:1234:1234:1234:1234:1234:1234"))
			return
		}

		txtValue := query.Get("txt")
		if len(txtValue) < 12 {
			_, _ = rw.Write([]byte("error - no valid acme txt record"))
			return
		}

		if txtValue != value {
			http.Error(rw, fmt.Sprintf("got: %q, expected: %q", txtValue, value), http.StatusBadRequest)
			return
		}

		_, _ = fmt.Fprintf(rw, "%s %s", successCode, txtValue)
	}
}
