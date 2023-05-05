package svc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("test", "secret")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	return client, mux
}

func TestClient_Send(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		all, _ := io.ReadAll(req.Body)

		if string(all) != "label=_acme-challenge&password=secret&type=TXT&username=test&value=123&zone=example.com" {
			http.Error(rw, fmt.Sprintf("invalid request: %q", string(all)), http.StatusBadRequest)
			return
		}

		_, err := rw.Write([]byte("OK: 1 inserted, 0 deleted"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	zone := "example.com"
	label := "_acme-challenge"
	value := "123"

	err := client.SendRequest(context.Background(), zone, label, value)
	require.NoError(t, err)
}

func TestClient_Send_empty(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		all, _ := io.ReadAll(req.Body)

		if string(all) != "label=_acme-challenge&password=secret&type=TXT&username=test&value=&zone=example.com" {
			http.Error(rw, fmt.Sprintf("invalid request: %q", string(all)), http.StatusBadRequest)
			return
		}

		_, err := rw.Write([]byte("OK: 1 inserted, 0 deleted"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	zone := "example.com"
	label := "_acme-challenge"
	value := ""

	err := client.SendRequest(context.Background(), zone, label, value)
	require.NoError(t, err)
}
