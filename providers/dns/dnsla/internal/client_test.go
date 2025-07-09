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

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/domainList",
			servermock.ResponseFromFixture("domains.json"),
			servermock.CheckQueryParameter().
				With("pageIndex", "1").
				With("pageSize", "10")).
		Build(t)

	pager := Pager{
		PageIndex: 1,
		PageSize:  10,
	}

	domains, err := client.ListDomains(t.Context(), pager)
	require.NoError(t, err)

	expected := []Domain{
		{
			ID:            "85369994254488576",
			CreatedAt:     1692856597,
			UpdatedAt:     1692856597,
			UserID:        "85068081529119744",
			UserAccount:   "example@example.com",
			Domain:        "example.com.",
			DisplayDomain: "example.com.",
			State:         1,
			NsState:       0,
			NsCheckedAt:   0,
			ExpiredAt:     4102416000,
			Suffix:        "com.",
			DisplaySuffix: "com.",
		},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/record",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json")).
		Build(t)

	record := Record{
		DomainID: "85369994254488576",
		Type:     TypeTXT,
		Host:     "_acme-challenge",
		Data:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:      120,
	}

	recordID, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	assert.Equal(t, "85371689655342080", recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/record",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckQueryParameter().
				With("id", "85371689655342080")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "85371689655342080")
	require.NoError(t, err)
}
