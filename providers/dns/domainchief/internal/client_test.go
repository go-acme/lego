package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret", "t123")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/dns/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
		Comment: "Created by go-acme/lego",
	}

	result, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "r123",
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
		Comment: "Created by go-acme/lego",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns/records/r123",
			servermock.Noop(),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "r123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns/records/r123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "r123")
	require.EqualError(t, err, "This is a message. (domain.tld_unsupported)")
}

func TestClient_CreateACMERecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /acme/challenges",
			servermock.ResponseFromFixture("create_acme_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_acme_record-request.json"),
		).
		Build(t)

	record := Challenge{
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.CreateACMERecord(t.Context(), record)
	require.NoError(t, err)

	expected := &ChallengeRecord{
		ID:        "r123",
		Domain:    "example.com",
		Name:      "_acme-challenge.example.com",
		Value:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		CreatedAt: time.Date(2024, time.August, 25, 15, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteACMERecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /acme/challenges",
			servermock.Noop(),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge.example.com").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	record := Challenge{
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.DeleteACMERecord(t.Context(), record)
	require.NoError(t, err)
}
