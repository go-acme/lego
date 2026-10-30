package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
)

func mockOAuthClient(t *testing.T) *http.Client {
	t.Helper()

	return servermock.NewBuilder[*http.Client](
		func(server *httptest.Server) (*http.Client, error) {
			config := CreateOAuthConfig(server.URL, "123", "xxx")

			return config.Client(t.Context()), nil
		}).
		Route("POST /",
			servermock.ResponseFromFixture("oauth.json"),
			servermock.CheckForm().Strict().UsePostForm().
				With("client_id", "123").
				With("client_secret", "xxx").
				With("grant_type", "client_credentials").
				With("scope", "dns"),
		).
		Build(t)
}
