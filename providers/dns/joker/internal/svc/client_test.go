package svc

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *clientmock.Builder[*Client] {
	return clientmock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("test", "secret")
			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		clientmock.CheckHeader().
			WithContentTypeFromURLEncoded())
}

func TestClient_Send(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			clientmock.RawStringResponse("OK: 1 inserted, 0 deleted"),
			clientmock.CheckForm().Strict().
				With("zone", "example.com").
				With("label", "_acme-challenge").
				With("type", "TXT").
				With("value", "123").
				With("username", "test").
				With("password", "secret"),
		).
		Build(t)

	zone := "example.com"
	label := "_acme-challenge"
	value := "123"

	err := client.SendRequest(t.Context(), zone, label, value)
	require.NoError(t, err)
}

func TestClient_Send_empty(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			clientmock.RawStringResponse("OK: 1 inserted, 0 deleted"),
			clientmock.CheckForm().Strict().
				With("zone", "example.com").
				With("label", "_acme-challenge").
				With("type", "TXT").
				With("value", "").
				With("username", "test").
				With("password", "secret"),
		).
		Build(t)

	zone := "example.com"
	label := "_acme-challenge"
	value := ""

	err := client.SendRequest(t.Context(), zone, label, value)
	require.NoError(t, err)
}
