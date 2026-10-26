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
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_AddChallenge(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/123/acme/challenge",
			servermock.ResponseFromFixture("create_challenge_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_challenge_record-request.json"),
		).
		Build(t)

	results, err := client.AddChallenge(t.Context(), 123, "@", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)

	expected := []ChallengeData{{
		Scope:  "@",
		FQDN:   "_acme-challenge.example.com",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Status: "foo",
	}}

	assert.Equal(t, expected, results)
}

func TestClient_DeleteChallenge(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/123/acme/challenge/@/ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteChallenge(t.Context(), 123, "@", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)
}

func TestClient_DeleteChallenge_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/123/acme/challenge/@/ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteChallenge(t.Context(), 123, "@", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.EqualError(t, err, "authenticationError: Authentication token is not valid or has expired.")
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain",
			servermock.ResponseFromFixture("domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("fqdn", "example.com"),
		).
		Build(t)

	request := &DomainsRequest{
		FQDN: "example.com",
	}

	result, err := client.GetDomains(t.Context(), request)
	require.NoError(t, err)

	expected := []Domain{
		{ID: 1, FQDN: "example.org", IsDNSManaged: true},
		{ID: 2, FQDN: "example.com", IsDNSManaged: true},
		{ID: 3, FQDN: "foo.example.com", IsDNSManaged: true},
	}

	assert.Equal(t, expected, result)
}
