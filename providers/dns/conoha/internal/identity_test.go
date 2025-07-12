package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifier(server *httptest.Server) (*Identifier, error) {
	identifier, err := NewIdentifier("tyo1")
	if err != nil {
		return nil, err
	}

	identifier.HTTPClient = server.Client()
	identifier.baseURL, _ = url.Parse(server.URL)

	return identifier, nil
}

func TestNewClient(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifier,
		servermock.CheckHeader().WithJSONHeaders(),
	).
		Route("POST /v2.0/tokens", servermock.ResponseFromFixture("tokens_POST.json")).
		Build(t)

	auth := Auth{
		TenantID: "487727e3921d44e3bfe7ebb337bf085e",
		PasswordCredentials: PasswordCredentials{
			Username: "ConoHa",
			Password: "paSSword123456#$%",
		},
	}

	token, err := identifier.GetToken(t.Context(), auth)
	require.NoError(t, err)

	expected := &IdentityResponse{Access: Access{Token: Token{ID: "sample00d88246078f2bexample788f7"}}}

	assert.Equal(t, expected, token)
}
