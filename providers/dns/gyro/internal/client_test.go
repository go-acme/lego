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

func TestClient_AddACMERecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /cloud/dns/acme-challenge",
			servermock.ResponseFromFixture("add_acme_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_acme_record-request.json"),
		).
		Build(t)

	record := Challenge{
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.AddACMERecord(t.Context(), record)
	require.NoError(t, err)

	expected := &ChallengeRecord{
		ServiceID: 123,
		Zone:      "example.com",
		Name:      "_acme-challenge.example.com",
		Status:    "accepted",
		Values:    1,
		Changed:   true,
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddACMERecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /cloud/dns/acme-challenge",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	record := Challenge{
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	_, err := client.AddACMERecord(t.Context(), record)
	require.EqualError(t, err, "invalid_token: Invalid API key.")
}

func TestClient_RemoveACMERecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /cloud/dns/acme-challenge",
			servermock.ResponseFromFixture("remove_acme_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge.example.com").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	record := Challenge{
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.RemoveACMERecord(t.Context(), record)
	require.NoError(t, err)

	expected := &ChallengeRecord{
		ServiceID: 123,
		Zone:      "example.com",
		Name:      "_acme-challenge.example.com",
		Status:    "accepted",
		Values:    0,
		Changed:   true,
	}

	assert.Equal(t, expected, result)
}
