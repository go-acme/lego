package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder(apiKey, pat string) *servermock.Builder[*Client] {
	checkHeaders := servermock.CheckHeader().WithJSONHeaders()

	if apiKey != "" {
		checkHeaders = checkHeaders.WithAuthorization("Apikey secret-apikey")
	} else {
		checkHeaders = checkHeaders.WithAuthorization("Bearer secret-pat")
	}

	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey, pat)
			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		checkHeaders,
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder("secret-apikey", "").
		Route("GET /domains/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("get_txt_record.json"),
		).
		Route("PUT /domains/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("api_response.json"),
			servermock.CheckRequestJSONBodyFromFixture("put_txt_record-request.json"),
		).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "_acme-challenge", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", 300)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder("", "secret-pat").
		Route("DELETE /domains/example.com/records/_acme-challenge/TXT",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "_acme-challenge")
	require.NoError(t, err)
}
