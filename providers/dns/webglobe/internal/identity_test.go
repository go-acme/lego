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

func setupIdentifierClient(server *httptest.Server) (*Identifier, error) {
	client, err := NewIdentifier("user", "secret")
	if err != nil {
		return nil, err
	}

	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, nil
}

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "secret")
}

func TestIdentifier_Authenticate(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("POST /auth/login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Build(t)

	token, err := identifier.Login(t.Context())
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
