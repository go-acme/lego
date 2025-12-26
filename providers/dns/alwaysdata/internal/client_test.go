package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/internal/clientdebug"
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
			client.HTTPClient = clientdebug.Wrap(server.Client())

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("secret", ""),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/",
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
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	client := mockBuilder().
		Route("POST /record/",
			servermock.Noop().WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("record_add-request.json")).
		Build(t)

	record := RecordRequest{
		DomainID:   132,
		Name:       "_acme-challenge",
		Type:       "TXT",
		Value:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:        120,
		Annotation: "lego",
	}

	err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /record/789/",
			servermock.Noop()).
		Build(t)

	err := client.DeleteRecord(t.Context(), 789)
	require.NoError(t, err)
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /record/",
			servermock.ResponseFromFixture("records.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain", "132").
				With("name", "_acme-challenge"),
		).
		Build(t)

	result, err := client.ListRecords(t.Context(), 132, "_acme-challenge")
	require.NoError(t, err)

	expected := []Record{
		{
			ID: 789,
			Domain: &Domain{
				Href: "/v1/domain/132/",
			},
			Type:       "TXT",
			Name:       "_acme-challenge",
			Value:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			TTL:        120,
			Annotation: "lego",
		},
		{
			ID: 11619270,
			Domain: &Domain{
				Href: "/v1/domain/118935/",
			},
			Name:          "home",
			Type:          "A",
			Value:         "149.202.90.65",
			TTL:           300,
			IsUserDefined: true,
			IsActive:      true,
		},
	}

	assert.Equal(t, expected, result)
}
