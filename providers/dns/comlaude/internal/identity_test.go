package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockTokenSource(ctx context.Context) *servermock.Builder[*TokenSource] {
	return servermock.NewBuilder[*TokenSource](func(server *httptest.Server) (*TokenSource, error) {
		cfg, err := NewAuthConfig("user", "secret", "key")
		if err != nil {
			return nil, err
		}

		cfg.BaseURL, _ = url.Parse(server.URL)
		cfg.HTTPClient = server.Client()

		return cfg.TokenSource(ctx), nil
	})
}

func TestTokenSource_Token(t *testing.T) {
	tokenSource := mockTokenSource(t.Context()).
		Route("POST /api_login",
			servermock.ResponseFromFixture("api_login.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("username", "user").
				With("password", "secret").
				With("api_key", "key"),
		).
		Build(t)

	token, err := tokenSource.Token()
	require.NoError(t, err)

	assert.Equal(t, "xxx", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.EqualValues(t, 7600, token.ExpiresIn)
	assert.WithinDuration(t, time.Now().Add(7600*time.Second), token.Expiry, 2*time.Second)
}

func TestTokenSource_login(t *testing.T) {
	tokenSource := mockTokenSource(t.Context()).
		Route("POST /api_login",
			servermock.ResponseFromFixture("api_login.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("username", "user").
				With("password", "secret").
				With("api_key", "key"),
		).
		Build(t)

	token, err := tokenSource.login(t.Context())
	require.NoError(t, err)

	expected := &TokenInfo{
		TokenType:        "Bearer",
		ExpiresIn:        7600,
		AccessToken:      "xxx",
		DNSDashboardLink: 1,
	}

	assert.Equal(t, expected, token)
}
