package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("clientID", "email@example.com", "secret", 300)
			client.HTTPClient = server.Client()
			client.apiBaseURL, _ = url.Parse(server.URL + "/api")
			client.loginURL, _ = url.Parse(server.URL + "/login")

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/domain/search",
			servermock.ResponseFromFixture("domain_search.json"),
			servermock.CheckRequestJSONBodyFromFile("domain_search-request.json")).
		Route("POST /api/record-txt", nil,
			servermock.CheckRequestJSONBodyFromFile("record_txt-request.json")).
		Route("PUT /api/domain/A/publish", nil,
			servermock.CheckRequestJSONBodyFromFile("publish-request.json")).
		Route("POST /login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFile("login-request.json")).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	err = client.AddRecord(ctx, "example.com", "_acme-challenge.example.com", "txt")
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/domain/search",
			servermock.ResponseFromFixture("domain_search.json"),
			servermock.CheckRequestJSONBodyFromFile("domain_search-request.json")).
		Route("GET /api/domain/A",
			servermock.ResponseFromFixture("domain-request.json")).
		Route("DELETE /api/record/R01", nil).
		Route("PUT /api/domain/A/publish", nil,
			servermock.CheckRequestJSONBodyFromFile("publish-request.json")).
		Route("POST /login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFile("login-request.json")).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	err = client.DeleteRecord(ctx, "example.com", "_acme-challenge.example.com")
	require.NoError(t, err)
}
