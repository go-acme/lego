package internal

import (
	"encoding/json"
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

	client := NewClient("clientID", "email@example.com", "secret", 300)
	client.HTTPClient = server.Client()
	client.apiBaseURL, _ = url.Parse(server.URL + "/api")
	client.loginURL, _ = url.Parse(server.URL + "/login")

	return client, mux
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/domain/search", func(rw http.ResponseWriter, req *http.Request) {
		response := SearchResponse{
			Items: []Domain{
				{
					ID:         "A",
					DomainName: "example.com",
				},
			},
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/api/record-txt", func(rw http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/api/domain/A/publish", func(rw http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {
		response := AuthResponse{
			Auth: Auth{
				AccessToken:  "at",
				RefreshToken: "",
			},
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := client.AddRecord(t.Context(), "example.com", "_acme-challenge.example.com", "txt")
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/domain/search", func(rw http.ResponseWriter, req *http.Request) {
		response := SearchResponse{
			Items: []Domain{
				{
					ID:         "A",
					DomainName: "example.com",
				},
			},
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/api/domain/A", func(rw http.ResponseWriter, req *http.Request) {
		response := DomainInfo{
			ID:         "Z",
			DomainName: "example.com",
			LastDomainRecordList: []Record{
				{
					ID:       "R01",
					DomainID: "A",
					Name:     "_acme-challenge.example.com",
					Value:    "txt",
					Type:     "TXT",
				},
			},
			SoaTTL: 300,
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/api/record/R01", func(rw http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/api/domain/A/publish", func(rw http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {
		response := AuthResponse{
			Auth: Auth{
				AccessToken:  "at",
				RefreshToken: "",
			},
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	err = client.DeleteRecord(ctx, "example.com", "_acme-challenge.example.com")
	require.NoError(t, err)
}
