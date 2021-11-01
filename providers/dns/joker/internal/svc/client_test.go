package svc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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
		}
	})

	client := NewClient("test", "secret")
	client.BaseURL = server.URL

	zone := "example.com"
	label := "_acme-challenge"
	value := "123"

	err := client.Send(zone, label, value)
	require.NoError(t, err)
}

func TestClient_Send_empty(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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
		}
	})

	client := NewClient("test", "secret")
	client.BaseURL = server.URL

	zone := "example.com"
	label := "_acme-challenge"
	value := ""

	err := client.Send(zone, label, value)
	require.NoError(t, err)
}
