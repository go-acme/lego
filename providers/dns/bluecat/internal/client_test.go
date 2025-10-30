package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient(server.URL, "user", "secret")
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_LookupParentZoneID(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /Services/REST/v1/getEntityByName",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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

				_, _ = rw.Write([]byte(`{}`))
			})).
		Build(t)

	parentID, name, err := client.LookupParentZoneID(t.Context(), 2, "foo.example.com")
	require.NoError(t, err)

	assert.EqualValues(t, 2, parentID)
	assert.Equal(t, "foo.example", name)
}
