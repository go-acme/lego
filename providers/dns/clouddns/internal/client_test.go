package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("clientID", "email@example.com", "secret", 300)
			client.HTTPClient = server.Client()
			client.apiBaseURL, _ = url.Parse(server.URL + "/api")
			client.loginURL, _ = url.Parse(server.URL + "/login")

			return client, nil
		},
		servermock2.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/domain/search",
			servermock2.ResponseFromFixture("domain_search.json"),
			servermock2.CheckRequestJSONBodyFromFixture("domain_search-request.json")).
		Route("POST /api/record-txt", nil,
			servermock2.CheckRequestJSONBodyFromFixture("record_txt-request.json")).
		Route("PUT /api/domain/A/publish", nil,
			servermock2.CheckRequestJSONBodyFromFixture("publish-request.json")).
		Route("POST /login",
			servermock2.ResponseFromFixture("login.json"),
			servermock2.CheckRequestJSONBodyFromFixture("login-request.json")).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	err = client.AddRecord(ctx, "example.com", "_acme-challenge.example.com", "txt")
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/domain/search",
			servermock2.ResponseFromFixture("domain_search.json"),
			servermock2.CheckRequestJSONBodyFromFixture("domain_search-request.json")).
		Route("GET /api/domain/A",
			servermock2.ResponseFromFixture("domain-request.json")).
		Route("DELETE /api/record/R01", nil).
		Route("PUT /api/domain/A/publish", nil,
			servermock2.CheckRequestJSONBodyFromFixture("publish-request.json")).
		Route("POST /login",
			servermock2.ResponseFromFixture("login.json"),
			servermock2.CheckRequestJSONBodyFromFixture("login-request.json")).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	err = client.DeleteRecord(ctx, "example.com", "_acme-challenge.example.com")
	require.NoError(t, err)
}
