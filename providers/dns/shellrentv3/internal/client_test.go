package internal

import (
	"net/http"
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
			client := NewClient(server.Client())
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_ListPurchases(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchases",
			servermock.ResponseFromFixture("purchases.json"),
			servermock.CheckQueryParameter().
				Strict().
				With("search", "example.com").
				With("page", "1").
				With("per_page", "50")).
		Build(t)

	purchases, err := client.ListPurchases(t.Context(), "example.com")
	require.NoError(t, err)

	require.Len(t, purchases, 1)
	assert.Equal(t, 123, purchases[0].PurchaseID)
	require.NotNil(t, purchases[0].DomainID)
	assert.Equal(t, 456, purchases[0].DomainID)
}

func TestClient_ListPurchases_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchases",
			servermock.ResponseFromFixture("error_400.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.ListPurchases(t.Context(), "example.com")
	require.EqualError(t, err, `1000: Invalid request data: domain_prefix: is required; configuration.ram: must be <= 64: {"errors":{"domain_prefix":["is required"],"configuration.ram":["must be <= 64"]}}`)
}

func TestClient_GetDomain(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/456",
			servermock.ResponseFromFixture("domain.json")).
		Build(t)

	domain, err := client.GetDomain(t.Context(), 456)
	require.NoError(t, err)

	require.NotNil(t, domain)
	assert.Equal(t, 456, domain.DomainID)
	assert.Equal(t, "example.com", domain.DomainName.Unicode)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/456/dns/records",
			servermock.ResponseFromFixture("record-create.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("record-create-request.json")).
		Build(t)

	record, err := client.CreateRecord(t.Context(), 456, Record{
		Type:        "TXT",
		Host:        "_acme-challenge",
		Destination: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	})
	require.NoError(t, err)

	assert.Equal(t, "abc789", record.ID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/456/dns/records/abc789",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 456, "abc789")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/456/dns/records/abc789",
			servermock.ResponseFromFixture("error_401.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 456, "abc789")
	require.EqualError(t, err, "1100: The access token is missing.")
}
