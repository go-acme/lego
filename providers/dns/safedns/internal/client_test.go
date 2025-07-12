package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester/stubrouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *stubrouter.Builder[*Client] {
	return stubrouter.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		stubrouter.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/records",
			stubrouter.ResponseFromFixture("add_record.json").
				WithStatusCode(http.StatusCreated),
			stubrouter.CheckRequestJSONBodyFromFile("add_record-request.json")).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: `"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"`,
		TTL:     dns01.DefaultTTL,
	}

	response, err := client.AddRecord(t.Context(), "example.com", record)
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

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/records",
			stubrouter.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: `"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"`,
		TTL:     dns01.DefaultTTL,
	}

	_, err := client.AddRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "add record: [status code: 401] Unauthenticated")
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/example.com/records/1234567",
			stubrouter.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.RemoveRecord(t.Context(), "example.com", 1234567)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/example.com/records/1234567",
			stubrouter.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.RemoveRecord(t.Context(), "example.com", 1234567)
	require.EqualError(t, err, "remove record: [status code: 401] Unauthenticated")
}
