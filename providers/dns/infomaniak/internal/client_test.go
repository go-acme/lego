package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return New(server.URL, "token"), mux
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/1/domain/666/dns/record", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Authorization") != "Bearer token" {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = req.Body.Close() }()

		if string(raw) != `{"source":"foo","type":"TXT","ttl":60,"target":"txtxtxttxt"}` {
			http.Error(rw, fmt.Sprintf("invalid request body: %s", string(raw)), http.StatusBadRequest)
			return
		}

		response := `{"result":"success","data": "123"}`

		_, err = rw.Write([]byte(response))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})

	domain := &DNSDomain{
		ID:           666,
		CustomerName: "test",
	}

	record := Record{
		Source: "foo",
		Target: "txtxtxttxt",
		Type:   "TXT",
		TTL:    60,
	}

	recordID, err := client.CreateDNSRecord(domain, record)
	require.NoError(t, err)

	assert.Equal(t, "123", recordID)
}

func TestClient_GetDomainByName(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/1/product", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Authorization") != "Bearer token" {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		serviceName := req.URL.Query().Get("service_name")
		if serviceName != "domain" {
			http.Error(rw, fmt.Sprintf("invalid service_name: %s", serviceName), http.StatusBadRequest)
			return
		}

		customerName := req.URL.Query().Get("customer_name")
		fmt.Println("customerName", customerName)
		if customerName == "" {
			http.Error(rw, fmt.Sprintf("invalid customer_name: %s", customerName), http.StatusBadRequest)
			return
		}

		response := `
{
  "result": "success",
  "data": [
    {
      "id": 123,
      "customer_name": "two.three.example.com"
    },
    {
      "id": 456,
      "customer_name": "three.example.com"
    }
  ]
}
`

		_, err := rw.Write([]byte(response))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})

	domain, err := client.GetDomainByName("one.two.three.example.com.")
	require.NoError(t, err)

	expected := &DNSDomain{ID: 123, CustomerName: "two.three.example.com"}
	assert.Equal(t, expected, domain)
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/1/domain/123/dns/record/456", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Authorization") != "Bearer token" {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		_, err := rw.Write([]byte((`{"result":"success"}`)))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})

	err := client.DeleteDNSRecord(123, "456")
	require.NoError(t, err)
}
