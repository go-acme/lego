package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_GetTxtRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/example.com/txt", servermock.ResponseFromFixture("get-txt-records.json")).
		Build(t)

	records, err := client.GetTxtRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []TXTRecord{
		{ID: "123", Name: "prefix.example.com", Destination: "server.example.com", Delete: true, Modify: true, ResourceURL: "string"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddTxtRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/example.com/txt",
			servermock.ResponseFromFixture("create-txt-record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"name":"prefix.example.com","destination":"server.example.com"}`)).
		Build(t)

	records, err := client.AddTxtRecord(t.Context(), "example.com", TXTRecord{Name: "prefix.example.com", Destination: "server.example.com"})
	require.NoError(t, err)

	expected := []TXTRecord{
		{ID: "123", Name: "prefix.example.com", Destination: "server.example.com", Delete: true, Modify: true, ResourceURL: "string"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/example.com/txt/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.RemoveTxtRecord(t.Context(), "example.com", "123")
	require.NoError(t, err)
}
