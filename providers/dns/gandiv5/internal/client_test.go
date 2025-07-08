package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *clientmock.Builder[*Client] {
	return clientmock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret", "xxx")
			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		clientmock.CheckHeader().WithJSONHeaders().
			With("X-Api-Key", "secret").
			WithAuthorization("Bearer xxx"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/records/foo/TXT",
			clientmock.ResponseFromFixture("add_txt_record_get.json")).
		Route("PUT /domains/example.com/records/foo/TXT",
			clientmock.ResponseFromFixture("api_response.json"),
			clientmock.CheckRequestJSONBody(`{"rrset_ttl":120,"rrset_values":["content","value1"]}`)).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "foo", "content", 120)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records/foo/TXT",
			clientmock.ResponseFromFixture("api_response.json")).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "foo")
	require.NoError(t, err)
}
