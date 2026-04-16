package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := NewClient(server.URL, "test-token")
	return client, server
}

func TestFindZoneByFQDN_ExactMatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Status: "success",
			Data:   ZonesData{Zones: []Zone{{ID: "z1", Name: "example.com"}}},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	id, name, err := client.FindZoneByFQDN(context.Background(), "example.com.")
	require.NoError(t, err)
	assert.Equal(t, "z1", id)
	assert.Equal(t, "example.com", name)
}

func TestFindZoneByFQDN_SubdomainMatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Status: "success",
			Data:   ZonesData{Zones: []Zone{{ID: "z1", Name: "example.com"}}},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	id, _, err := client.FindZoneByFQDN(context.Background(), "_acme-challenge.sub.example.com.")
	require.NoError(t, err)
	assert.Equal(t, "z1", id)
}

func TestFindZoneByFQDN_CaseInsensitive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Status: "success",
			Data:   ZonesData{Zones: []Zone{{ID: "z1", Name: "Example.COM"}}},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	id, _, err := client.FindZoneByFQDN(context.Background(), "example.com.")
	require.NoError(t, err)
	assert.Equal(t, "z1", id)
}

func TestFindZoneByFQDN_NoMatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Status: "success",
			Data:   ZonesData{Zones: []Zone{{ID: "z1", Name: "other.com"}}},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	_, _, err := client.FindZoneByFQDN(context.Background(), "example.com.")
	require.Error(t, err)
}

func TestFindZoneByFQDN_ZoneWithTrailingDot(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Status: "success",
			Data:   ZonesData{Zones: []Zone{{ID: "z1", Name: "example.com."}}},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	id, _, err := client.FindZoneByFQDN(context.Background(), "example.com.")
	require.NoError(t, err)
	assert.Equal(t, "z1", id)
}

func TestFindZoneByFQDN_MostSpecificMatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ZonesResponse{
			Status: "success",
			Data: ZonesData{
				Zones: []Zone{
					{ID: "z-parent", Name: "example.com"},
					{ID: "z-sub", Name: "sub.example.com"},
				},
			},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	id, _, err := client.FindZoneByFQDN(context.Background(), "_acme-challenge.sub.example.com.")
	require.NoError(t, err)
	assert.Equal(t, "z-sub", id)
}

func TestCreateTXTRecord_Success(t *testing.T) {
	mux := http.NewServeMux()

	var received RecordRequest
	mux.HandleFunc("/v1/zones/z1/records", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusCreated)
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	err := client.CreateTXTRecord(context.Background(), "z1", "_acme-challenge.example.com", "token-value", 120)
	require.NoError(t, err)
	assert.Equal(t, "_acme-challenge.example.com", received.Name)
	assert.Equal(t, "TXT", received.Type)
	assert.Equal(t, "token-value", received.Content)
	assert.Equal(t, 120, received.TTL)
}

func TestCreateTXTRecord_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones/z1/records", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(APIError{
			Status: "error",
			Error:  APIErrorDetails{Code: "FORBIDDEN", Message: "insufficient scopes"},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	err := client.CreateTXTRecord(context.Background(), "z1", "test", "val", 120)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FORBIDDEN")
}

func TestDeleteTXTRecord_Success(t *testing.T) {
	mux := http.NewServeMux()

	var receivedContent string
	mux.HandleFunc("/v1/zones/z1/records/by-name/", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		receivedContent = r.URL.Query().Get("content")
		w.WriteHeader(http.StatusNoContent)
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	err := client.DeleteTXTRecord(context.Background(), "z1", "_acme-challenge.example.com", "token-value")
	require.NoError(t, err)
	assert.Equal(t, "token-value", receivedContent)
}

func TestDeleteTXTRecord_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones/z1/records/by-name/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{
			Status: "error",
			Error:  APIErrorDetails{Code: "NOT_FOUND", Message: "record not found"},
		})
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	err := client.DeleteTXTRecord(context.Background(), "z1", "test", "val")
	require.Error(t, err)
}

func TestListZones_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	callCount := 0

	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		offset := r.URL.Query().Get("offset")
		if offset == "" || offset == "0" {
			zones := make([]Zone, 100)
			for i := range zones {
				zones[i] = Zone{ID: "z-filler", Name: "filler.com"}
			}
			json.NewEncoder(w).Encode(ZonesResponse{Status: "success", Data: ZonesData{Zones: zones}})
		} else {
			json.NewEncoder(w).Encode(ZonesResponse{
				Status: "success",
				Data:   ZonesData{Zones: []Zone{{ID: "z-last", Name: "last.com"}}},
			})
		}
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	id, _, err := client.FindZoneByFQDN(context.Background(), "last.com.")
	require.NoError(t, err)
	assert.Equal(t, "z-last", id)
	assert.GreaterOrEqual(t, callCount, 2)
}

func TestListZones_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	_, _, err := client.FindZoneByFQDN(context.Background(), "example.com.")
	require.Error(t, err)
}

func TestListZones_InvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})

	client, server := newTestClient(t, mux)
	defer server.Close()

	_, _, err := client.FindZoneByFQDN(context.Background(), "example.com.")
	require.Error(t, err)
}
