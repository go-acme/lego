package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("bob", "user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("bob", "user", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders())
}

func TestClient_Publish(t *testing.T) {
	client := mockBuilder().
		Route("PUT /Zone/example.com", servermock.ResponseFromFixture("publish.json"),
			servermock.CheckRequestJSONBody(`{"publish":true,"notes":"my message"}`)).
		Build(t)

	err := client.Publish(t.Context(), "example.com", "my message")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /TXTRecord/example.com/example.com.", servermock.ResponseFromFixture("create-txt-record.json"),
			servermock.CheckRequestJSONBody(`{"rdata":{"txtdata":"txt"},"ttl":"120"}`)).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "example.com.", "txt", 120)
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /TXTRecord/example.com/example.com.", nil).
		Build(t)

	err := client.RemoveTXTRecord(t.Context(), "example.com", "example.com.")
	require.NoError(t, err)
}
