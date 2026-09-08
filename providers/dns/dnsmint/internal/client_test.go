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

func TestClient_Add(t *testing.T) {
	client := mockBuilder().
		Route("POST /httpreq/present",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("post-request.json"),
		).
		Build(t)

	data := Data{
		FQDN:  "_acme-challenge.example.com.",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.Add(t.Context(), data)
	require.NoError(t, err)
}

func TestClient_Add_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /httpreq/present",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	data := Data{
		FQDN:  "_acme-challenge.example.com.",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.Add(t.Context(), data)

	assert.EqualError(t, err, "Missing or invalid API key (UNAUTHORIZED)")
}

func TestClient_Remove(t *testing.T) {
	client := mockBuilder().
		Route("POST /httpreq/cleanup",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("post-request.json"),
		).
		Build(t)

	data := Data{
		FQDN:  "_acme-challenge.example.com.",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.Remove(t.Context(), data)
	require.NoError(t, err)
}

func TestClient_Remove_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /httpreq/cleanup",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	data := Data{
		FQDN:  "_acme-challenge.example.com.",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.Remove(t.Context(), data)

	assert.EqualError(t, err, "Missing or invalid API key (UNAUTHORIZED)")
}
