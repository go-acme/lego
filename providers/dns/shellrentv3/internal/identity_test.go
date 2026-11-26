package internal

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"
)

func TestCreateOAuthConfig_Token(t *testing.T) {
	config := servermock.NewBuilder[*clientcredentials.Config](
		func(server *httptest.Server) (*clientcredentials.Config, error) {
			return CreateOAuthConfig(server.URL, "user", "secret"), nil
		}).
		Route("POST /",
			servermock.ResponseFromFixture("token.json"),
			servermock.CheckHeader().
				WithBasicAuth("user", "secret").
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("grant_type", "client_credentials"),
		).
		Build(t)

	token, err := config.Token(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "xxx", token.AccessToken)
	assert.Equal(t, "Bearer", token.Type())
	assert.WithinDuration(t, time.Now().Add(600*time.Second), token.Expiry, 2*time.Second)
}
