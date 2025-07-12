package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "xxx"

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, &Token{Token: fakeToken})
}

func mockBuilderIdentity() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.HTTPClient = server.Client()
			client.AuthEndpoint, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithBasicAuth("user", "secret"),
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded())
}

func TestClient_obtainToken(t *testing.T) {
	client := mockBuilderIdentity().
		Route("POST /",
			servermock.ResponseFromFixture("token.json"),
			servermock.CheckForm().Strict().
				With("grant_type", "client_credentials")).
		Build(t)

	assert.Nil(t, client.token)

	tok, err := client.obtainToken(t.Context())
	require.NoError(t, err)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, fakeToken, tok.Token)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilderIdentity().
		Route("POST /",
			servermock.ResponseFromFixture("token.json"),
			servermock.CheckForm().Strict().
				With("grant_type", "client_credentials")).
		Build(t)

	assert.Nil(t, client.token)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	tok := getToken(ctx)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, fakeToken, tok.Token)
}
