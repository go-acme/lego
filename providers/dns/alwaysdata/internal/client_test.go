package internal

import (
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
			client, err := NewClient("secret", "")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("secret", ""),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain",
			servermock.ResponseFromFixture("domains.json")).
		Build(t)

	result, err := client.ListDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{
		{ID: 132, Name: "example.com", Annotation: "test"},
		{ID: 133, Name: "example.net", IsInternal: true},
		{ID: 134, Name: "example.org"},
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /record",
			servermock.ResponseFromFixture("record.json"),
			servermock.CheckRequestJSONBodyFromFixture("record_add-request.json")).
		Build(t)

	record := Record{
		DomainID:   132,
		Name:       "_acme-challenge.example.com.",
		Type:       "TXT",
		Value:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:        120,
		Annotation: "lego",
	}

	result, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &Record{
		ID:         789,
		DomainID:   132,
		Name:       "_acme-challenge.example.com.",
		Type:       "TXT",
		Value:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:        120,
		Annotation: "lego",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /record/789",
			servermock.Noop()).
		Build(t)

	err := client.DeleteRecord(t.Context(), 789)
	require.NoError(t, err)
}
