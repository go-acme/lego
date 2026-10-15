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
			With(APIKeyHeader, "secret"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/acme-challenge",
			servermock.ResponseFromFixture("create.json"),
			servermock.CheckRequestJSONBodyFromFixture("request.json"),
		).
		Build(t)

	payload := APIRequest{
		Domain: "example.com",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.AddTXTRecord(t.Context(), payload)
	require.NoError(t, err)

	expected := &Record{
		ID:    12345,
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   60,
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/acme-challenge",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity),
		).
		Build(t)

	payload := APIRequest{
		Domain: "example.com",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	_, err := client.AddTXTRecord(t.Context(), payload)
	require.EqualError(t, err, "VALIDATION_ERROR: Validation failed")
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/acme-challenge",
			servermock.ResponseFromFixture("delete.json"),
			servermock.CheckRequestJSONBodyFromFixture("request.json"),
		).
		Build(t)

	payload := APIRequest{
		Domain: "example.com",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.RemoveTXTRecord(t.Context(), payload)
	require.NoError(t, err)
}
