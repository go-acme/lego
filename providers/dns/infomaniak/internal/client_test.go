package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			return NewClient(OAuthStaticAccessToken(server.Client(), "secret"), server.URL)
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /2/zones/example.com/records",
			servermock.ResponseFromFixture("record_create.json"),
			servermock.CheckRequestJSONBodyFromFixture("record_create-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("with", "idn"),
		).
		Build(t)

	record := RecordRequest{
		Source: "_acme-challenge",
		Target: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:    300,
		Type:   "TXT",
	}

	result, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:        32824,
		Source:    "_acme-challenge",
		SourceIDN: "_acme-challenge.example.com",
		Type:      "TXT",
		TTL:       300,
		Target:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateRecord_errors(t *testing.T) {
	client := mockBuilder().
		Route("POST /2/zones/example.com/records",
			servermock.ResponseFromFixture("errors.json"),
		).
		Build(t)

	record := RecordRequest{
		Source: "_acme-challenge",
		Target: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:    3600,
		Type:   "TXT",
	}

	_, err := client.CreateRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "error: [validation_failed] Validation failed (attribute_required: The name attribute is required) (attribute_min_value: You must be at least 18 years old)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /2/zones/example.com/records/32824",
			servermock.ResponseFromFixture("record_delete.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 32824)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /2/zones/example.com/records/32824",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 32824)
	require.EqualError(t, err, "error: [object_not_found] Object not found")
}

func TestClient_ZoneExists(t *testing.T) {
	client := mockBuilder().
		Route("GET /2/zones/example.com/exists",
			servermock.ResponseFromFixture("zone_exists.json"),
		).
		Build(t)

	result, err := client.ZoneExists(t.Context(), "example.com")
	require.NoError(t, err)

	assert.True(t, result)
}

func TestClient_ZoneExists_notFound(t *testing.T) {
	client := mockBuilder().
		Route("GET /2/zones/example.com/exists",
			servermock.ResponseFromFixture("zone_exists_not.json").
				WithStatusCode(http.StatusNotFound),
		).
		Build(t)

	result, err := client.ZoneExists(t.Context(), "example.com")
	require.NoError(t, err)

	assert.False(t, result)
}
