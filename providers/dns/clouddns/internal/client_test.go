package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_AddRecord(t *testing.T) {
	mux := http.NewServeMux()

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

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("clientID", "email@example.com", "secret", 300)
	client.apiBaseURL = server.URL + "/api"
	client.loginURL = server.URL + "/login"

	err := client.AddRecord("example.com", "_acme-challenge.example.com", "txt")
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	mux := http.NewServeMux()

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

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("clientID", "email@example.com", "secret", 300)
	client.apiBaseURL = server.URL + "/api"
	client.loginURL = server.URL + "/login"

	err := client.DeleteRecord("example.com", "_acme-challenge.example.com")
	require.NoError(t, err)
}
