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
			With(AccessKeyHeader, "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/records",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("name", "_acme-challenge.example.com").
				With("ttl", "120").
				With("type", "TXT").
				With("value", `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`),
		).
		Build(t)

	record := Record{
		Name:  "_acme-challenge.example.com",
		Value: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`,
		Type:  "TXT",
		TTL:   120,
	}

	resp, err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &RecordResponse{
		RecordID:     123,
		RecordStatus: "added",
	}

	assert.Equal(t, expected, resp)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	record := Record{
		Name:  "_acme-challenge.example.com",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Type:  "TXT",
		TTL:   120,
	}

	_, err := client.AddRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "fail: Payment required: 3: Authorization failed")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("record_id", "123"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.EqualError(t, err, "fail: Payment required: 3: Authorization failed")
}
