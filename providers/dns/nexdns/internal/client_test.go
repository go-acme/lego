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

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("list_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("search", "example.com").
				With("per_page", "100"),
		).
		Build(t)

	zones, err := client.ListZones(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Zone{
		{ID: "xK9mQ2vR", Name: "example.com", Type: "master", Status: "active"},
		{ID: "pL4nT7wB", Name: "test.example.com", Type: "master", Status: "active"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/xK9mQ2vR/records",
			servermock.ResponseFromFixture("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), "xK9mQ2vR", record)
	require.NoError(t, err)

	// The API returns the TXT content quoted.
	expected := &Record{
		ID:      "Rk3pW9xd",
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`,
		TTL:     120,
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/xK9mQ2vR/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	_, err := client.CreateRecord(t.Context(), "xK9mQ2vR", record)
	require.EqualError(t, err, "validation_error: Validation failed. (content: This value is not valid.)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/xK9mQ2vR/records/Rk3pW9xd",
			servermock.Noop().WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "xK9mQ2vR", "Rk3pW9xd")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/xK9mQ2vR/records/Rk3pW9xd",
			servermock.ResponseFromFixture("error_not_found.json").
				WithStatusCode(http.StatusNotFound),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "xK9mQ2vR", "Rk3pW9xd")
	require.EqualError(t, err, "not_found: Record not found.")
}
