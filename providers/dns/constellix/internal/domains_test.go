package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(server.Client())
			client.BaseURL = server.URL

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
}

func TestDomainService_GetAll(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains", servermock.ResponseFromFixture("domains-GetAll.json")).
		Build(t)

	data, err := client.Domains.GetAll(t.Context(), nil)
	require.NoError(t, err)

	expected := []Domain{
		{ID: 273301, Name: "aaa.example", TypeID: 1, Version: 9, Status: "ACTIVE"},
		{ID: 273302, Name: "bbb.example", TypeID: 1, Version: 9, Status: "ACTIVE"},
		{ID: 273303, Name: "ccc.example", TypeID: 1, Version: 9, Status: "ACTIVE"},
		{ID: 273304, Name: "ddd.example", TypeID: 1, Version: 9, Status: "ACTIVE"},
	}

	assert.Equal(t, expected, data)
}

func TestDomainService_Search(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/search",
			servermock.ResponseFromFixture("domains-Search.json"),
			servermock.CheckQueryParameter().Strict().
				With("exact", "example.com")).
		Build(t)

	data, err := client.Domains.Search(t.Context(), Exact, "example.com")
	require.NoError(t, err)

	expected := []Domain{
		{ID: 273302, Name: "example.com", TypeID: 1, Version: 9, Status: "ACTIVE"},
	}

	assert.Equal(t, expected, data)
}
