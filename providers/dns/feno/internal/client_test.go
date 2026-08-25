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
		Route("GET /domains/example.no", servermock.ResponseFromFixture("get_domain.json")).
		Build(t)

	domain, err := client.GetDomain(t.Context(), "example.no")
	require.NoError(t, err)

	expected := &Domain{Domain: "example.no", Status: "active", FenoNameservers: true}

	assert.Equal(t, expected, domain)
}

func TestClient_GetDomain_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.no",
			servermock.ResponseFromFixture("error_not_managed.json").
				WithStatusCode(http.StatusNotFound).
				WithHeader("X-Request-Id", "0b0a3d3e-6f7c-4a1e-9a1e-2c7c9e0c1a11"),
		).
		Build(t)

	_, err := client.GetDomain(t.Context(), "example.no")
	require.EqualError(t, err, "404: Domain not found or not managed by this account (DOMAIN_NOT_MANAGED) [request 0b0a3d3e-6f7c-4a1e-9a1e-2c7c9e0c1a11]")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)

	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "DOMAIN_NOT_MANAGED", apiErr.Code)
}

func TestClient_GetDomain_error_rateLimited(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.no",
			servermock.ResponseFromFixture("error_rate_limited.json").
				WithStatusCode(http.StatusTooManyRequests).
				WithHeader("Retry-After", "17"),
		).
		Build(t)

	_, err := client.GetDomain(t.Context(), "example.no")
	require.EqualError(t, err, "429: Rate limit exceeded (RATE_LIMITED) retry after 17s")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)

	assert.Equal(t, 17*time.Second, apiErr.RetryAfter)
}

func TestClient_GetDomain_error_unexpected(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.no",
			servermock.RawStringResponse("<html>Bad Gateway</html>").
				WithStatusCode(http.StatusBadGateway),
		).
		Build(t)

	_, err := client.GetDomain(t.Context(), "example.no")
	require.EqualError(t, err, "unexpected status code: [status code: 502] body: <html>Bad Gateway</html>")

	var apiErr *APIError
	require.NotErrorAs(t, err, &apiErr)
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.no/dns", servermock.ResponseFromFixture("list_records.json")).
		Build(t)

	records, err := client.ListRecords(t.Context(), "example.no")
	require.NoError(t, err)

	expected := []Record{
		{ID: 481522, Type: "TXT", Name: "_acme-challenge", Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", TTL: 120},
		{ID: 481523, Type: "TXT", Name: "_acme-challenge", Value: "7SZcH8jldJ5zSS3kgbe2KZDOO-PHTMEqGU37zLnmPyk", TTL: 120},
	}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.no/dns",
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

	result, err := client.CreateRecord(t.Context(), "example.no", record)
	require.NoError(t, err)

	expected := &Record{
		ID:    481522,
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
		Route("DELETE /domains/example.no/dns/481522", servermock.ResponseFromFixture("delete_record.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.no", 481522)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.no/dns/481522",
			servermock.ResponseFromFixture("error_not_managed.json").
				WithStatusCode(http.StatusNotFound),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.no", 481522)
	require.EqualError(t, err, "404: Domain not found or not managed by this account (DOMAIN_NOT_MANAGED)")
}
