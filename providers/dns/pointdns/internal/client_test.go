package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:  120,
		Type: "TXT",
	}

	result, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:     141,
		Name:   "_acme-challenge.example.com.",
		Data:   "txtTXTtxt",
		TTL:    120,
		Type:   "TXT",
		ZoneID: 1,
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/example.com/records/123",
			servermock.ResponseFromFixture("delete_record.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.NoError(t, err)
}
