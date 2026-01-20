package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifierClient(server *httptest.Server) (*Identifier, error) {
	client := NewIdentifier("user", "secret")
	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, nil
}

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "593959ca04f0de9689b586c6a647d15d")
}

func TestIdentifier_Authentication(t *testing.T) {
	client := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("POST /KasAuth.php",
			servermock.ResponseFromFixture("auth.xml"),
			servermock.CheckRequestBodyFromFixture("auth-request.xml").
				IgnoreWhitespace()).
		Build(t)

	credentialToken, err := client.Authentication(t.Context(), 60, true)
	require.NoError(t, err)

	assert.Equal(t, "593959ca04f0de9689b586c6a647d15d", credentialToken)
}

func TestIdentifier_Authentication_error(t *testing.T) {
	client := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("POST /KasAuth.php", servermock.ResponseFromFixture("auth_fault.xml")).
		Build(t)

	_, err := client.Authentication(t.Context(), 60, false)
	require.Error(t, err)
}
