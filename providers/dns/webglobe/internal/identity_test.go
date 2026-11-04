package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockTokenSource(ctx context.Context) *servermock.Builder[*TokenSource] {
	return servermock.NewBuilder[*TokenSource](func(server *httptest.Server) (*TokenSource, error) {
		cfg, err := NewAuthConfig("user", "secret")
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
		Route("POST /auth/login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Build(t)

	token, err := tokenSource.Token()
	require.NoError(t, err)

	assert.Equal(t, "secrettoken", token.AccessToken)
	assert.Equal(t, "bearer", token.TokenType)
	assert.EqualValues(t, 7200, token.ExpiresIn)
	assert.NotZero(t, token.Expiry)
}

func TestTokenSource_login(t *testing.T) {
	tokenSource := mockTokenSource(t.Context()).
		Route("POST /auth/login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Build(t)

	token, err := tokenSource.login(t.Context())
	require.NoError(t, err)

	expected := &Token{
		CustomerLoginID:   123456,
		CustomerID:        654321,
		CustomerLoginName: "EXAMPLE001",
		SessionID:         "abcd1234",
		Token:             "secrettoken",
		TokenType:         "bearer",
		ExpiresIn:         7200,
	}

	assert.Equal(t, expected, token)
}
