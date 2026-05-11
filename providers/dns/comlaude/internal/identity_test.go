package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilderIdentifier() *servermock.Builder[*Identifier] {
	return servermock.NewBuilder[*Identifier](
		func(server *httptest.Server) (*Identifier, error) {
			client := NewIdentifier()
			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
	)
}

func TestIdentifier_APILogin(t *testing.T) {
	client := mockBuilderIdentifier().
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

	tok, err := client.APILogin(t.Context(), "user", "secret", "key")
	require.NoError(t, err)

	expected := &TokenInfo{
		TokenType:        "Bearer",
		ExpiresIn:        7600,
		AccessToken:      "xxx",
		DNSDashboardLink: 1,
	}

	assert.Equal(t, expected, tok)
}
