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
			client, err := NewClient("3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("3bffdb32-0c5e-4a8e-9fa7-4afb4b44ad36", "secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:     "TXT",
		Lookup:   "_acme-challenge",
		DomainID: 123,
		Data:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:      120,
	}

	records, err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)

	expected := []Record{{
		ID:       456,
		Name:     "_acme-challenge.example.com",
		Status:   "ACTIVE",
		Type:     "TXT",
		Lookup:   "_acme-challenge",
		StdName:  "yyy",
		Serial:   "xxx",
		DomainID: 123,
		Data:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:      120,
	}}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/records",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	record := Record{
		Type:     "TXT",
		Lookup:   "_acme-challenge",
		DomainID: 123,
		Data:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:      120,
	}

	_, err := client.CreateRecord(t.Context(), record)
	require.EqualError(t, err, "ACCESS_DENIED: Access Denied")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/records/456",
			servermock.ResponseFromFixture("delete_record.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 456)
	require.NoError(t, err)
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/domain",
			servermock.ResponseFromFixture("get_domains.json"),
		).
		Build(t)

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{
		{
			ID:         123,
			Name:       "example.com",
			Status:     "ACTIVE",
			CustomerID: 42,
			STDName:    "yyy",
			Serial:     "xxx",
		},
		{
			ID:         369,
			Name:       "example.org",
			Status:     "ACTIVE",
			CustomerID: 42,
			STDName:    "yyy",
			Serial:     "xxx",
		},
	}

	assert.Equal(t, expected, domains)
}
