package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_LookupParentZoneID(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "user", "secret")
	client.HTTPClient = server.Client()

	mux.HandleFunc("/Services/REST/v1/getEntityByName", func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		if query.Get("name") == "com" {
			_ = json.NewEncoder(rw).Encode(EntityResponse{
				ID:         2,
				Name:       "com",
				Type:       ZoneType,
				Properties: "test",
			})
			return
		}

		http.Error(rw, "{}", http.StatusOK)
	})

	parentID, name, err := client.LookupParentZoneID(t.Context(), 2, "foo.example.com")
	require.NoError(t, err)

	assert.EqualValues(t, 2, parentID)
	assert.Equal(t, "foo.example", name)
}
