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

func mockBuilder(customerUUID string) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret", customerUUID)
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("token secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder("").
		Route("POST /domain/example.com/add_dns_record",
			servermock.ResponseFromFixture("add_dns_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("name", "_acme-challenge").
				With("ttl", "3600").
				With("type", "TXT"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		TTL:     3600,
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "abcde12345",
		Name:    "_acme-challenge",
		Type:    "TXT",
		TTL:     3600,
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddRecord_customerUUID(t *testing.T) {
	client := mockBuilder("customer1234").
		Route("POST /domain/example.com/add_dns_record",
			servermock.ResponseFromFixture("add_dns_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("name", "_acme-challenge").
				With("ttl", "3600").
				With("type", "TXT"),
			servermock.CheckHeader().
				With(customerUUIDHeader, "customer1234"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		TTL:     3600,
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "abcde12345",
		Name:    "_acme-challenge",
		Type:    "TXT",
		TTL:     3600,
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder("").
		Route("DELETE /domain/example.com/delete_dns_record",
			servermock.ResponseFromFixture("delete_dns_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("id", "abcde12345"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "abcde12345")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder("").
		Route("DELETE /domain/example.com/delete_dns_record",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "abcde12345")
	require.EqualError(t, err, "400(no_customer): Reason of failure")
}
