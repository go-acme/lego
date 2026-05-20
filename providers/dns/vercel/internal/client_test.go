package internal

import (
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
			client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"), "123")
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v2/domains/example.com/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("teamId", "123"),
		).
		Build(t)

	record := Record{
		Name:  "_acme-challenge.example.com.",
		Type:  "TXT",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   60,
	}

	resp, err := client.CreateRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)

	expected := &CreateRecordResponse{
		UID:     "9e2eab60-0ba5-4dff-b481-2999c9764b84",
		Updated: 1,
	}

	assert.Equal(t, expected, resp)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/domains/example.com/records/1234567",
			servermock.Noop(),
			servermock.CheckQueryParameter().Strict().
				With("teamId", "123"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com.", "1234567")
	require.NoError(t, err)
}
