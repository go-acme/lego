package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/stubrouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, &Token{AccessToken: "xxx"})
}

func setupIdentityClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.AuthEndpoint, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_obtainToken(t *testing.T) {
	client := stubrouter.NewBuilder[*Client](setupIdentityClient,
		stubrouter.CheckHeader().
			WithContentTypeFromURLEncoded(),
	).
		Route("POST /", stubrouter.JSONEncode(Token{
			AccessToken: "xxx",
			TokenID:     "yyy",
			ExpiresIn:   666,
			TokenType:   "Bearer",
			Scope:       "openid profile email roles",
		}),
			stubrouter.CheckForm().Strict().
				With("client_id", "user").
				With("client_secret", "secret").
				With("grant_type", "access_key"),
		).
		Build(t)

	assert.Nil(t, client.token)

	tok, err := client.obtainToken(t.Context())
	require.NoError(t, err)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, "xxx", tok.AccessToken)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := stubrouter.NewBuilder[*Client](setupIdentityClient,
		stubrouter.CheckHeader().
			WithContentTypeFromURLEncoded(),
	).
		Route("POST /", stubrouter.JSONEncode(Token{
			AccessToken: "xxx",
			TokenID:     "yyy",
			ExpiresIn:   666,
			TokenType:   "Bearer",
			Scope:       "openid profile email roles",
		}),
			stubrouter.CheckForm().Strict().
				With("client_id", "user").
				With("client_secret", "secret").
				With("grant_type", "access_key"),
		).
		Build(t)

	assert.Nil(t, client.token)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	tok := getToken(ctx)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, "xxx", tok.AccessToken)
}
