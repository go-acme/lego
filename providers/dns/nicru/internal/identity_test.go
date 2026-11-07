package internal

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func mockTokenSource(ctx context.Context) *servermock.Builder[oauth2.TokenSource] {
	return servermock.NewBuilder[oauth2.TokenSource](
		func(server *httptest.Server) (oauth2.TokenSource, error) {
			authConfig := CreateOAuthConfig(server.URL, "client123", "secret123")

			ts := NewTokenSource(ctx, authConfig, "user123", "pwd123")

			return oauth2.ReuseTokenSource(nil, ts), nil
		},
	)
}

func TestTokenSource_Token(t *testing.T) {
	rt := mockTokenSource(t.Context()).
		Route("POST /",
			// NOTE(ldez): I don't know if the real response is JSON or x-www-form-urlencoded.
			servermock.ResponseFromFixture("auth.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("client_id", "client123").
				With("client_secret", "secret123").
				With("grant_type", "password").
				With("password", "pwd123").
				With("scope", ".+:/dns-master/.+").
				With("username", "user123"),
		).
		Build(t)

	token, err := rt.Token()
	require.NoError(t, err)

	require.Equal(t, "at123", token.AccessToken)

	// caught by ReuseTokenSource: no call to the custom TokenSource.
	token, err = rt.Token()
	require.NoError(t, err)

	require.Equal(t, "at123", token.AccessToken)
}
