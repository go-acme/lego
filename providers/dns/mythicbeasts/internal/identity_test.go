package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"
)

func mockOAuthServer[T any](t *testing.T, fn func(server *httptest.Server) (T, error)) T {
	t.Helper()

	return servermock.NewBuilder[T](fn).
		Route("POST /",
			servermock.ResponseFromFixture("token.json"),
			servermock.CheckHeader().
				WithBasicAuth("user", "secret").
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("grant_type", "client_credentials"),
		).
		Build(t)
}

func buildOAuthClient(server *httptest.Server) (*http.Client, error) {
	config, err := buildOAuthConfig(server)
	if err != nil {
		return nil, err
	}

	return config.Client(context.Background()), nil
}

func buildOAuthConfig(server *httptest.Server) (*clientcredentials.Config, error) {
	config := CreateOAuthConfig(server.URL, "user", "secret")

	return config, nil
}

func mockOAuthClient(t *testing.T) *http.Client {
	t.Helper()

	return mockOAuthServer(t, buildOAuthClient)
}

func Test_Token(t *testing.T) {
	token, err := mockOAuthServer(t, buildOAuthConfig).Token(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "xxx", token.AccessToken)
	assert.Equal(t, "bearer", token.TokenType)
	assert.NotZero(t, token.Expiry)
}
