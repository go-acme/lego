package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("API-TOKEN", "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/dns",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckQueryParameter().Strict(),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json")).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: strconv.Quote("ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		TTL:     120,
	}

	newRecord, err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "12345",
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: strconv.Quote("ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		TTL:     120,
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckQueryParameter().Strict(),
			servermock.CheckRequestJSONBodyFromFixture("delete_record-request.json")).
		Build(t)

	err := client.DeleteRecord(context.Background(), "example.com", "12345")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(context.Background(), "example.com", "12345")
	require.EqualError(t, err, "[status code: 401] Something went wrong")
}

func TestClient_DeleteRecord_error_other(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns",
			servermock.ResponseFromFixture("error_other.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := client.DeleteRecord(context.Background(), "example.com", "12345")
	require.EqualError(t, err, "[status code: 404] Resource not found")
}
