package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user123", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/dns/add-domain-record.json",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("Domain", "example.com").
				With("Host", "_acme-challenge").
				With("Type", "TXT").
				With("Value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("Ttl", "600").
				With("auth-userid", "user123").
				With("api-key", "secret"),
		).
		Build(t)

	record := Record{
		Domain: "example.com",
		Host:   "_acme-challenge",
		Type:   "TXT",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:    "600",
	}

	recordID, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	assert.Equal(t, 11554102, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/dns/add-domain-record.json",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusNotFound),
		).
		Build(t)

	record := Record{
		Domain: "example.com",
		Host:   "_acme-challenge",
		Type:   "TXT",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:    "600",
	}

	_, err := client.AddRecord(t.Context(), record)
	require.EqualError(t, err, "host.repeat (2d5876b2-f272-43e9-acc1-4c6a3d3683b1)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/dns/delete-domain-record.json",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("Id", "123").
				With("auth-userid", "user123").
				With("api-key", "secret"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123)
	require.NoError(t, err)
}
