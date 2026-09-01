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

func TestClient_GetDomain(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com",
			servermock.ResponseFromFixture("get_domain.json"),
		).
		Build(t)

	domain, err := client.GetDomain(t.Context(), "example.com")
	require.NoError(t, err)

	expected := &Domain{
		Domain: "example.com",
		Status: "active",
		FenoNS: true,
	}

	assert.Equal(t, expected, domain)
}

func TestClient_GetDomain_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com",
			servermock.ResponseFromFixture("error_rate_limited.json").
				WithStatusCode(http.StatusTooManyRequests).
				WithHeader("Retry-After", "42"),
		).
		Build(t)

	_, err := client.GetDomain(t.Context(), "example.com")

	require.EqualError(t, err, "429: Rate limit exceeded (RATE_LIMITED)")
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/dns",
			servermock.ResponseFromFixture("list_records.json"),
		).
		Build(t)

	records, err := client.ListRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{ID: 1, Type: "TXT", Name: "_acme-challenge", Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", TTL: 120},
		{ID: 2, Type: "TXT", Name: "sub", Value: "xxx", TTL: 240},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/dns",
			servermock.ResponseFromFixture("error_rate_limited.json").
				WithStatusCode(http.StatusTooManyRequests).
				WithHeader("Retry-After", "42"),
		).
		Build(t)

	_, err := client.ListRecords(t.Context(), "example.com")

	require.EqualError(t, err, "429: Rate limit exceeded (RATE_LIMITED)")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/dns",
			servermock.ResponseFromFixture("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	result, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:    6,
		Type:  "TXT",
		Name:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.no/dns",
			servermock.ResponseFromFixture("error_insufficient_scope.json").
				WithStatusCode(http.StatusForbidden),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	_, err := client.CreateRecord(t.Context(), "example.no", record)

	require.EqualError(t, err, "403: This key does not hold the required scope (INSUFFICIENT_SCOPE)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns/2",
			servermock.ResponseFromFixture("delete_record.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 2)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/dns/2",
			servermock.ResponseFromFixture("error_rate_limited.json").
				WithStatusCode(http.StatusTooManyRequests).
				WithHeader("Retry-After", "42"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 2)

	require.EqualError(t, err, "429: Rate limit exceeded (RATE_LIMITED)")
}
