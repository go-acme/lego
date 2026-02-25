package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock2.CheckHeader().
			WithJSONHeaders().
			With("API-TOKEN", "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/dns",
			servermock2.ResponseFromFixture("add_record.json"),
			servermock2.CheckQueryParameter().Strict(),
			servermock2.CheckRequestJSONBodyFromFixture("add_record-request.json")).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: strconv.Quote("ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		TTL:     120,
	}

	newRecord, err := client.AddRecord(t.Context(), "example.com", record)
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
			servermock2.ResponseFromFixture("delete_record.json"),
			servermock2.CheckQueryParameter().Strict(),
			servermock2.CheckRequestJSONBodyFromFixture("delete_record-request.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "12345")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns",
			servermock2.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "12345")
	require.EqualError(t, err, "[status code: 401] Something went wrong")
}

func TestClient_DeleteRecord_error_other(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns",
			servermock2.ResponseFromFixture("error_other.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "12345")
	require.EqualError(t, err, "[status code: 404] Resource not found")
}
