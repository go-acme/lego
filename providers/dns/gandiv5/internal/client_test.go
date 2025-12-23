package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
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
		Route("GET /domains/example.com/records/foo/TXT",
			servermock.ResponseFromFixture("add_txt_record_get.json")).
		Route("PUT /domains/example.com/records/foo/TXT",
			servermock.ResponseFromFixture("api_response.json"),
			servermock.CheckRequestJSONBody(`{"rrset_ttl":120,"rrset_values":["content","value1"]}`)).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "foo", "content", 120)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder("", "secret-pat").
		Route("DELETE /domains/example.com/records/foo/TXT",
			servermock.ResponseFromFixture("api_response.json")).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "foo")
	require.NoError(t, err)
}
