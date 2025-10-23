package hetznerhcloud

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSProvider_Present_Success(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"zones": []map[string]any{{"id": "123", "name": "example.com"}},
		})
	})

	mux.HandleFunc("/v1/zones/123/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"record": map[string]any{"id": "456"},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := NewDefaultConfig()
	config.Token = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_Success(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"zones": []map[string]any{{"id": "123", "name": "example.com"}},
		})
	})

	mux.HandleFunc("/v1/zones/123/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			json.NewEncoder(w).Encode(map[string]any{
				"record": map[string]any{"id": "456"},
			})
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.HandleFunc("/v1/zones/123/records/456", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := NewDefaultConfig()
	config.Token = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)

	err = provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_ZoneNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"zones": []map[string]any{}})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := NewDefaultConfig()
	config.Token = "test-token"
	config.BaseURL = server.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
