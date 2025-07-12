package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Login(t *testing.T) {
	var serverURL *url.URL

	client := servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret", "example.com", "test")
			client.HTTPClient = server.Client()
			client.IdentityEndpoint = server.URL + "/v3/auth/token"

			serverURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	).
		Route("POST /v3/auth/token", IdentityHandlerMock()).
		Build(t)

	err := client.Login(t.Context())
	require.NoError(t, err)

	assert.Equal(t, serverURL.JoinPath("v2").String(), client.baseURL.String())
	assert.Equal(t, fakeOTCToken, client.token)
}
