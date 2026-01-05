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

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /CloudDNS/Domain/d123/TXTRecord",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
		).
		Build(t)

	record := TXTRecord{
		HostName: "_acme-challenge",
		Text:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:      "120",
	}

	result, err := client.AddTXTRecord(t.Context(), "d123", record)
	require.NoError(t, err)

	expected := &APIResponse{
		ID:         "r123",
		Options:    "string",
		Data:       []Data{{Type: "string", StatusCode: "string"}},
		Message:    "string",
		Status:     "string",
		Type:       "string",
		StatusCode: "string",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /CloudDNS/Domain/d123/TXTRecord/r123",
			servermock.ResponseFromFixture("delete_record.json"),
		).
		Build(t)

	result, err := client.DeleteTXTRecord(t.Context(), "d123", "r123")
	require.NoError(t, err)

	expected := &APIResponse{
		ID:         "r123",
		Options:    "string",
		Data:       []Data{{Type: "string", StatusCode: "string"}},
		Message:    "string",
		Status:     "string",
		Type:       "string",
		StatusCode: "string",
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /CloudDNS/Domain/ALL",
			servermock.ResponseFromFixture("domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("pageno", "1").
				With("perpage", "20").
				With("searchField", "domainName").
				With("searchOper", "cn").
				With("searchString", "example.com").
				With("sortField", "domainName").
				With("sortOrder", "asc"),
		).
		Build(t)

	opts := Options{
		PerPage:         20,
		PageNumber:      1,
		SortField:       FieldDomainName,
		SortOrder:       "asc",
		SearchField:     FieldDomainName,
		SearchString:    "example.com",
		SearchOperation: SearchOperationContains,
	}

	result, err := client.GetDomains(t.Context(), opts)
	require.NoError(t, err)

	expected := []Domain{
		{ID: "d123", DomainName: "example.com"},
	}

	assert.Equal(t, expected, result)
}
