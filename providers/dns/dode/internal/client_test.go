package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_UpdateTxtRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /letsencrypt", servermock.ResponseFromFixture("success.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com").
				With("token", "secret").
				With("value", "value")).
		Build(t)

	err := client.UpdateTxtRecord(t.Context(), "example.com.", "value", false)
	require.NoError(t, err)
}

func TestClient_UpdateTxtRecord_clear(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /letsencrypt", servermock.ResponseFromFixture("success.json"),
			servermock.CheckQueryParameter().Strict().
				With("action", "delete").
				With("domain", "example.com").
				With("token", "secret")).
		Build(t)

	err := client.UpdateTxtRecord(t.Context(), "example.com.", "value", true)
	require.NoError(t, err)
}
