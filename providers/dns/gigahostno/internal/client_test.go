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

func setupTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("user@example.com", "password", "")
	require.NoError(t, err)

	client.baseURL = server.URL

	return client
}

func TestAuthenticate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/authenticate" {
			http.NotFound(w, r)
			return
		}

		assert.Contains(t, r.Header.Get("User-Agent"), "goacme-lego")

		response := AuthResponse{
			Meta: Meta{
				Status:        200,
				StatusMessage: "200 OK",
			},
			Data: AuthData{
				Token:       "test-token-12345",
				TokenExpire: 1735234567,
				CustomerID:  "123",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client := setupTestClient(t, handler)

	token, err := client.Authenticate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "test-token-12345", token.Token)
	assert.False(t, token.Deadline.IsZero())
}

func TestListZones(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/zones" {
			http.NotFound(w, r)
			return
		}

		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := ZonesResponse{
			Meta: Meta{
				Status:        200,
				StatusMessage: "200 OK",
			},
			Data: []Zone{
				{
					ID:          "123",
					Name:        "example.com",
					Type:        "NATIVE",
					Active:      "1",
					Status:      "active",
					ExpiryDate:  "2025-12-31 23:59:59",
					RecordCount: 10,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client := setupTestClient(t, handler)

	zones, err := client.ListZones(context.Background(), "test-token")
	require.NoError(t, err)
	require.Len(t, zones, 1)
	assert.Equal(t, "example.com", zones[0].Name)
}

func TestCreateRecord(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/zones/123/records" {
			http.NotFound(w, r)
			return
		}

		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodPost, r.Method)

		var record CreateRecordRequest

		err := json.NewDecoder(r.Body).Decode(&record)
		require.NoError(t, err)

		assert.Equal(t, "_acme-challenge", record.Name)
		assert.Equal(t, "TXT", record.Type)

		w.WriteHeader(http.StatusCreated)
	})

	client := setupTestClient(t, handler)

	record := CreateRecordRequest{
		Name:  "_acme-challenge",
		Type:  "TXT",
		Value: "validation-value",
		TTL:   60,
	}

	err := client.CreateRecord(context.Background(), "test-token", "123", record)
	require.NoError(t, err)
}

func TestDeleteRecord(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/dns/zones/123/records/rec123"
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
		}

		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "_acme-challenge", r.URL.Query().Get("name"))
		assert.Equal(t, "TXT", r.URL.Query().Get("type"))

		w.WriteHeader(http.StatusOK)
	})

	client := setupTestClient(t, handler)

	err := client.DeleteRecord(context.Background(), "test-token", "123", "rec123", "_acme-challenge", "TXT")
	require.NoError(t, err)
}

func TestListRecords(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/zones/123/records" {
			http.NotFound(w, r)
			return
		}

		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := RecordsResponse{
			Meta: Meta{
				Status:        200,
				StatusMessage: "200 OK",
			},
			Data: []Record{
				{
					ID:    "rec123",
					Name:  "_acme-challenge",
					Type:  "TXT",
					Value: "validation-value",
					TTL:   60,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client := setupTestClient(t, handler)

	records, err := client.ListRecords(context.Background(), "test-token", "123")
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "_acme-challenge", records[0].Name)
	assert.Equal(t, "TXT", records[0].Type)
}
