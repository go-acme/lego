package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("secret")
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/example.com/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		if req.Header.Get("Authorization") != "secret" {
			http.Error(rw, `{"message":"Unauthenticated"}`, http.StatusUnauthorized)
			return
		}

		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		expectedReqBody := `{"name":"_acme-challenge.example.com","type":"TXT","content":"\"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI\"","ttl":120}`
		if string(reqBody) != expectedReqBody {
			http.Error(rw, `{"message":"invalid request"}`, http.StatusBadRequest)
			return
		}

		resp := `{
				"data": {
					"id": 1234567
				},
				"meta": {
					"location": "https://api.ukfast.io/safedns/v1/zones/example.com/records/1234567"
				}
			}`

		rw.WriteHeader(http.StatusCreated)
		_, err = fmt.Fprint(rw, resp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: `"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"`,
		TTL:     dns01.DefaultTTL,
	}

	response, err := client.AddRecord("example.com", record)
	require.NoError(t, err)

	expected := &AddRecordResponse{
		Data: struct {
			ID int `json:"id"`
		}{
			ID: 1234567,
		},
		Meta: struct {
			Location string `json:"location"`
		}{
			Location: "https://api.ukfast.io/safedns/v1/zones/example.com/records/1234567",
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_RemoveRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/example.com/records/1234567", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		if req.Header.Get("Authorization") != "secret" {
			http.Error(rw, `{"message":"Unauthenticated"}`, http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.RemoveRecord("example.com", 1234567)
	require.NoError(t, err)
}
