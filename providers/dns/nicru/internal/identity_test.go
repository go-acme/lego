package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
)

func mockOAuthClient(ctx context.Context) *servermock.Builder[*http.Client] {
	return servermock.NewBuilder[*http.Client](
		func(server *httptest.Server) (*http.Client, error) {
			authConfig := CreateOAuthConfig(server.URL, "client123", "secret123")

			return NewOAuthClient(ctx, authConfig, "user123", "pwd123")
		},
	)
}

func TestNewOAuthClient(t *testing.T) {
	_ = mockOAuthClient(t.Context()).
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
}
